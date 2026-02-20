# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Go RESTful API starter kit demonstrating clean architecture with PostgreSQL.

Module: `github.com/qiangxue/go-rest-api`

## Commands

### Development

```bash
make run              # Start API server on :8080
make run-live         # Live reload (requires fswatch)
make run-with-db      # Start server + Docker PostgreSQL
```

### Build

```bash
make build            # Build optimized binary to ./tmp/server
make build-docker     # Build Docker image
```

### Testing

```bash
make test             # Run all tests with coverage (outputs coverage-all.out)
make test-cover       # Run tests and open HTML coverage report
go test ./internal/album/... -run TestAlbumAPI  # Run a single test

# Run only tests that don't require PostgreSQL
go test $(go list ./... | grep -v 'pkg/dbcontext\|internal/album$\|internal/test$\|cmd/server\|docs')
```

### Linting & Formatting

```bash
make lint             # golint
make fmt              # gofmt
```

### Database

```bash
make db-start         # Start PostgreSQL container (NOTE: the Makefile mounts
                      # testdata/postgres as the data dir; if that volume contains
                      # stale data the container exits immediately. If db-start fails,
                      # run: docker run -d --name postgres -e POSTGRES_DB=go_restful \
                      #   -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres:14.10)
make db-stop          # Stop PostgreSQL container
make migrate          # Run pending migrations
make migrate-new      # Create new migration file
make migrate-down     # Revert last migration
make migrate-reset    # Reset and replay all migrations
make testdata         # Reset DB and load sample fixtures
```

### Documentation

```bash
make generate-docs    # Generate Swagger docs via swag
```

## Key Dependencies

| Concern | Package |
|---------|---------|
| Router | `github.com/go-chi/chi/v5` |
| JWT middleware | `github.com/go-chi/jwtauth/v5` |
| JWT signing | `github.com/golang-jwt/jwt/v4` |
| Database | `github.com/jmoiron/sqlx` + `github.com/lib/pq` |
| Validation | `github.com/go-playground/validator/v10` |
| Logging | `go.uber.org/zap` (wrapped by `pkg/log`) |

## Architecture

Each feature follows a strict three-layer pattern. New features must mirror this structure:

```
internal/<feature>/
  api.go         # HTTP handlers, routing, request/response structs
  service.go     # Business logic, validation (depends on repository interface)
  repository.go  # Database queries (implements repository interface)
  *_test.go      # Tests for each layer
```

Shared domain models live in `internal/entity/`. Cross-cutting packages (logging, pagination, DB context) live in `pkg/`.

**Dependency direction:** `api` → `service` → `repository` → `entity`. Each layer depends on interfaces, not concrete types — enabling mock-based unit testing.

### Request flow

`chi router` → auth middleware → handler (`api.go`) → service (`service.go`) → repository (`repository.go`) → PostgreSQL via `sqlx`

### Key interfaces

Each feature defines its own `Repository` and `Service` interfaces in `service.go`. Tests mock these interfaces using structs in `internal/test/mock.go`.

## Configuration

Config file: `./config/local.yml` (selected via `APP_ENV` env var; defaults to `local`)

```yaml
dsn: "postgres://127.0.0.1/go_restful?sslmode=disable&user=postgres&password=postgres"
jwt_signing_key: "..."
```

All config fields can be overridden with `APP_`-prefixed environment variables (e.g., `APP_DSN`, `APP_SERVER_PORT`).

## Routing

Routes are registered in `cmd/server/main.go`. The chi router is used with middleware stacked in this order: access log → CORS → error handler → JWT auth (on protected routes).

Protected routes under `/v1/` require a `Bearer` JWT token. The JWT middleware is in `internal/auth/middleware.go`.

## Testing Patterns

- Unit tests mock the layer below via interfaces (no real DB required)
- Integration tests use helpers in `internal/test/` (`api.go` for HTTP, `db.go` for DB setup)
- Test fixtures in `testdata/testdata.sql`
- Run a specific test: `go test ./internal/<pkg>/... -run <TestName>`
