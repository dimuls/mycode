package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/c2h5oh/datasize"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/sirupsen/logrus"

	"github.com/dimuls/mycode"
	"github.com/dimuls/mycode/docker"
)

const (
	memoryLimitBytes = 107_374_182_400 // 100MB
)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	if len(os.Args) < 3 {
		logrus.WithField("args", os.Args).
			Fatal("expected 3 or more  args")
	}

	languageID, exists := mycode.Language_value[os.Args[1]]
	if !exists {
		logrus.WithField("language", os.Args[1]).
			Fatal("invalid language")
	}

	language := mycode.Language(languageID)

	srcBytes, err := ioutil.ReadFile(os.Args[2])
	if err != nil {
		logrus.WithField("src_path", os.Args[2]).
			Fatal("failed to read source")
	}

	src := string(srcBytes)

	stdin := os.Args[3]

	r, err := docker.Prepare(language, src)
	if err != nil {
		logrus.WithError(err).Fatal("failed to prepare run")
	}

	srcPath, err := r.SourcePath()
	if err != nil {
		logrus.WithError(err).Fatal("failed to get source path")
	}

	additionalBinds, err := r.AdditionalBinds()
	if err != nil {
		logrus.WithError(err).Fatal("failed to get additional binds")
	}

	compileCmd, err := r.CompileCommand()
	if err != nil {
		logrus.WithError(err).Fatal("failed to get compile command")
	}

	if len(compileCmd) != 0 {
		cmd := exec.Command(compileCmd[0], compileCmd[1:]...)
		cmd.Dir = srcPath

		var stdOutBuf, stdErrBuf bytes.Buffer

		cmd.Stdout = &stdOutBuf
		cmd.Stderr = &stdErrBuf

		err := cmd.Run()
		if err != nil {
			err = printRun(&mycode.Run{
				Duration:   "0",
				UsedMemory: "0",
				Stdout:     stdOutBuf.String(),
				Stderr:     stdErrBuf.String(),
			})
			if err != nil {
				logrus.WithError(err).Fatal("failed to print run")
			}
		}
	}

	runCmd, err := r.RunCommand()
	if err != nil {
		logrus.WithError(err).Fatal("failed to get run command")
	}

	if len(runCmd) == 0 {
		logrus.WithField("language", language).
			Fatal("run preparator returned empty run command")
	}

	args := []string{
		"--max_cpus", "1",
		"--rlimit_as", strconv.Itoa(memoryLimitBytes),
		"--user", "99999", "--group", "99999",
		"--bindmount_ro", "/lib",
		"--bindmount_ro", "/usr/lib",
		"--bindmount_ro", "/usr/bin/time",
	}

	for _, ab := range additionalBinds {
		args = append(args, "--bindmount_ro", ab)
	}

	args = append(args,
		"--bindmount", srcPath,
		"--cwd", srcPath,
		"--",
		"/usr/bin/time", "-q", "-f", "%M", "--")

	args = append(args, runCmd...)

	cmd := exec.Command("/usr/bin/nsjail", args...)

	var stdOutBuf, stdErrBuf bytes.Buffer

	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = &stdOutBuf
	cmd.Stderr = &stdErrBuf

	start := time.Now()
	err = cmd.Run()
	duration := time.Since(start)
	if err != nil {
		if !okStdErr(stdErrBuf) {
			logrus.WithError(err).WithFields(logrus.Fields{
				"stdout": stdOutBuf.String(),
				"stderr": stdErrBuf.String(),
			}).Fatal("failed to run")
		}
	}

	stdErr, memoryUsageKB, err := parseStdErr(stdErrBuf)
	if err != nil {
		logrus.WithError(err).Fatal("failed to parse stderr")
	}

	err = printRun(&mycode.Run{
		Duration:   duration.String(),
		UsedMemory: (datasize.ByteSize(memoryUsageKB) * datasize.KB).String(),
		Stdout:     stdOutBuf.String(),
		Stderr:     stdErr,
	})
	if err != nil {
		logrus.WithError(err).Fatal("failed to print run")
	}
}

var exitWithStatusRe = regexp.MustCompile(`exited with status`)

func okStdErr(stdErr bytes.Buffer) bool {
	lines := strings.Split(strings.TrimSpace(stdErr.String()), "\n")
	if len(lines) == 0 {
		return false
	}
	return exitWithStatusRe.MatchString(lines[len(lines)-1])
}

var executingRe = regexp.MustCompile(`Executing '/usr/bin/time' for`)

func parseStdErr(stdErr bytes.Buffer) (string, int, error) {

	lines := strings.Split(strings.TrimSpace(stdErr.String()), "\n")

	if len(lines) < 3 {
		return "", 0, fmt.Errorf("expected at least 3 lines")
	}

	lines = lines[:len(lines)-1]

	for i, l := range lines {
		if executingRe.MatchString(l) {
			lines = lines[i+1:]
			break
		}
	}

	memoryUsageStr := lines[len(lines)-1]
	lines = lines[:len(lines)-1]

	memoryUsage, err := strconv.Atoi(memoryUsageStr)
	if err != nil {
		return "", 0, fmt.Errorf("parse memory usage: %w", err)
	}

	return strings.Join(lines, "\n"), memoryUsage, nil
}

func printRun(run *mycode.Run) error {

	err := (&jsonpb.Marshaler{}).Marshal(os.Stdout, run)
	if err != nil {
		return fmt.Errorf("JSON marshal run to stdout: %w", err)
	}

	return nil
}
