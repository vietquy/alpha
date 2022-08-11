package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/vietquy/alpha"
	authapi "github.com/vietquy/alpha/authn/api/grpc"
	"github.com/vietquy/alpha/logger"
	"github.com/vietquy/alpha/things"
	"github.com/vietquy/alpha/things/api"
	authgrpcapi "github.com/vietquy/alpha/things/api/grpc"
	thhttpapi "github.com/vietquy/alpha/things/api/http"
	"github.com/vietquy/alpha/things/postgres"
	"github.com/vietquy/alpha/things/uuid"
	"google.golang.org/grpc"
)

const (
	defLogLevel        = "error"
	defDBHost          = "localhost"
	defDBPort          = "5432"
	defDBUser          = "alpha"
	defDBPass          = "alpha"
	defDB              = "things"
	defHTTPPort        = "8182"
	defGRPCPort    = "8181"
	defAuthnURL        = "localhost:8181"
	defAuthnTimeout    = "1" // in seconds

	envLogLevel        = "AP_THINGS_LOG_LEVEL"
	envDBHost          = "AP_THINGS_DB_HOST"
	envDBPort          = "AP_THINGS_DB_PORT"
	envDBUser          = "AP_THINGS_DB_USER"
	envDBPass          = "AP_THINGS_DB_PASS"
	envDB              = "AP_THINGS_DB"
	envHTTPPort        = "AP_THINGS_HTTP_PORT"
	envGRPCPort    	   = "AP_THINGS_GRPC_PORT"
	envAuthnURL        = "AP_AUTHN_GRPC_URL"
	envAuthnTimeout    = "AP_AUTHN_GRPC_TIMEOUT"
)

type config struct {
	logLevel        string
	dbConfig        postgres.Config
	httpPort        string
	grpcPort    string
	authnURL        string
	authnTimeout    time.Duration
}

func main() {
	cfg := loadConfig()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	db := connectToDB(cfg.dbConfig, logger)
	defer db.Close()

	auth, close := createAuthClient(cfg, logger)
	if close != nil {
		defer close()
	}

	svc := newService(auth, db, logger)
	errs := make(chan error, 2)

	go startHTTPServer(thhttpapi.MakeHandler(svc), cfg.httpPort, cfg, logger, errs)
	go startGRPCServer(svc, cfg, logger, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("Things service terminated: %s", err))
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
		logLevel:        alpha.Env(envLogLevel, defLogLevel),
		dbConfig:        dbConfig,
		httpPort:        alpha.Env(envHTTPPort, defHTTPPort),
		grpcPort:    	 alpha.Env(envGRPCPort, defGRPCPort),
		authnURL:        alpha.Env(envAuthnURL, defAuthnURL),
		authnTimeout:    time.Duration(timeout) * time.Second,
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

func createAuthClient(cfg config, logger logger.Logger) (alpha.AuthNServiceClient, func() error) {
	conn := connectToAuth(cfg, logger)
	return authapi.NewClient(conn, cfg.authnTimeout), conn.Close
}

func connectToAuth(cfg config, logger logger.Logger) *grpc.ClientConn {
	var opts []grpc.DialOption

	opts = append(opts, grpc.WithInsecure())
	logger.Info("gRPC communication is not encrypted")

	conn, err := grpc.Dial(cfg.authnURL, opts...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to authn service: %s", err))
		os.Exit(1)
	}

	return conn
}

func newService(auth alpha.AuthNServiceClient, db *sqlx.DB, logger logger.Logger) things.Service {
	thingsRepo := postgres.NewThingRepository(db)
	projectsRepo := postgres.NewProjectRepository(db)

	idp := uuid.New()

	svc := things.New(auth, thingsRepo, projectsRepo, idp)
	svc = api.LoggingMiddleware(svc, logger)

	return svc
}

func startHTTPServer(handler http.Handler, port string, cfg config, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", port)
	logger.Info(fmt.Sprintf("Things service started using http on port %s", cfg.httpPort))
	errs <- http.ListenAndServe(p, handler)
}

func startGRPCServer(svc things.Service, cfg config, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", cfg.grpcPort)
	listener, err := net.Listen("tcp", p)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to listen on port %s: %s", cfg.grpcPort, err))
		os.Exit(1)
	}

	var server *grpc.Server
	logger.Info(fmt.Sprintf("Things gRPC service started using http on port %s", cfg.grpcPort))
	server = grpc.NewServer()
	

	alpha.RegisterThingsServiceServer(server, authgrpcapi.NewServer(svc))
	errs <- server.Serve(listener)
}
