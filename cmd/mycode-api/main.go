package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/twitchtv/twirp"

	"github.com/dimuls/mycode"
	"github.com/dimuls/mycode/pg"
	"github.com/dimuls/mycode/rmq"
)

func main() {
	os.Exit(run())
}

func run() int {
	var (
		postgresURI            string
		jwtSecret              string
		listenAddress          string
		rmqURI                 string
		runHandlingParallelism int
	)

	flag.StringVar(&postgresURI, "postgres-uri", "", "postgres URI")
	flag.StringVar(&jwtSecret, "jwt-secret", "", "JWT token secret")
	flag.StringVar(&listenAddress, "listen-address", "127.0.0.1:3000", "listen address")
	flag.StringVar(&rmqURI, "rmq-uri", "amqp://guest:guest@localhost:5672/", "rmq URI")
	flag.IntVar(&runHandlingParallelism, "run-handling-parallelism", 30, "run handling parallelism")
	flag.Parse()

	switch "" {
	case postgresURI, jwtSecret, listenAddress, rmqURI:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if runHandlingParallelism <= 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	sig := make(chan os.Signal)

	var stopTime time.Time
	defer func() {
		if stopTime.IsZero() {
			return
		}
		logrus.WithField("duration", time.Now().Sub(stopTime)).
			Info("stopped")
	}()

	rmqCodePublisher, err := rmq.NewCodePublisher(rmqURI)
	if err != nil {
		logrus.WithError(err).Error("failed to create rmq_code_publisher")
		return 1
	}
	defer func() {
		err := rmqCodePublisher.Close()
		if err != nil {
			logrus.WithError(err).Error("failed to close rmq_code_publisher")
		}

		logrus.Info("rmq_code_publisher closed")
	}()

	logrus.Info("rmq_code_publisher created")

	pgMyCodeAPI, err := pg.NewMyCodeAPI(postgresURI, jwtSecret, rmqCodePublisher)
	if err != nil {
		logrus.WithError(err).Error("failed to create pg_mycode_api")
		return 2
	}

	defer func() {
		err := pgMyCodeAPI.Close()
		if err != nil {
			logrus.WithError(err).Error("failed to close pg_mycode_api")
			return
		}

		logrus.Info("pg_mycode_api closed")
	}()

	logrus.Info("pg_mycode_api created and started")

	rmqRunConsumer, err := rmq.NewRunConsumer(rmqURI, runHandlingParallelism,
		pgMyCodeAPI)
	if err != nil {
		logrus.WithError(err).Error("failed to create rmq_run_consumer")
		return 3
	}
	defer func() {
		err := rmqRunConsumer.Close()
		if err != nil {
			logrus.WithError(err).Error("failed to close rmq_run_consumer")
		}

		logrus.Info("rmq_run_consumer closed")
	}()

	logrus.Info("rmq_run_consumer created and started")

	err = pgMyCodeAPI.Migrate()
	if err != nil {
		logrus.WithError(err).Error("failed to migrate")
		return 4
	}

	hooks := &twirp.ServerHooks{}
	hooks.RequestRouted = func(ctx context.Context) (context.Context, error) {

		method, ok := twirp.MethodName(ctx)
		if !ok {
			return ctx, twirp.NewError(twirp.Internal,
				"missing method name")
		}

		ctx, err := pgMyCodeAPI.Authorize(ctx, method)
		if err != nil {
			return ctx, twirp.NewError(twirp.PermissionDenied,
				err.Error())
		}

		return ctx, nil
	}

	apiSrv := cors.AllowAll().Handler(
		pg.WithJWT(mycode.NewAPIServer(pgMyCodeAPI,
			twirp.WithServerPathPrefix(""),
			twirp.WithServerHooks(hooks))))

	s := &http.Server{
		Addr:    listenAddress,
		Handler: apiSrv,
	}

	go func() {
		err := s.ListenAndServe()
		if err != nil {
			if err == http.ErrServerClosed {
				return
			}
			logrus.WithError(err).Error("failed to start http_server")
			sig <- syscall.SIGTERM
		}
	}()

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := s.Shutdown(ctx)
		if err != nil {
			logrus.WithError(err).Error("failed to stop http_server")
			return
		}

		logrus.Info("http_server stopped")
	}()

	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	logrus.Info("started")

	logrus.Infof("catch %s signal, stopping", <-sig)

	stopTime = time.Now()

	return 0
}
