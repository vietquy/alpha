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
	mflog "github.com/vietquy/alpha/logger"
	"github.com/vietquy/alpha/messaging"
	mqttpub "github.com/vietquy/alpha/messaging/mqtt"
	"github.com/vietquy/alpha/messaging/nats"
	"github.com/vietquy/alpha/mqtt"
	thingsapi "github.com/vietquy/alpha/things/api/grpc"
	mp "github.com/vietquy/alpha/mqtt/proxy/mqtt"
	"github.com/vietquy/alpha/mqtt/proxy/session"
	ws "github.com/vietquy/alpha/mqtt/proxy/websocket"
	"google.golang.org/grpc"
)

const (
	// Logging
	defLogLevel = "error"
	envLogLevel = "AP_MQTT_ADAPTER_LOG_LEVEL"
	// MQTT
	defMQTTHost             = "0.0.0.0"
	defMQTTPort             = "1883"
	defMQTTTargetHost       = "0.0.0.0"
	defMQTTTargetPort       = "1883"
	defMQTTForwarderTimeout = "30" // in seconds

	envMQTTHost             = "AP_MQTT_ADAPTER_MQTT_HOST"
	envMQTTPort             = "AP_MQTT_ADAPTER_MQTT_PORT"
	envMQTTTargetHost       = "AP_MQTT_ADAPTER_MQTT_TARGET_HOST"
	envMQTTTargetPort       = "AP_MQTT_ADAPTER_MQTT_TARGET_PORT"
	envMQTTForwarderTimeout = "AP_MQTT_ADAPTER_FORWARDER_TIMEOUT"
	// HTTP
	defHTTPHost       = "0.0.0.0"
	defHTTPPort       = "8080"
	defHTTPScheme     = "ws"
	defHTTPTargetHost = "localhost"
	defHTTPTargetPort = "8080"
	defHTTPTargetPath = "/mqtt"
	envHTTPHost       = "AP_MQTT_ADAPTER_WS_HOST"
	envHTTPPort       = "AP_MQTT_ADAPTER_WS_PORT"
	envHTTPScheme     = "AP_MQTT_ADAPTER_WS_SCHEMA"
	envHTTPTargetHost = "AP_MQTT_ADAPTER_WS_TARGET_HOST"
	envHTTPTargetPort = "AP_MQTT_ADAPTER_WS_TARGET_PORT"
	envHTTPTargetPath = "AP_MQTT_ADAPTER_WS_TARGET_PATH"
	// Things
	defThingsAuthURL     = "localhost:8181"
	defThingsAuthTimeout = "1" // in seconds
	envThingsAuthURL     = "AP_THINGS_AUTH_GRPC_URL"
	envThingsAuthTimeout = "AP_THINGS_AUTH_GRPC_TIMMEOUT"
	// Nats
	defNatsURL = "nats://localhost:4222"
	envNatsURL = "AP_NATS_URL"
)

type config struct {
	mqttHost             string
	mqttPort             string
	mqttTargetHost       string
	mqttTargetPort       string
	mqttForwarderTimeout time.Duration
	httpHost             string
	httpPort             string
	httpScheme           string
	httpTargetHost       string
	httpTargetPort       string
	httpTargetPath       string
	logLevel             string
	thingsURL            string
	thingsAuthURL        string
	thingsAuthTimeout    time.Duration
	natsURL              string
}

func main() {
	cfg := loadConfig()

	logger, err := mflog.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	conn := connectToThings(cfg, logger)
	defer conn.Close()

	cc := thingsapi.NewClient(conn, cfg.thingsAuthTimeout)

	nps, err := nats.NewPubSub(cfg.natsURL, "mqtt", logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer nps.Close()
	mp, err := mqttpub.NewPublisher(fmt.Sprintf("%s:%s", cfg.mqttTargetHost, cfg.mqttTargetPort), cfg.mqttForwarderTimeout)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create MQTT publisher: %s", err))
		os.Exit(1)
	}
	fwd := mqtt.NewForwarder(nats.SubjectAllProjects, logger)
	if err := fwd.Forward(nps, mp); err != nil {
		logger.Error(fmt.Sprintf("Failed to forward NATS messages: %s", err))
		os.Exit(1)
	}

	np, err := nats.NewPublisher(cfg.natsURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer np.Close()

	// Event handler for MQTT hooks
	h := mqtt.NewHandler([]messaging.Publisher{np}, cc, logger)

	errs := make(chan error, 2)

	logger.Info(fmt.Sprintf("Starting MQTT proxy on port %s", cfg.mqttPort))
	go proxyMQTT(cfg, logger, h, errs)

	logger.Info(fmt.Sprintf("Starting MQTT over WS  proxy on port %s", cfg.httpPort))
	go proxyWS(cfg, logger, h, errs)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("mProxy terminated: %s", err))
}

func loadConfig() config {
	authTimeout, err := strconv.ParseInt(alpha.Env(envThingsAuthTimeout, defThingsAuthTimeout), 10, 64)
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envThingsAuthTimeout, err.Error())
	}

	mqttTimeout, err := strconv.ParseInt(alpha.Env(envMQTTForwarderTimeout, defMQTTForwarderTimeout), 10, 64)
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envThingsAuthTimeout, err.Error())
	}

	return config{
		mqttHost:             alpha.Env(envMQTTHost, defMQTTHost),
		mqttPort:             alpha.Env(envMQTTPort, defMQTTPort),
		mqttTargetHost:       alpha.Env(envMQTTTargetHost, defMQTTTargetHost),
		mqttTargetPort:       alpha.Env(envMQTTTargetPort, defMQTTTargetPort),
		mqttForwarderTimeout: time.Duration(mqttTimeout) * time.Second,
		httpHost:             alpha.Env(envHTTPHost, defHTTPHost),
		httpPort:             alpha.Env(envHTTPPort, defHTTPPort),
		httpScheme:           alpha.Env(envHTTPScheme, defHTTPScheme),
		httpTargetHost:       alpha.Env(envHTTPTargetHost, defHTTPTargetHost),
		httpTargetPort:       alpha.Env(envHTTPTargetPort, defHTTPTargetPort),
		httpTargetPath:       alpha.Env(envHTTPTargetPath, defHTTPTargetPath),
		thingsAuthURL:        alpha.Env(envThingsAuthURL, defThingsAuthURL),
		thingsAuthTimeout:    time.Duration(authTimeout) * time.Second,
		thingsURL:            alpha.Env(envThingsAuthURL, defThingsAuthURL),
		natsURL:              alpha.Env(envNatsURL, defNatsURL),
		logLevel:             alpha.Env(envLogLevel, defLogLevel),
	}
}


func connectToThings(cfg config, logger mflog.Logger) *grpc.ClientConn {
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

func proxyMQTT(cfg config, logger mflog.Logger, handler session.Handler, errs chan error) {
	address := fmt.Sprintf("%s:%s", cfg.mqttHost, cfg.mqttPort)
	target := fmt.Sprintf("%s:%s", cfg.mqttTargetHost, cfg.mqttTargetPort)
	mp := mp.New(address, target, handler, logger)

	errs <- mp.Proxy()
}
func proxyWS(cfg config, logger mflog.Logger, handler session.Handler, errs chan error) {
	target := fmt.Sprintf("%s:%s", cfg.httpTargetHost, cfg.httpTargetPort)
	wp := ws.New(target, cfg.httpTargetPath, cfg.httpScheme, handler, logger)
	http.Handle("/mqtt", wp.Handler())

	p := fmt.Sprintf(":%s", cfg.httpPort)
	errs <- http.ListenAndServe(p, nil)
}
