# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go RESTful API starter kit demonstrating clean architecture with PostgreSQL.

- **Module:** `github.com/leoluyi/go-api-template`
- **Go version:** 1.24.0
- **Default port:** 8080

## Commands

### Development

```bash
make run              # Start API server on :8080
make run-live         # Live reload (requires fswatch)
make run-with-db      # Start PostgreSQL container then start server
```

### Build

```bash
make build            # Build optimized binary to ./server (CGO_ENABLED=0)
make build-docker     # Build Docker image (two-stage, distroless runtime)
make clean            # Remove ./server, coverage.out, coverage-all.out
make version          # Show version derived from git describe --tags
```

### Testing

```bash
make test             # Run all tests with coverage (outputs coverage-all.out)
make test-cover       # Run tests and open HTML coverage report

# Run a single test suite
go test ./internal/album/... -run TestAPI

# Run only tests that don't require PostgreSQL (unit tests only)
go test $(go list ./... | grep -v 'pkg/dbcontext\|internal/album$\|internal/test$\|cmd/server\|docs')
```

### Linting & Formatting

```bash
make lint             # golint on all packages
make fmt              # gofmt on all packages
```

### Database

```bash
make db-start         # Start postgres:18-alpine Docker container on 5432
make db-stop          # Stop the container
make migrate          # Run pending migrations (via migrate/migrate Docker image)
make migrate-new      # Interactively create a new migration file
make migrate-down     # Revert the last migration
make migrate-reset    # Drop all tables, re-run all migrations
make testdata         # migrate-reset then load testdata/testdata.sql
```

### Documentation

```bash
make generate-docs    # Generate/update docs/ via swag init -g cmd/server/main.go
```

## Key Dependencies

| Concern | Package |
|---------|---------|
| Router | `github.com/go-chi/chi/v5` v5.2.5 |
| CORS | `github.com/go-chi/cors` v1.2.2 |
| JWT middleware | `github.com/go-chi/jwtauth/v5` v5.3.3 |
| JWT signing | `github.com/golang-jwt/jwt/v4` v4.5.2 |
| Database | `github.com/jmoiron/sqlx` + `github.com/lib/pq` |
| Migrations | `github.com/golang-migrate/migrate/v4` v4.19.1 |
| Validation | `github.com/go-playground/validator/v10` v10.30.1 |
| Logging | `go.uber.org/zap` v1.27.1 (wrapped by `pkg/log`) |
| Metrics | `github.com/prometheus/client_golang` v1.23.2 |
| ID generation | `github.com/google/uuid` v1.6.0 |
| Env config | `github.com/qiangxue/go-env` v1.0.1 |
| Swagger docs | `github.com/swaggo/swag` + `github.com/swaggo/http-swagger` |
| Testing | `github.com/stretchr/testify` v1.11.1 |

## Directory Structure

```
.
├── cmd/server/
│   ├── main.go          # Entry point: config, DB, migrations, HTTP server setup
│   └── Dockerfile       # Two-stage build (golang:1.24 → bci-micro distroless)
├── config/
│   ├── local.yml        # Default dev config (APP_ENV=local)
│   ├── dev.yml
│   ├── qa.yml
│   └── prod.yml
├── docs/                # Auto-generated Swagger/OpenAPI (do not edit manually)
├── internal/
│   ├── album/           # Example CRUD feature (api, service, repository + tests)
│   ├── auth/            # JWT login + middleware + context helpers
│   ├── config/          # Config struct, YAML + env-var loading
│   ├── entity/          # Domain models: Album, User, GenerateID()
│   ├── errors/          # ErrorResponse type, HTTP helpers, panic-recovery middleware
│   ├── healthcheck/     # GET /healthcheck with DB ping
│   └── test/            # Shared test helpers: MockRouter, APITestCase, Endpoint()
├── migrations/          # Embedded SQL migrations (*.up.sql / *.down.sql)
├── pkg/
│   ├── accesslog/       # HTTP request/response logging middleware
│   ├── dbcontext/       # DB wrapper with transaction propagation via context
│   ├── log/             # zap-based structured logger with context correlation
│   ├── metrics/         # Prometheus middleware + /metrics handler
│   └── pagination/      # Pages struct, NewFromRequest(), Offset(), Limit()
├── scripts/
│   └── stress-test.sh   # Apache Bench stress testing
├── testdata/
│   └── testdata.sql     # 5 sample album fixtures
├── docker-compose.yml   # Full-stack: server + postgres:18-alpine
├── Makefile
└── go.mod / go.sum
```

## Architecture

Each feature follows a strict three-layer pattern. New features **must** mirror this structure:

```
internal/<feature>/
  api.go         # HTTP handlers, route registration, request/response types
  service.go     # Business logic, validation, Service & Repository interfaces
  repository.go  # SQL queries implementing the Repository interface
  *_test.go      # One test file per layer
```

Shared domain models live in `internal/entity/`. Cross-cutting packages (logging, pagination, DB context) live in `pkg/`.

**Dependency direction:** `api` → `service` → `repository` → `entity`

Each layer depends only on interfaces, not concrete types — this enables mock-based unit testing without a real database.

### Adding a New Feature

1. Create `internal/<feature>/` with `api.go`, `service.go`, `repository.go`
2. Define `Service` and `Repository` interfaces in `service.go`
3. Implement the repository against `*dbcontext.DB` (use `r.db.With(ctx)` to honour transactions)
4. Implement the service, inject the repository interface
5. In `api.go`, implement `RegisterHandlers(r chi.Router, svc Service, ...)` and add Swagger annotations
6. Wire up in `cmd/server/main.go` inside `buildHandler()`
7. Add domain model to `internal/entity/` if needed
8. Create migration in `migrations/` with `make migrate-new`

### Request Flow

```
chi router
  → accesslog middleware
  → errors (panic recovery) middleware
  → metrics middleware
  → AllowContentType("application/json")
  → cors middleware
  → handler (api.go)
      → service (service.go)
          → repository (repository.go)
              → PostgreSQL via sqlx
```

### Middleware Stack Order (in `buildHandler`)

The order in `cmd/server/main.go:146` is:
1. `accesslog.Handler` — logs every request with timing and request ID
2. `errors.Handler` — panic recovery, maps errors to JSON responses
3. `metrics.Middleware` — Prometheus counter and histogram
4. `middleware.AllowContentType("application/json")` — rejects non-JSON bodies
5. `cors.New(...).Handler` — CORS headers

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/healthcheck` | None | Service status + DB ping |
| HEAD | `/healthcheck` | None | Same as GET (no body) |
| GET | `/metrics` | None | Prometheus metrics scrape |
| GET | `/v1/swagger/*` | None | Swagger UI |
| POST | `/v1/login` | None | Returns JWT token |
| GET | `/v1/albums` | None | Paginated album list |
| GET | `/v1/albums/{id}` | None | Single album |
| POST | `/v1/albums` | Bearer JWT | Create album |
| PUT | `/v1/albums/{id}` | Bearer JWT | Update album |
| DELETE | `/v1/albums/{id}` | Bearer JWT | Delete album |

Pagination query params: `?page=1&per_page=100` (default 100, max 1000).

## Configuration

Config file selected by `APP_ENV` env var (defaults to `local` → `./config/local.yml`).
Override the config file path with the `-config` CLI flag.

```yaml
# config/local.yml — all fields shown with their defaults
server_port: 8080                          # optional; default 8080
dsn: "postgres://..."                      # required
jwt_signing_key: "..."                     # required; min 32 chars (HS256)
jwt_expiration: 72                         # optional; hours, default 72
auth_username: "demo"                      # required
auth_password: "changeme"                  # required
cors_allowed_origins: ["*"]               # optional; use specific origins in prod
```

All fields can be overridden with `APP_`-prefixed environment variables:
`APP_DSN`, `APP_SERVER_PORT`, `APP_JWT_SIGNING_KEY`, `APP_JWT_EXPIRATION`,
`APP_AUTH_USERNAME`, `APP_AUTH_PASSWORD`.

## Database & Migrations

Migrations live in `migrations/` as paired `*.up.sql` / `*.down.sql` files.
They are **embedded into the binary** at compile time via `//go:embed` in `migrations/embed.go`.
The server runs `migrate.Up()` automatically on startup before accepting requests.

Current migrations:
- `20191217202658_init` — creates `album` table
- `20260220000000_add_indexes_and_constraints` — widens columns, adds indexes on `created_at`/`updated_at`

To create a new migration:
```bash
make migrate-new   # prompts for name, creates timestamped up/down files
```

Connection pool settings (hardcoded in `cmd/server/main.go`):
- Max open connections: 25
- Max idle connections: 5
- Connection max lifetime: 5 minutes

## Error Handling

`internal/errors/response.go` provides `ErrorResponse`:

```go
type ErrorResponse struct {
    Status    int         `json:"status"`
    Message   string      `json:"message"`
    RequestID string      `json:"request_id,omitempty"`
    Details   interface{} `json:"details,omitempty"`
}
```

Constructor helpers: `InternalServerError`, `NotFound`, `Unauthorized`, `Forbidden`, `BadRequest`, `InvalidInput`.

Use `errors.RespondWithError(w, r, err)` in handlers — it automatically attaches the request ID from context.

Validation errors from `go-playground/validator` are converted via `errors.InvalidInput(errs)`, which produces a `details` array of `{field, error}` objects.

## Testing Patterns

- **Unit tests** mock the layer below via interfaces — no real DB required for `api_test.go` and `service_test.go`
- **Integration tests** (`repository_test.go`) require a live PostgreSQL instance
- `internal/test/mock.go` provides `MockRouter()` for HTTP handler tests
- `internal/auth/middleware.go` provides `auth.MockAuthHandler` and `auth.MockAuthHeader()` for protected route tests
- Fixtures in `testdata/testdata.sql`

### Test Helpers

```go
// Create a router with standard middleware (accesslog, errors, cors)
router := test.MockRouter(logger)

// Register handlers with mock auth
RegisterHandlers(router, NewService(repo, logger), auth.MockAuthHandler, logger)

// Make authenticated requests
header := auth.MockAuthHeader()   // sets Authorization: TEST

// Run table-driven HTTP tests
test.Endpoint(t, router, test.APITestCase{
    Name: "get 123", Method: "GET", URL: "/albums/123",
    WantStatus: http.StatusOK, WantResponse: `*album123*`,
})
```

`WantResponse` supports glob patterns (`*`) for flexible matching.

### Running Tests

```bash
# All tests (requires PostgreSQL at APP_DSN)
make test

# Unit tests only (no DB needed)
go test $(go list ./... | grep -v 'pkg/dbcontext\|internal/album$\|internal/test$\|cmd/server\|docs')

# Single test
go test ./internal/album/... -run TestAPI
```

## DB Transactions

`pkg/dbcontext.DB` wraps `sqlx.DB` and propagates transactions via context:

```go
// In a repository method — automatically uses tx if present in ctx
r.db.With(ctx).ExecContext(ctx, query, args...)

// In a service or handler — wrap multiple repo calls in one transaction
err := db.Transactional(ctx, func(ctx context.Context) error {
    return repo1.Create(ctx, ...) // uses same tx
})

// As middleware (wraps entire request in a transaction)
r.Use(db.TransactionHandler())
```

## Routing

Routes are registered in `cmd/server/main.go` inside `buildHandler()`.

- Global routes (no auth): `/healthcheck`, `/metrics`
- All API routes are under the `basePath` prefix (currently `/v1`, from `docs.SwaggerInfo.BasePath`)
- Protected routes use `auth.Handler(jwtSigningKey)` middleware, which chains jwtauth Verifier + Authenticator
- The current user identity is stored in context via `auth.WithUser()` and retrieved with `auth.CurrentUser(ctx)`

## Logging

`pkg/log` wraps `go.uber.org/zap`. Use the `log.Logger` interface throughout:

```go
logger.Info("message")
logger.With(ctx).Infof("request %s processed", id)   // attaches request/correlation ID from ctx
```

In tests:
```go
logger, _ := log.NewForTest()   // captures log output for assertions
```

## Swagger / API Docs

Swagger annotations use `//go:generate`-style swag comments in handler files.
Run `make generate-docs` to regenerate `docs/`. The Swagger UI is served at `/v1/swagger/`.

The Swagger host and version are set dynamically:
- Host: from `docs.SwaggerInfo` (default `localhost:8080`)
- Version: injected at build time via `-ldflags "-X main.Version=<tag>"`

## Docker

`docker-compose.yml` starts two services:
- `db` — `postgres:18-alpine` with healthcheck, DB name `go_restful`
- `server` — built from `cmd/server/Dockerfile`, depends on `db` being healthy, `APP_DSN` points to `db`

The Dockerfile uses a two-stage build:
1. **Build stage** (`golang:1.24` SUSE BCI): compiles binary with `CGO_ENABLED=0`
2. **Runtime stage** (`bci-micro` distroless): copies binary + config, runs as UID/GID 65532

## CI/CD

`.github/workflows/build.yml` runs on every push and pull request:
1. Checks out code with `actions/checkout@v4`
2. Sets up Go 1.24 with `actions/setup-go@v5`
3. Builds binary (`make build`)
4. Runs migrations against PostgreSQL service container (`make migrate`)
5. Runs all tests with coverage (`make test-cover`)
6. Uploads coverage to Codecov (requires `CODECOV_TOKEN` secret)

## Conventions

- IDs are UUIDs generated by `entity.GenerateID()` (`github.com/google/uuid`)
- All JSON field names use `snake_case`
- `db` struct tags on entity fields map to PostgreSQL column names
- Requests to protected endpoints need `Authorization: Bearer <token>` header
- All API responses are `Content-Type: application/json`
- Requests **must** send `Content-Type: application/json` for POST/PUT (enforced by `AllowContentType` middleware)
- Error responses always include `status` (HTTP code) and `message`; optionally `request_id` and `details`
- Pagination response wraps items in `{page, per_page, page_count, total_count, items}`
