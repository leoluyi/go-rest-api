# Go RESTful API Starter Kit

This starter kit is designed to get you up and running with a project structure optimized for developing
RESTful API services in Go. It promotes the best practices that follow the [SOLID principles](https://en.wikipedia.org/wiki/SOLID)
and [clean architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html).
It encourages writing clean and idiomatic Go code.

The kit provides the following features right out of the box:

- RESTful endpoints in the widely accepted format
- Standard CRUD operations of a database table
- JWT-based authentication
- Environment dependent application configuration management
- Structured logging with contextual information
- Error handling with proper error response generation (includes `request_id` in every error response)
- Database migration (embedded in binary, applied automatically on startup)
- Data validation
- Swagger API documentation
- Full test coverage
- Live reloading during development
- Health check endpoint with database connectivity probe (returns 503 when DB is unreachable)
- Prometheus metrics (`/metrics`) with per-route request count and latency histograms
- Configurable CORS (per-environment allowed origins)

## Best Practices

### Log version on startup

The server version is injected into the root logger context at startup, so every log line carries it automatically.
On `ListenAndServe`, the version and bound address are logged explicitly:

```go
logger := log.New().With(ctx, "version", Version)
// ...
logger.Infof("server %v is running at %v", Version, address)
```

`Version` defaults to `"dev"` in source. When built via `make build` or run via `make run`, the Makefile injects the real version from the latest git tag at compile time:

```makefile
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-X main.Version=${VERSION}"
```

Tag a release with `git tag v2.8.0` and `make build` will embed `v2.8.0` in the binary automatically.

### Graceful exit

The server listens for `SIGINT` and `SIGTERM` and gives in-flight requests up to 10 seconds to complete before shutting down:

```go
signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
<-quit
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
hs.Shutdown(ctx)
```

### Staged Dockerfile

`cmd/server/Dockerfile` uses a two-stage build. The build stage uses `registry.suse.com/bci/golang:1.24`; the runtime stage uses `registry.suse.com/bci/bci-micro` — a SUSE BCI distroless-style image with no shell and no package manager. The binary and config files are copied in with `--chown=65532:65532` and the container runs as UID/GID `65532:65532` (non-root).

### Embedded database migrations

Migration SQL files are embedded directly into the binary at compile time using `//go:embed`:

```go
// migrations/migrations.go
//go:embed *.sql
var FS embed.FS
```

On startup, `main.go` runs all pending migrations via the `golang-migrate/migrate/v4` library before the HTTP server begins accepting requests. `migrate.ErrNoChange` (no pending migrations) is treated as success. This eliminates the need for a separate `migrate` CLI binary, a shell script entrypoint, or a shell in the runtime image.

### Custom middlewares

Four custom middlewares are wired into the router in `cmd/server/main.go`:

| Middleware | Location | Purpose |
|------------|----------|---------|
| Access log | `pkg/accesslog` | Logs method, path, status, duration, and bytes; sets `X-Request-ID` response header |
| Error handler | `internal/errors` | Recovers from panics and maps errors to structured JSON responses with `request_id` |
| Metrics | `pkg/metrics` | Records `http_requests_total` and `http_request_duration_seconds` for Prometheus |
| JWT auth | `internal/auth` | Verifies and authenticates Bearer tokens on protected routes |

### Swagger API documentation

Each handler is annotated individually with `@Summary`, `@Param`, `@Success`, `@Failure`, and `@Security` tags. The security scheme is declared once in `main.go`:

```go
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
```

Protected endpoints reference it with `// @Security BearerAuth`. The Swagger UI is served at `/v1/swagger/` using the `/*` wildcard pattern required by chi:

```go
r.Get("/swagger/*", httpSwagger.Handler())
```

Regenerate the spec after changing annotations:

```shell
make generate-docs
```

The kit uses the following Go packages:

| Concern | Package |
|---------|---------|
| Router | [go-chi/chi/v5](https://github.com/go-chi/chi) |
| JWT middleware | [go-chi/jwtauth/v5](https://github.com/go-chi/jwtauth) |
| JWT signing | [golang-jwt/jwt/v4](https://github.com/golang-jwt/jwt) |
| CORS | [go-chi/cors](https://github.com/go-chi/cors) |
| Database access | [jmoiron/sqlx](https://github.com/jmoiron/sqlx) + [lib/pq](https://github.com/lib/pq) |
| Database migration | [golang-migrate](https://github.com/golang-migrate/migrate) |
| Data validation | [go-playground/validator/v10](https://github.com/go-playground/validator) |
| Logging | [uber-go/zap](https://github.com/uber-go/zap) |
| Metrics | [prometheus/client_golang](https://github.com/prometheus/client_golang) |
| API docs | [swaggo/swag](https://github.com/swaggo/swag) |

## Getting Started

If this is your first time encountering Go, please follow [the instructions](https://golang.org/doc/install) to
install Go on your computer. The kit requires **Go 1.24 or above**.

[Docker](https://www.docker.com/get-started) is also needed if you want to try the kit without setting up your
own database server.

After installing Go and Docker, run the following commands to start experiencing this starter kit:

```shell
# download the starter kit
git clone https://github.com/leoluyi/go-api-template.git

cd go-api-template

# start a PostgreSQL database server in a Docker container
make db-start

# seed the database with some test data
make testdata

# run the RESTful API server
make run

# or run the API server with live reloading, which is useful during development
# requires fswatch (https://github.com/emcrisostomo/fswatch)
make run-live
```

At this time, you have a RESTful API server running at `http://127.0.0.1:8080`. It provides the following endpoints:

- `GET /healthcheck`: returns JSON `{"status":"ok","version":"...","db":"ok"}` (HTTP 503 with `"status":"degraded"` if the database is unreachable)
- `GET /metrics`: Prometheus metrics scrape endpoint
- `POST /v1/login`: authenticates a user and generates a JWT
- `GET /v1/albums`: returns a paginated list of the albums
- `GET /v1/albums/:id`: returns the detailed information of an album
- `POST /v1/albums`: creates a new album
- `PUT /v1/albums/:id`: updates an existing album
- `DELETE /v1/albums/:id`: deletes an album
- `GET /v1/swagger/*`: Swagger UI for interactive API documentation

Try the URL `http://localhost:8080/healthcheck` in a browser, and you should see a JSON response like:

```json
{"status":"ok","version":"v2.7.0","db":"ok"}
```

If you have `cURL` or some API client tools (e.g. [Postman](https://www.getpostman.com/)), you may try the following
more complex scenarios:

```shell
# authenticate the user via: POST /v1/login
# credentials are set in config/local.yml (auth_username / auth_password)
curl -X POST -H "Content-Type: application/json" \
  -d '{"username": "demo", "password": "changeme"}' \
  http://localhost:8080/v1/login
# should return a JWT token like: {"token":"...JWT token here..."}

# with the above JWT token, access the album resources, such as: GET /v1/albums
curl -X GET -H "Authorization: Bearer ...JWT token here..." http://localhost:8080/v1/albums
# should return a list of album records in the JSON format

# scrape Prometheus metrics
curl http://localhost:8080/metrics
```

To use the starter kit as a starting point of a real project whose package name is `github.com/myorg/myproject`, do a global
replacement of the string `github.com/leoluyi/go-api-template` in all of project files with the string `github.com/myorg/myproject`.

## Project Layout

```
.
├── cmd                  main applications of the project
│   └── server           the API server application
├── config               configuration files for different environments
├── docs                 generated Swagger documentation
├── internal             private application and library code
│   ├── album            album-related features
│   ├── auth             authentication feature
│   ├── config           configuration library
│   ├── entity           entity definitions and domain logic
│   ├── errors           error types and handling
│   ├── healthcheck      healthcheck feature
│   └── test             helpers for testing purpose
├── migrations           database migrations
├── pkg                  public library code
│   ├── accesslog        access log middleware
│   ├── dbcontext        database context and transaction helpers
│   ├── log              structured and context-aware logger
│   ├── metrics          Prometheus metrics middleware and handler
│   └── pagination       paginated list
├── scripts              utility and operational scripts
└── testdata             test data scripts
```

The top level directories `cmd`, `internal`, `pkg` are commonly found in other popular Go projects, as explained in
[Standard Go Project Layout](https://github.com/golang-standards/project-layout).

Within `internal` and `pkg`, packages are structured by features in order to achieve the so-called
[screaming architecture](https://blog.cleancoder.com/uncle-bob/2011/09/30/Screaming-Architecture.html). For example,
the `album` directory contains the application logic related with the album feature.

Within each feature package, code is organized in layers (API, service, repository), following the dependency guidelines
as described in the [clean architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html).

## Common Development Tasks

This section describes some common development tasks using this starter kit.

### Implementing a New Feature

Implementing a new feature typically involves the following steps:

1. Develop the service that implements the business logic supporting the feature. Please refer to `internal/album/service.go` as an example.
2. Develop the RESTful API exposing the service about the feature. Please refer to `internal/album/api.go` as an example.
3. Develop the repository that persists the data entities needed by the service. Please refer to `internal/album/repository.go` as an example.
4. Wire up the above components together by injecting their dependencies in the main function. Please refer to
   the `album.RegisterHandlers()` call in `cmd/server/main.go`.

### Working with DB Transactions

It is the responsibility of the service layer to determine whether DB operations should be enclosed in a transaction.
The DB operations implemented by the repository layer should work both with and without a transaction.

You can use `dbcontext.DB.Transactional()` in a service method to enclose multiple repository method calls in
a transaction. For example,

```go
func serviceMethod(ctx context.Context, repo Repository, transactional dbcontext.TransactionFunc) error {
    return transactional(ctx, func(ctx context.Context) error {
        repo.method1(...)
        repo.method2(...)
        return nil
    })
}
```

If needed, you can also enclose method calls of different repositories in a single transaction. The return value
of the function in `transactional` above determines if the transaction should be committed or rolled back.

You can also use `dbcontext.DB.TransactionHandler()` as a middleware to enclose a whole API handler in a transaction.
This is especially useful if an API handler needs to put method calls of multiple services in a transaction.

### Updating Database Schema

The starter kit uses [database migration](https://en.wikipedia.org/wiki/Schema_migration) to manage the changes of the
database schema over the whole project development phase.

Migrations are embedded into the server binary at compile time and run automatically on startup. There is no separate
migration step required when deploying; the server applies any pending migrations before accepting requests.

The following commands are commonly used during local development:

```shell
# Execute new migrations made by you or other team members.
# Usually you should run this command each time after you pull new code from the code repo.
make migrate

# Create a new database migration.
# In the generated `migrations/*.up.sql` file, write the SQL statements that implement the schema changes.
# In the `*.down.sql` file, write the SQL statements that revert the schema changes.
make migrate-new

# Revert the last database migration.
# This is often used when a migration has some issues and needs to be reverted.
make migrate-down

# Clean up the database and rerun the migrations from the very beginning.
# Note that this command will first erase all data and tables in the database, and then
# run all migrations.
make migrate-reset
```

### Generating API Documentation

The kit uses [swaggo/swag](https://github.com/swaggo/swag) to generate Swagger documentation from annotations in the source code.

```shell
# Install swag CLI (first time only)
go install github.com/swaggo/swag/cmd/swag@latest

# Regenerate docs after updating annotations
make generate-docs
```

The generated docs are served at `http://localhost:8080/v1/swagger/` when the server is running.

### Managing Configurations

The application configuration is represented in `internal/config/config.go`. When the application starts,
it loads the configuration from a configuration file as well as environment variables. The path to the configuration
file is specified via the `-config` command line argument which defaults to `./config/local.yml`. Configurations
specified in environment variables should be named with the `APP_` prefix and in upper case. When a configuration
is specified in both a configuration file and an environment variable, the latter takes precedence.

The `config` directory contains the configuration files named after different environments. For example,
`config/local.yml` corresponds to the local development environment and is used when running the application
via `make run`.

The following fields are required and validated on startup:

| Field | Env var | Notes |
|-------|---------|-------|
| `dsn` | `APP_DSN` | PostgreSQL connection string |
| `jwt_signing_key` | `APP_JWT_SIGNING_KEY` | Must be **at least 32 characters** (HS256 requirement) |
| `auth_username` | `APP_AUTH_USERNAME` | Login username |
| `auth_password` | `APP_AUTH_PASSWORD` | Login password |

The following fields are optional:

| Field | Env var | Default | Notes |
|-------|---------|---------|-------|
| `server_port` | `APP_SERVER_PORT` | `8080` | |
| `jwt_expiration` | `APP_JWT_EXPIRATION` | `72` | Hours until token expires |
| `cors_allowed_origins` | — | `[]` (none) | Set to `["*"]` for dev, specific origins for prod |

Do not keep secrets in the configuration files. Provide them via environment variables instead. Secrets can be
populated from a secret store (e.g. HashiCorp Vault) into environment variables before the process starts.

## Deployment

The application can be run as a Docker container. Use `make build-docker` to build the image:

```shell
make build-docker
```

The container runs the server binary directly. Database migrations are applied automatically on startup before
the HTTP server begins accepting requests. Configure the application via `APP_`-prefixed environment variables:

```shell
docker run \
  -e APP_DSN="postgres://..." \
  -e APP_JWT_SIGNING_KEY="<at-least-32-char-secret>" \
  -e APP_AUTH_USERNAME="myuser" \
  -e APP_AUTH_PASSWORD="mypassword" \
  -e APP_CORS_ALLOWED_ORIGINS="https://app.example.com" \
  -p 8080:8080 server:latest
```

You can also run `make build` to build an executable binary named `server` and start it directly:

```shell
./server -config=./config/prod.yml
```

To run the full stack (API server + PostgreSQL) locally:

```shell
docker-compose up
```
