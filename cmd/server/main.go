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
	"github.com/golang-migrate/migrate/v4"
	migratepg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/leoluyi/go-api-template/docs"
	"github.com/leoluyi/go-api-template/internal/album"
	"github.com/leoluyi/go-api-template/internal/auth"
	"github.com/leoluyi/go-api-template/internal/config"
	"github.com/leoluyi/go-api-template/internal/errors"
	"github.com/leoluyi/go-api-template/internal/healthcheck"
	"github.com/leoluyi/go-api-template/migrations"
	"github.com/leoluyi/go-api-template/pkg/accesslog"
	"github.com/leoluyi/go-api-template/pkg/dbcontext"
	"github.com/leoluyi/go-api-template/pkg/log"
	"github.com/leoluyi/go-api-template/pkg/metrics"
)

// Version is set at build time via -ldflags "-X main.Version=$(git describe --tags)".
// Falls back to "dev" when built without the Makefile (e.g. go run ./cmd/server).
var Version = "dev"

var flagConfig = flag.String("config", "./config/local.yml", "path to the config file")

// @title           Go API Template
// @version         1.0
// @description     RESTful API starter kit built with Go, chi, and PostgreSQL.

// @contact.name    Issues
// @contact.url     https://github.com/leoluyi/go-api-template/issues

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:8080
// @BasePath        /v1

// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
func main() {
	flag.Parse()
	docs.SwaggerInfo.Version = Version

	// create root logger tagged with server version
	logger := log.New().With(context.TODO(), "version", Version)

	// load application configurations
	cfg, err := config.Load(*flagConfig)
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
	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.DBConnMaxLifetimeMinutes) * time.Minute)
	defer func() {
		if err := db.Close(); err != nil {
			logger.Error(err)
		}
	}()

	// run database migrations
	if err := runMigrations(db, logger); err != nil {
		logger.Errorf("failed to run database migrations: %s", err)
		os.Exit(-1)
	}

	// build HTTP server
	address := fmt.Sprintf(":%v", cfg.ServerPort)
	hs := &http.Server{
		Addr:         address,
		Handler:      buildHandler(logger, db, dbcontext.New(db), cfg, docs.SwaggerInfo.BasePath),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// start the HTTP server with graceful shutdown
	shutdownTimeout := time.Duration(cfg.ShutdownTimeoutSeconds) * time.Second
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
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

// runMigrations applies all pending database migrations embedded in the binary.
// It returns nil if there are no pending migrations (migrate.ErrNoChange).
func runMigrations(db *sqlx.DB, logger log.Logger) error {
	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}
	driver, err := migratepg.WithInstance(db.DB, &migratepg.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("init migrator: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("apply migrations: %w", err)
	}
	logger.Info("database migrations applied")
	return nil
}

// buildHandler sets up the HTTP routing and builds an HTTP handler.
// basePath is the API route prefix (e.g. "/v1"), sourced from docs.SwaggerInfo.BasePath.
func buildHandler(logger log.Logger, rawDB *sqlx.DB, db *dbcontext.DB, cfg *config.Config, basePath string) http.Handler {
	router := chi.NewRouter()

	router.Use(
		accesslog.Handler(logger),
		errors.Handler(logger),
		metrics.Middleware,
		middleware.RequestSize(1<<20), // 1 MB max request body
		middleware.AllowContentType("application/json"),
		cors.New(cors.Options{
			AllowedOrigins: cfg.CORSAllowedOrigins,
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"Authorization", "Content-Type"},
			MaxAge:         300,
		}).Handler,
	)

	healthcheck.RegisterHandlers(router, Version, rawDB)
	router.Method(http.MethodGet, "/metrics", metrics.Handler())

	router.Route(basePath, func(r chi.Router) {
		r.Get("/swagger/*", httpSwagger.Handler())

		authHandler := auth.Handler(cfg.JWTSigningKey)
		auth.RegisterHandlers(r, auth.NewService(cfg.JWTSigningKey, cfg.JWTExpiration, cfg.AuthUsername, cfg.AuthPassword, logger), logger)
		album.RegisterHandlers(r, album.NewService(album.NewRepository(db, logger), logger), authHandler, logger)
	})

	return router
}
