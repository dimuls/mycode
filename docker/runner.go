package docker

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/sirupsen/logrus"

	"github.com/dimuls/mycode"
)

type RunPublisher interface {
	PublishRun(*mycode.Run) error
}

type Runner struct {
	docker       *client.Client
	runPublisher RunPublisher

	log *logrus.Entry
	wg  sync.WaitGroup
}

func NewRunner(dockerHost string, rp RunPublisher) (
	r *Runner, err error) {

	docker, err := client.NewClientWithOpts(client.WithHost(dockerHost),
		client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	return &Runner{
		docker:       docker,
		runPublisher: rp,
		log:          logrus.WithField("subsystem", "nats_runner"),
	}, nil
}

func (r *Runner) Close() error {
	err := r.docker.Close()
	if err != nil {
		return fmt.Errorf("close docker client: %w", err)
	}
	return nil
}

const (
	myCodeRun = "mycode-run"

	containerRunTimeout = 60 * time.Second
)

var images = map[mycode.Language]string{
	mycode.Language_c:      "mycode-c",
	mycode.Language_cpp:    "mycode-cpp",
	mycode.Language_go:     "mycode-go",
	mycode.Language_java:   "mycode-java",
	mycode.Language_pascal: "mycode-pascal",
	mycode.Language_python: "mycode-python",
}

func (r *Runner) run(ctx context.Context, log *logrus.Entry,
	c *mycode.Code) (run *mycode.Run, err error) {

	defer func() {
		if err != nil {
			log.WithError(err).Error("failed to process code")
		} else {
			log.Info("code processed")
		}
	}()

	image, exists := images[c.Language]
	if !exists {
		return nil, fmt.Errorf("unsupported language")
	}

	srcPath, err := r.createSrcFile(c.Language, c.Source)
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}

	defer func() {
		err := r.removeSrcFile(srcPath)
		if err != nil {
			log.WithError(err).WithField("src_path", srcPath).
				Error("failed to remove src file")
		}
	}()

	flag.Parse()

	createResp, err := r.docker.ContainerCreate(ctx,
		&container.Config{
			Image: image,
			Cmd: []string{
				myCodeRun,
				c.Language.String(),
				srcPath,
				c.Stdin,
			},
			NetworkDisabled: true,
		}, &container.HostConfig{
			Privileged: true,
			Mounts: []mount.Mount{
				{
					Type:   mount.TypeBind,
					Source: srcPath,
					Target: srcPath,
				},
			},
		}, nil, "")
	if err != nil {
		return nil, fmt.Errorf("create container: %w", err)
	}

	defer func() {
		err := r.docker.ContainerRemove(ctx, createResp.ID,
			types.ContainerRemoveOptions{
				Force: true,
			})
		if err != nil {
			log.WithError(err).WithField("container_id", createResp.ID).
				Error("failed to remove container")
		}
	}()

	err = r.docker.ContainerStart(ctx, createResp.ID,
		types.ContainerStartOptions{})
	if err != nil {
		return nil, fmt.Errorf("start container: %w", err)
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx,
		containerRunTimeout)
	defer cancel()

	waitRespChan, errChan := r.docker.ContainerWait(ctxWithTimeout,
		createResp.ID, container.WaitConditionNotRunning)

	var waitResp container.ContainerWaitOKBody

	select {
	case err = <-errChan:
		return nil, fmt.Errorf("wait container: %w", err)
	case waitResp = <-waitRespChan:
	}

	if waitResp.Error != nil {
		return nil, fmt.Errorf("wait response error: %s",
			waitResp.Error.Message)
	}

	out, err := r.docker.ContainerLogs(ctx, createResp.ID,
		types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return nil, fmt.Errorf("log container: %w", err)
	}

	defer func() {
		err := out.Close()
		if err != nil {
			logrus.WithError(err).Error("failed to close container logs out")
		}
	}()

	var stdOut, stdErr bytes.Buffer

	_, err = stdcopy.StdCopy(&stdOut, &stdErr, out)
	if err != nil {
		return nil, fmt.Errorf("std copy out: %w", err)
	}

	if stdErr.Len() != 0 {
		log.WithField("stderr", stdErr.String()).
			Warn("not empty stderr")
	}

	if waitResp.StatusCode != 0 {
		return nil, fmt.Errorf("not zero container exit code")
	}

	run = &mycode.Run{}

	err = jsonpb.Unmarshal(&stdOut, run)
	if err != nil {
		fmt.Println(run)
		return nil, fmt.Errorf("JSON unmarshal run: %w", err)
	}

	return run, nil
}

func (r *Runner) HandleCode(ctx context.Context, c *mycode.Code) (err error) {

	log := r.log.WithFields(logrus.Fields{
		"solution_test_id": c.SolutionTestId,
	})

	log.Info("code recieved")

	solutionLog := log.WithField("code_type", "solution")

	solutionRun, err := r.run(ctx, solutionLog, c)
	if err != nil {
		solutionLog.WithError(err).Error("failed to run solution code")
		return fmt.Errorf("run solution code: %w", err)
	}

	solutionRun.SolutionTestId = c.SolutionTestId

	if c.WithChecker {

		checkerLog := log.WithField("code_type", "checker")

		checkerRun, err := r.run(ctx, checkerLog, &mycode.Code{
			Language: c.CheckerLanguage,
			Source:   c.CheckerSource,
			Stdin:    solutionRun.Stdout,
		})
		if err != nil {
			checkerLog.WithError(err).Error("failed to run checker code")
			return fmt.Errorf("run checker code: %w", err)
		}

		solutionRun.CheckerStdout = checkerRun.Stdout
		solutionRun.CheckerStderr = checkerRun.Stderr
	}

	err = r.runPublisher.PublishRun(solutionRun)
	if err != nil {
		return fmt.Errorf("publish run: %w", err)
	}

	return nil
}

func (r *Runner) createSrcFile(lang mycode.Language, content string) (
	string, error) {

	f, err := ioutil.TempFile("", fmt.Sprintf("%s-*", lang))
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}

	_, err = f.WriteString(content)
	if err != nil {
		return f.Name(), fmt.Errorf("write source file: %w", err)
	}

	err = f.Close()
	if err != nil {
		return "", fmt.Errorf("close source file: %w", err)
	}

	return f.Name(), err
}

func (r *Runner) removeSrcFile(srcPath string) error {
	return os.Remove(srcPath)
}
