package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/leoluyi/go-api-template/docs"
	"github.com/leoluyi/go-api-template/internal/album"
	"github.com/leoluyi/go-api-template/internal/auth"
	"github.com/leoluyi/go-api-template/internal/config"
	"github.com/leoluyi/go-api-template/internal/errors"
	"github.com/leoluyi/go-api-template/internal/healthcheck"
	"github.com/leoluyi/go-api-template/pkg/accesslog"
	"github.com/leoluyi/go-api-template/pkg/dbcontext"
	"github.com/leoluyi/go-api-template/pkg/log"
)

// Version indicates the current version of the application.
var Version = "1.0.0"

var flagConfig = flag.String("config", "./config/local.yml", "path to the config file")

// @title Swagger Example API
// @version 1.0
// @description This is a sample server Petstore server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host petstore.swagger.io
// @BasePath /v1
func main() {
	flag.Parse()
	// create root logger tagged with server version
	logger := log.New().With(context.TODO(), "version", Version)

	// load application configurations
	cfg, err := config.Load(*flagConfig, logger)
	if err != nil {
		logger.Errorf("failed to load application configuration: %s", err)
		os.Exit(-1)
	}

	// connect to the database
	db, err := sqlx.Open("postgres", cfg.DSN)
	if err != nil {
		logger.Error(err)
		os.Exit(-1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error(err)
		}
	}()

	// build HTTP server
	address := fmt.Sprintf(":%v", cfg.ServerPort)
	hs := &http.Server{
		Addr:    address,
		Handler: buildHandler(logger, dbcontext.New(db), cfg),
	}

	// start the HTTP server with graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := hs.Shutdown(ctx); err != nil {
			logger.Error(err)
		}
	}()

	logger.Infof("server %v is running at %v", Version, address)
	if err := hs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error(err)
		os.Exit(-1)
	}
}

// buildHandler sets up the HTTP routing and builds an HTTP handler.
func buildHandler(logger log.Logger, db *dbcontext.DB, cfg *config.Config) http.Handler {
	router := chi.NewRouter()

	router.Use(
		accesslog.Handler(logger),
		errors.Handler(logger),
		middleware.AllowContentType("application/json"),
		cors.AllowAll().Handler,
	)

	healthcheck.RegisterHandlers(router, Version)

	router.Route("/v1", func(r chi.Router) {
		r.Get("/swagger*", httpSwagger.Handler())

		authHandler := auth.Handler(cfg.JWTSigningKey)
		auth.RegisterHandlers(r, auth.NewService(cfg.JWTSigningKey, cfg.JWTExpiration, logger), logger)
		album.RegisterHandlers(r, album.NewService(album.NewRepository(db, logger), logger), authHandler, logger)
	})

	return router
}
