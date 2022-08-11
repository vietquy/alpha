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

	"github.com/vietquy/alpha/users"
	"google.golang.org/grpc"

	"github.com/jmoiron/sqlx"
	"github.com/vietquy/alpha"
	authapi "github.com/vietquy/alpha/authn/api/grpc"
	"github.com/vietquy/alpha/logger"
	"github.com/vietquy/alpha/users/api"
	"github.com/vietquy/alpha/users/bcrypt"
	"github.com/vietquy/alpha/users/postgres"

)

const (
	defLogLevel      = "error"
	defDBHost        = "localhost"
	defDBPort        = "5432"
	defDBUser        = "alpha"
	defDBPass        = "alpha"
	defDB            = "users"
	defHTTPPort      = "8180"
	defAuthnURL     = "localhost:8181"
	defAuthnTimeout = "1" // in seconds

	envLogLevel      = "AP_USERS_LOG_LEVEL"
	envDBHost        = "AP_USERS_DB_HOST"
	envDBPort        = "AP_USERS_DB_PORT"
	envDBUser        = "AP_USERS_DB_USER"
	envDBPass        = "AP_USERS_DB_PASS"
	envDB            = "AP_USERS_DB"
	envHTTPPort      = "AP_USERS_HTTP_PORT"

	envAuthnURL     = "AP_AUTHN_GRPC_URL"
	envAuthnTimeout = "AP_AUTHN_GRPC_TIMEOUT"
)

type config struct {
	logLevel     string
	dbConfig     postgres.Config
	httpPort     string
	authnURL     string
	authnTimeout time.Duration
}

func main() {
	cfg := loadConfig()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	db := connectToDB(cfg.dbConfig, logger)
	defer db.Close()


	auth, close := connectToAuthn(cfg, logger)
	if close != nil {
		defer close()
	}


	svc := newService(db, auth, cfg, logger)
	errs := make(chan error, 2)

	go startHTTPServer(svc, cfg.httpPort, logger, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("Users service terminated: %s", err))
}

func loadConfig() config {
	timeout, err := strconv.ParseInt(alpha.Env(envAuthnTimeout, defAuthnTimeout), 10, 64)
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envAuthnTimeout, err.Error())
	}

	dbConfig := postgres.Config{
		Host:        alpha.Env(envDBHost, defDBHost),
		Port:        alpha.Env(envDBPort, defDBPort),
		User:        alpha.Env(envDBUser, defDBUser),
		Pass:        alpha.Env(envDBPass, defDBPass),
		Name:        alpha.Env(envDB, defDB),
	}

	return config{
		logLevel:     alpha.Env(envLogLevel, defLogLevel),
		dbConfig:     dbConfig,
		httpPort:     alpha.Env(envHTTPPort, defHTTPPort),
		authnURL:     alpha.Env(envAuthnURL, defAuthnURL),
		authnTimeout: time.Duration(timeout) * time.Second,
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

func connectToAuthn(cfg config, logger logger.Logger) (alpha.AuthNServiceClient, func() error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	logger.Info("gRPC communication is not encrypted")

	conn, err := grpc.Dial(cfg.authnURL, opts...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to authn service: %s", err))
		os.Exit(1)
	}

	return authapi.NewClient(conn, cfg.authnTimeout), conn.Close
}

func newService(db *sqlx.DB, auth alpha.AuthNServiceClient, c config, logger logger.Logger) users.Service {
	repo := postgres.New(db)
	hasher := bcrypt.New()

	svc := users.New(repo, hasher, auth)
	svc = api.LoggingMiddleware(svc, logger)

	return svc
}

func startHTTPServer(svc users.Service, port string, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", port)
	logger.Info(fmt.Sprintf("Users service started using http, exposed port %s", port))
	errs <- http.ListenAndServe(p, api.MakeHandler(svc, logger))
}
