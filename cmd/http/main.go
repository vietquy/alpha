package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/vietquy/alpha"
	adapter "github.com/vietquy/alpha/http"
	"github.com/vietquy/alpha/http/api"
	"github.com/vietquy/alpha/logger"
	"github.com/vietquy/alpha/messaging/nats"
	thingsapi "github.com/vietquy/alpha/things/api/grpc"
	"google.golang.org/grpc"
)

const (
	defLogLevel          = "error"
	defPort              = "8180"
	defNatsURL           = "nats://localhost:4222"
	defThingsAuthURL     = "localhost:8181"
	defThingsAuthTimeout = "1" // in seconds

	envLogLevel          = "AP_HTTP_ADAPTER_LOG_LEVEL"
	envPort              = "AP_HTTP_ADAPTER_PORT"
	envNatsURL           = "AP_NATS_URL"
	envThingsAuthURL     = "AP_THINGS_AUTH_GRPC_URL"
	envThingsAuthTimeout = "AP_THINGS_AUTH_GRPC_TIMEOUT"
)

type config struct {
	natsURL           string
	logLevel          string
	port              string
	thingsAuthURL     string
	thingsAuthTimeout time.Duration
}

func main() {
	cfg := loadConfig()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	conn := connectToThings(cfg, logger)
	defer conn.Close()


	pub, err := nats.NewPublisher(cfg.natsURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer pub.Close()

	tc := thingsapi.NewClient(conn, cfg.thingsAuthTimeout)
	svc := adapter.New(pub, tc)

	svc = api.LoggingMiddleware(svc, logger)

	errs := make(chan error, 2)

	go func() {
		p := fmt.Sprintf(":%s", cfg.port)
		logger.Info(fmt.Sprintf("HTTP adapter service started on port %s", cfg.port))
		errs <- http.ListenAndServe(p, api.MakeHandler(svc))
	}()

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("HTTP adapter terminated: %s", err))
}

func loadConfig() config {
	timeout, err := strconv.ParseInt(alpha.Env(envThingsAuthTimeout, defThingsAuthTimeout), 10, 64)
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envThingsAuthTimeout, err.Error())
	}

	return config{
		natsURL:           alpha.Env(envNatsURL, defNatsURL),
		logLevel:          alpha.Env(envLogLevel, defLogLevel),
		port:              alpha.Env(envPort, defPort),
		thingsAuthURL:     alpha.Env(envThingsAuthURL, defThingsAuthURL),
		thingsAuthTimeout: time.Duration(timeout) * time.Second,
	}
}


func connectToThings(cfg config, logger logger.Logger) *grpc.ClientConn {
	var opts []grpc.DialOption

	logger.Info("gRPC communication is not encrypted")
	opts = append(opts, grpc.WithInsecure())
	
	conn, err := grpc.Dial(cfg.thingsAuthURL, opts...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to things service: %s", err))
		os.Exit(1)
	}
	return conn
}
