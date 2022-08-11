package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	influxdata "github.com/influxdata/influxdb/client/v2"
	"github.com/vietquy/alpha"
	"github.com/vietquy/alpha/logger"
	"github.com/vietquy/alpha/messaging/nats"
	"github.com/vietquy/alpha/transformer"
	"github.com/vietquy/alpha/writer"
	"github.com/vietquy/alpha/writer/api"
	"github.com/vietquy/alpha/writer/influxdb"
)

const (
	svcName = "influxdb-writer"

	defNatsURL         = "nats://localhost:4222"
	defLogLevel        = "error"
	defPort            = "8180"
	defDB              = "messages"
	defDBHost          = "localhost"
	defDBPort          = "8086"
	defDBUser          = "alpha"
	defDBPass          = "alpha"

	envNatsURL         = "AP_NATS_URL"
	envLogLevel        = "AP_WRITER_LOG_LEVEL"
	envPort            = "AP_WRITER_PORT"
	envDB              = "AP_WRITER_DB"
	envDBHost          = "AP_WRITER_DB_HOST"
	envDBPort          = "AP_WRITER_DB_PORT"
	envDBUser          = "AP_WRITER_DB_USER"
	envDBPass          = "AP_WRITER_DB_PASS"
)

type config struct {
	natsURL         string
	logLevel        string
	port            string
	dbName          string
	dbHost          string
	dbPort          string
	dbUser          string
	dbPass          string
	contentType     string
}

func main() {
	cfg, clientCfg := loadConfigs()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	pubSub, err := nats.NewPubSub(cfg.natsURL, "", logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer pubSub.Close()

	client, err := influxdata.NewHTTPClient(clientCfg)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create InfluxDB client: %s", err))
		os.Exit(1)
	}
	defer client.Close()

	repo := influxdb.New(client, cfg.dbName)

	repo = api.LoggingMiddleware(repo, logger)
	t := transformer.New()

	if err := writer.Start(pubSub, repo, t, logger); err != nil {
		logger.Error(fmt.Sprintf("Failed to start InfluxDB writer: %s", err))
		os.Exit(1)
	}

	errs := make(chan error, 2)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go startHTTPService(cfg.port, logger, errs)

	err = <-errs
	logger.Error(fmt.Sprintf("InfluxDB writer service terminated: %s", err))
}

func loadConfigs() (config, influxdata.HTTPConfig) {
	cfg := config{
		natsURL:         alpha.Env(envNatsURL, defNatsURL),
		logLevel:        alpha.Env(envLogLevel, defLogLevel),
		port:            alpha.Env(envPort, defPort),
		dbName:          alpha.Env(envDB, defDB),
		dbHost:          alpha.Env(envDBHost, defDBHost),
		dbPort:          alpha.Env(envDBPort, defDBPort),
		dbUser:          alpha.Env(envDBUser, defDBUser),
		dbPass:          alpha.Env(envDBPass, defDBPass),
	}

	clientCfg := influxdata.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%s", cfg.dbHost, cfg.dbPort),
		Username: cfg.dbUser,
		Password: cfg.dbPass,
	}

	return cfg, clientCfg
}

func startHTTPService(port string, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", port)
	logger.Info(fmt.Sprintf("InfluxDB writer service started, exposed port %s", p))
	errs <- http.ListenAndServe(p, api.MakeHandler(svcName))
}
