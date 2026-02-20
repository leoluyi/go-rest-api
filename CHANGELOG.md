# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Changed
- `Version` in `cmd/server/main.go` now defaults to `"dev"`; the Makefile's `git describe --tags` ldflags injection is the sole owner of the real version at build time. Makefile fallback also changed from `"1.0.0"` to `"dev"` for consistency
- Swagger UI version now reflects the build-time `Version` variable: `docs.SwaggerInfo.Version` is overwritten at startup via `docs.SwaggerInfo.Version = Version` instead of relying on the static value embedded by `swag init`
- API route prefix (`/v1`) is now derived from `docs.SwaggerInfo.BasePath` (populated by the `@BasePath` annotation) instead of being hardcoded in `buildHandler`; the `@BasePath` annotation is the single source of truth for the versioned route prefix

### Fixed
- Corrected `.gitignore` pattern for the server binary from `./server` (invalid) to `/server` (rooted)

## [v2.7.0] - 2026-02-20

### Added
- Prometheus metrics at `GET /metrics`: `http_requests_total` (counter) and `http_request_duration_seconds` (histogram), labelled by method, chi route pattern, and status code (`pkg/metrics`)
- Health check endpoint now returns JSON `{"status","version","db"}` and pings the database on every request; returns HTTP 503 with `"status":"degraded"` when the DB is unreachable
- `request_id` field in all JSON error responses (populated from the `X-Request-ID` request header or a generated UUID); `X-Request-ID` response header set by the access-log middleware for client-side correlation

### Changed
- Auth credentials (`auth_username`, `auth_password`) are now loaded from config / `APP_AUTH_USERNAME` / `APP_AUTH_PASSWORD` env vars — hardcoded `demo`/`pass` credentials removed
- CORS is now configurable via `cors_allowed_origins` in config (replaces hardcoded `cors.AllowAll()`); dev configs default to `["*"]`
- JWT signing key requires a minimum of 32 characters (enforced by config validation)
- Database connection pool configured: `SetMaxOpenConns(25)`, `SetMaxIdleConns(5)`, `SetConnMaxLifetime(5m)`
- `tokenAuth` in auth middleware changed from a package-level global to a local variable (eliminates data race)
- JSON decode errors in album handlers now logged at `Error` level (was `Info`)
- `repository.Delete()` no longer issues a redundant `Get()` before deleting; uses `RowsAffected()` instead
- Transaction rollback/commit errors in `dbcontext` now logged to stderr instead of silently discarded
- `json.Encoder` error in `RespondJSON` now logged instead of silently discarded

### Fixed
- Thread-safety: removed package-level `tokenAuth` global from `internal/auth/middleware.go`

### Database
- New migration (`20260220000000`): adds `VARCHAR(36)` constraint on `album.id`, `VARCHAR(128)` on `album.name`, and indexes on `album.created_at` and `album.updated_at`

## [v2.6.0] - 2026-02-21

### Changed
- Upgraded all Go dependencies to latest versions
- `go-chi/chi/v5` v5.0.7 → v5.2.5
- `go-chi/jwtauth/v5` v5.0.2 → v5.3.3
- `go-chi/cors` v1.2.1 → v1.2.2
- `lib/pq` v1.10.9 → v1.11.2
- `go.uber.org/zap` v1.23.0 → v1.27.1
- `stretchr/testify` v1.10.0 → v1.11.1
- `golang-migrate/migrate/v4` moved from indirect to direct dependency

### Fixed
- Updated `internal/auth/middleware.go` for breaking API change in `go-chi/jwtauth/v5` v5.3.x: `Authenticator` now takes `*JWTAuth` instead of `http.Handler`

## [v2.5.0] - 2026-02-21

### Changed
- Upgraded PostgreSQL image from `postgres:17-alpine` to `postgres:18-alpine` across Makefile, docker-compose.yml, and CI workflow

## [v2.4.0] - 2026-02-21

### Changed
- Updated CI PostgreSQL service image from `postgres:10.8` to `postgres:17-alpine`
- Updated CI Go version from 1.13 to 1.24
- Updated `actions/checkout` from v1 to v4
- Updated `actions/setup-go` from v1 to v5
- Updated `codecov/codecov-action` from v1 to v5
- Removed manual `go get` steps for deprecated tools (`golint`, `goveralls`, `cover`)

## [v2.3.0] - 2026-02-21

### Changed
- Replaced `golang:1.24-alpine` build image with `registry.suse.com/bci/golang:1.24`
- Replaced `alpine:3.21` runtime image with `registry.suse.com/bci/bci-micro:latest` (SUSE BCI distroless-style image — no shell, no package manager)
- Runtime container now runs as UID/GID `65532:65532` (non-root, no `/etc/passwd` dependency)
- Removed `RUN apk add ca-certificates` from build stage (BCI golang image includes CA certificates)
- Removed `addgroup`/`adduser` and `chown` RUN layers from runtime stage; ownership set via `COPY --chown`

## [v2.2.0] - 2026-02-21

### Added
- `migrations/migrations.go` embeds all `*.sql` migration files into the binary via `//go:embed`, making the binary fully self-contained
- Database migrations now run automatically at startup via `golang-migrate/migrate/v4` library (iofs source + postgres driver), eliminating the need for the `migrate` CLI binary or `entrypoint.sh`

### Changed
- Removed `cmd/server/entrypoint.sh` from Docker image — `CMD` is now `["./server"]` directly
- Removed `migrate` binary download from the Dockerfile build stage
- Removed `bash` from the runtime Alpine image (no longer needed without `entrypoint.sh`)
- Removed `/var/log/app` directory creation (was only used by a commented-out log redirect in `entrypoint.sh`)
- Removed `APP_ENV` from `docker-compose.yml` (was only consumed by `entrypoint.sh`)
- Removed log volume mount from `docker-compose.yml` (`/tmp/app:/var/log/app`)
- Added `.dockerignore` to exclude `.git`, `vendor`, `testdata`, `docs`, `tmp`, and other non-essential files from the build context

## [v2.1.0] - 2026-02-21

### Changed
- Moved `stress-test.sh` to `scripts/stress-test.sh`
- Added `make stress-test` Makefile target

### Fixed
- Set `SHELL := /bin/bash` in Makefile to fix bash-specific syntax (`&>/dev/null`, `read -p`) under `/bin/sh`
- Added missing `.PHONY` for `run-live` target
- Replaced bare `make` with `$(MAKE)` in `testdata` target
- Replaced backtick command substitution with `$$()` in `run-restart`
- Dropped `-t` from `docker exec` in `testdata` target to allow non-TTY execution (e.g. CI)
- Added help comment to `generate-docs` target so it appears in `make help`
- Fixed `/swagger*` route to `/swagger/*` for correct chi wildcard sub-path matching
- Updated Swagger `@host` to `localhost:8080` and replaced placeholder title/description
- Added `BearerAuth` security definition to Swagger spec
- Moved swag annotations from `RegisterHandlers` onto individual handler functions with proper `@Param`, `@Success`, `@Failure`, and `@Security` tags
- Fixed login route annotation from `/login/{id}` to `/login`
- Extracted anonymous login structs to named `LoginRequest` and `LoginResponse` types
- Upgraded `swaggo/swag` to v1.16.6 and `swaggo/http-swagger` to v1.3.4 to match swag CLI; regenerated docs

## [v2.0.0] - 2026-02-21

### Added
- Swagger API documentation via `swaggo/swag` and `swaggo/http-swagger`
- Swagger UI served at `GET /v1/swagger/*`
- `make generate-docs` target to regenerate Swagger docs from annotations
- CORS support via `go-chi/cors`
- `docker-compose.yml` for running the full stack locally
- `stress-test.sh` for basic load testing
- `make run-with-db` target to start database and server together
- CLAUDE.md project guidance for Claude Code

### Changed
- Replaced `ozzo-routing` with `go-chi/chi/v5` for HTTP routing
- Replaced `ozzo-dbx` with `jmoiron/sqlx` + `lib/pq` for database access
- Replaced `ozzo-validation` with `go-playground/validator/v10` for input validation
- Replaced `dgrijalva/jwt-go` with `golang-jwt/jwt/v4` + `go-chi/jwtauth/v5` for JWT auth
- Updated module path to `github.com/leoluyi/go-api-template`
- Upgraded minimum Go version to 1.24
- Upgraded PostgreSQL Docker image to `postgres:14.10` in `make db-start`
- Updated README to reflect current dependencies, module path, and endpoints

### Fixed
- Removed PostgreSQL data volume mount from `db-start` to avoid permission issues
- Resolved test failures after ozzo package migration
- Fixed Docker image versioning

## [v1.0.1] - 2020-02-19

### Fixed
- Use symmetric key for JWT signing

## [v1.0.0] - 2020-01-31

Initial release (upstream: [qiangxue/go-rest-api](https://github.com/qiangxue/go-rest-api)).

### Features
- RESTful CRUD endpoints for albums resource
- JWT-based authentication
- PostgreSQL via ozzo-dbx
- Database migrations via golang-migrate
- Structured logging via uber-go/zap
- Environment-based configuration
- Graceful shutdown
- Full test coverage with mock-based unit tests

[Unreleased]: https://github.com/leoluyi/go-api-template/compare/v2.7.0...HEAD
[v2.7.0]: https://github.com/leoluyi/go-api-template/compare/v2.6.0...v2.7.0
[v2.6.0]: https://github.com/leoluyi/go-api-template/compare/v2.5.0...v2.6.0
[v2.5.0]: https://github.com/leoluyi/go-api-template/compare/v2.4.0...v2.5.0
[v2.4.0]: https://github.com/leoluyi/go-api-template/compare/v2.3.0...v2.4.0
[v2.3.0]: https://github.com/leoluyi/go-api-template/compare/v2.2.0...v2.3.0
[v2.2.0]: https://github.com/leoluyi/go-api-template/compare/v2.1.0...v2.2.0
[v2.1.0]: https://github.com/leoluyi/go-api-template/compare/v2.0.0...v2.1.0
[v2.0.0]: https://github.com/leoluyi/go-api-template/compare/v1.0.1...v2.0.0
[v1.0.1]: https://github.com/leoluyi/go-api-template/compare/v1.0.0...v1.0.1
[v1.0.0]: https://github.com/leoluyi/go-api-template/releases/tag/v1.0.0
