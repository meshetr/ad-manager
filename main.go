package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	viper.AutomaticEnv()
	var (
		httpAddr = ":8080"
		dsn      = "host=" + viper.GetString("DB_HOST") +
			" user=" + viper.GetString("DB_USER") +
			" password=" + viper.GetString("DB_PASS") +
			" dbname=" + viper.GetString("DB_NAME") +
			" port=" + viper.GetString("DB_PORT") +
			" sslmode=" + viper.GetString("DB_SSL") +
			" TimeZone=" + viper.GetString("DB_TIMEZONE")
		db, _ = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		opts  []grpc.DialOption
	)

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(viper.GetString("GCP_CLIENT_SECRET"))))
	if err != nil {
		level.Error(logger).Log("component", "storage.NewClient", "msg", err)
	} else {
		defer storageClient.Close()
	}

	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial(viper.GetString("IMAGE_PROCESSOR_URL"), opts...)
	if err != nil {
		level.Error(logger).Log("component", "grpc.DIal", "msg", err)
	} else {
		defer conn.Close()
	}

	var service Service
	{
		service = MakeService(logger, db, storageClient, conn)
	}

	var httpHandler http.Handler
	{
		httpHandler = MakeHTTPHandler(logger, service)
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		level.Info(logger).Log("component", "HTTPServer", "msg", "Server started successfully!", "context", "port"+httpAddr)
		errs <- http.ListenAndServe(httpAddr, httpHandler)
	}()

	level.Error(logger).Log("status", "exit", "msg", <-errs)
}
