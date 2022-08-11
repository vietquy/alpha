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

	influxdata "github.com/influxdata/influxdb/client/v2"
	"github.com/vietquy/alpha"
	"github.com/vietquy/alpha/logger"
	"github.com/vietquy/alpha/reader"
	"github.com/vietquy/alpha/reader/api"
	"github.com/vietquy/alpha/reader/influxdb"
	thingsapi "github.com/vietquy/alpha/things/api/grpc"
	"google.golang.org/grpc"
)

const (
	defLogLevel          = "error"
	defPort              = "8180"
	defDB                = "messages"
	defDBHost            = "localhost"
	defDBPort            = "8086"
	defDBUser            = "alpha"
	defDBPass            = "alpha"
	defThingsAuthURL     = "localhost:8181"
	defThingsAuthTimeout = "1" // in seconds

	envLogLevel          = "AP_READER_LOG_LEVEL"
	envPort              = "AP_READER_PORT"
	envDB                = "AP_READER_DB"
	envDBHost            = "AP_READER_DB_HOST"
	envDBPort            = "AP_READER_DB_PORT"
	envDBUser            = "AP_READER_DB_USER"
	envDBPass            = "AP_READER_DB_PASS"
	envThingsAuthURL     = "AP_THINGS_AUTH_GRPC_URL"
	envThingsAuthTimeout = "AP_THINGS_AUTH_GRPC_TIMEOUT"
)

type config struct {
	logLevel          string
	port              string
	dbName            string
	dbHost            string
	dbPort            string
	dbUser            string
	dbPass            string
	thingsAuthURL     string
	thingsAuthTimeout time.Duration
}

func main() {
	cfg, clientCfg := loadConfigs()
	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}
	conn := connectToThings(cfg, logger)
	defer conn.Close()

	tc := thingsapi.NewClient(conn, cfg.thingsAuthTimeout)

	client, err := influxdata.NewHTTPClient(clientCfg)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create InfluxDB client: %s", err))
		os.Exit(1)
	}
	defer client.Close()

	repo := newService(client, cfg.dbName, logger)

	errs := make(chan error, 2)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go startHTTPServer(repo, tc, cfg, logger, errs)

	err = <-errs
	logger.Error(fmt.Sprintf("InfluxDB writer service terminated: %s", err))
}

func loadConfigs() (config, influxdata.HTTPConfig) {
	timeout, err := strconv.ParseInt(alpha.Env(envThingsAuthTimeout, defThingsAuthTimeout), 10, 64)
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envThingsAuthTimeout, err.Error())
	}

	cfg := config{
		logLevel:          alpha.Env(envLogLevel, defLogLevel),
		port:              alpha.Env(envPort, defPort),
		dbName:            alpha.Env(envDB, defDB),
		dbHost:            alpha.Env(envDBHost, defDBHost),
		dbPort:            alpha.Env(envDBPort, defDBPort),
		dbUser:            alpha.Env(envDBUser, defDBUser),
		dbPass:            alpha.Env(envDBPass, defDBPass),
		thingsAuthURL:     alpha.Env(envThingsAuthURL, defThingsAuthURL),
		thingsAuthTimeout: time.Duration(timeout) * time.Second,
	}

	clientCfg := influxdata.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%s", cfg.dbHost, cfg.dbPort),
		Username: cfg.dbUser,
		Password: cfg.dbPass,
	}

	return cfg, clientCfg
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

func newService(client influxdata.Client, dbName string, logger logger.Logger) reader.MessageRepository {
	repo := influxdb.New(client, dbName)
	repo = api.LoggingMiddleware(repo, logger)

	return repo
}

func startHTTPServer(repo reader.MessageRepository, tc alpha.ThingsServiceClient, cfg config, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", cfg.port)
	logger.Info(fmt.Sprintf("InfluxDB reader service started, exposed port %s", cfg.port))
	errs <- http.ListenAndServe(p, api.MakeHandler(repo, tc, "influxdb-reader"))
}
