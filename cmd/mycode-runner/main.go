package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"

	"github.com/dimuls/mycode/docker"
	"github.com/dimuls/mycode/rmq"
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		dockerHost              string
		rmqURI                  string
		codeHandlingParallelism int
	)

	flag.StringVar(&dockerHost, "docker-host", "unix:///var/run/docker.sock", "docker host")
	flag.StringVar(&rmqURI, "rmq-uri", "amqp://guest:guest@localhost:5672/", "rabbitMQ URI")
	flag.IntVar(&codeHandlingParallelism, "code-handling-parallelism", 30, "code handling parallelism")
	flag.Parse()

	switch "" {
	case rmqURI:
		flag.PrintDefaults()
		return 1
	}

	if codeHandlingParallelism <= 0 {
		flag.PrintDefaults()
		return 2
	}

	var stopTime time.Time
	defer func() {
		if stopTime.IsZero() {
			return
		}
		logrus.WithField("duration", time.Now().Sub(stopTime)).
			Info("stopped")
	}()

	rmqRunPublisher, err := rmq.NewRunPublisher(rmqURI)
	if err != nil {
		logrus.WithError(err).Error("failed to create rmq_run_publisher")
		return 3
	}
	defer func() {
		err := rmqRunPublisher.Close()
		if err != nil {
			logrus.WithError(err).Error("failed to close rmq_run_publisher")
			return
		}
		logrus.Info("rmq_run_publisher closed")
	}()

	logrus.Info("rmq_run_publisher created")

	dockerRunner, err := docker.NewRunner(dockerHost, rmqRunPublisher)
	if err != nil {
		logrus.WithError(err).Error("failed to create docker_runner")
		return 4
	}

	logrus.Info("docker_runner created")

	rmqCodeConsumer, err := rmq.NewCodeConsumer(rmqURI, codeHandlingParallelism,
		dockerRunner)
	if err != nil {
		logrus.WithError(err).Error("failed to create rmq_code_consumer")
		return 5
	}
	defer func() {
		err := rmqCodeConsumer.Close()
		if err != nil {
			logrus.WithError(err).Error("failed to close rmq_code_consumer")
			return
		}
		logrus.Info("rmq_code_consumer closed")
	}()

	logrus.Info("rmq_code_consumer created and started")

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	logrus.Info("started")

	logrus.Infof("catch %s signal, stopping", <-sig)

	stopTime = time.Now()

	return 0
}

func logrusMessageProducer(ctx context.Context, format string,
	level logrus.Level, code codes.Code, err error, fields logrus.Fields) {
	if err != nil {
		fields[logrus.ErrorKey] = err
	}
	entry := ctxlogrus.Extract(ctx).WithContext(ctx).WithFields(fields)

	delete(entry.Data, "system")

	switch level {
	case logrus.DebugLevel:
		entry.Debugf(format)
	case logrus.InfoLevel:
		entry.Infof(format)
	case logrus.WarnLevel:
		entry.Warningf(format)
	case logrus.ErrorLevel:
		entry.Errorf(format)
	case logrus.FatalLevel:
		entry.Fatalf(format)
	case logrus.PanicLevel:
		entry.Panicf(format)
	}
}
