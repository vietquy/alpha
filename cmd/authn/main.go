package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jmoiron/sqlx"
	"github.com/vietquy/alpha"
	"github.com/vietquy/alpha/authn"
	api "github.com/vietquy/alpha/authn/api"
	grpcapi "github.com/vietquy/alpha/authn/api/grpc"
	httpapi "github.com/vietquy/alpha/authn/api/http"
	"github.com/vietquy/alpha/authn/jwt"
	"github.com/vietquy/alpha/authn/postgres"
	mfidp "github.com/vietquy/alpha/authn/uuid"
	"github.com/vietquy/alpha/logger"
	"google.golang.org/grpc"
)

const (
	defLogLevel      = "error"
	defDBHost        = "localhost"
	defDBPort        = "5432"
	defDBUser        = "alpha"
	defDBPass        = "alpha"
	defDB            = "authn"
	defHTTPPort      = "8180"
	defGRPCPort      = "8181"
	defSecret        = "authn"

	envLogLevel      = "AP_AUTHN_LOG_LEVEL"
	envDBHost        = "AP_AUTHN_DB_HOST"
	envDBPort        = "AP_AUTHN_DB_PORT"
	envDBUser        = "AP_AUTHN_DB_USER"
	envDBPass        = "AP_AUTHN_DB_PASS"
	envDB            = "AP_AUTHN_DB"
	envHTTPPort      = "AP_AUTHN_HTTP_PORT"
	envGRPCPort      = "AP_AUTHN_GRPC_PORT"
	envSecret        = "AP_AUTHN_SECRET"
)

type config struct {
	logLevel   string
	dbConfig   postgres.Config
	httpPort   string
	grpcPort   string
	secret     string
}

type tokenConfig struct {
	hmacSampleSecret []byte // secret for signing token
	tokenDuration    string // token in duration in min
}

func main() {
	cfg := loadConfig()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	db := connectToDB(cfg.dbConfig, logger)
	defer db.Close()

	svc := newService(db, cfg.secret, logger)
	errs := make(chan error, 2)

	go startHTTPServer(svc, cfg.httpPort, logger, errs)
	go startGRPCServer(svc, cfg.grpcPort, logger, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("Authentication service terminated: %s", err))
}

func loadConfig() config {
	dbConfig := postgres.Config{
		Host:        alpha.Env(envDBHost, defDBHost),
		Port:        alpha.Env(envDBPort, defDBPort),
		User:        alpha.Env(envDBUser, defDBUser),
		Pass:        alpha.Env(envDBPass, defDBPass),
		Name:        alpha.Env(envDB, defDB),
	}

	return config{
		logLevel:   alpha.Env(envLogLevel, defLogLevel),
		dbConfig:   dbConfig,
		httpPort:   alpha.Env(envHTTPPort, defHTTPPort),
		grpcPort:   alpha.Env(envGRPCPort, defGRPCPort),
		secret:     alpha.Env(envSecret, defSecret),
	}

}


func connectToDB(dbConfig postgres.Config, logger logger.Logger) *sqlx.DB {
	db, err := postgres.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to postgres: %s", err))
		os.Exit(1)
	}
	return db
}

func newService(db *sqlx.DB, secret string, logger logger.Logger) authn.Service {
	repo := postgres.New(db)

	idp := mfidp.New()
	t := jwt.New(secret)
	svc := authn.New(repo, idp, t)
	svc = api.LoggingMiddleware(svc, logger)

	return svc
}

func startHTTPServer(svc authn.Service, port string, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", port)
	logger.Info(fmt.Sprintf("Authentication service started using http, exposed port %s", port))
	errs <- http.ListenAndServe(p, httpapi.MakeHandler(svc))

}

func startGRPCServer(svc authn.Service, port string, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", port)
	listener, err := net.Listen("tcp", p)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to listen on port %s: %s", port, err))
	}

	var server *grpc.Server
	
	logger.Info(fmt.Sprintf("Authentication gRPC service started using http on port %s", port))
	server = grpc.NewServer()

	alpha.RegisterAuthNServiceServer(server, grpcapi.NewServer(svc))
	logger.Info(fmt.Sprintf("Authentication gRPC service started, exposed port %s", port))
	errs <- server.Serve(listener)
}
