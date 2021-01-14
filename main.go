package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/go-kit/kit/log"
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
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	ctx := context.Background()
	storageClient, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(viper.GetString("GCP_CLIENT_SECRET"))))
	if err != nil {
		logger.Log("storage.NewClient: %v", err)
	} else {
		defer storageClient.Close()
	}

	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())
	conn, err := grpc.Dial("image-processor-service:50051", opts...)
	if err != nil {
		logger.Log("gRPC: fail to dial: %v", err)
	} else {
		defer conn.Close()
	}

	var service Service
	{
		service = MakeService(db, storageClient, conn)
		//service = LoggingMiddleware(logger)(service)
	}

	var httpHandler http.Handler
	{
		httpHandler = MakeHTTPHandler(service, log.With(logger, "component", "HTTP"))
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "HTTP", "addr", httpAddr)
		errs <- http.ListenAndServe(httpAddr, httpHandler)
	}()

	logger.Log("exit", <-errs)
}
