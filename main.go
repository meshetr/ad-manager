package main

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	var (
		httpAddr = ":8080"
		dsn      = "host=localhost user=dbuser password=verisikret dbname=meshetr port=5432 sslmode=disable TimeZone=Europe/Belgrade"
		db, _    = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	)

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var service Service
	{
		service = MakeService(db)
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
