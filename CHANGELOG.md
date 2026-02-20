# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

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

[Unreleased]: https://github.com/leoluyi/go-api-template/compare/v2.0.0...HEAD
[v2.0.0]: https://github.com/leoluyi/go-api-template/compare/v1.0.1...v2.0.0
[v1.0.1]: https://github.com/leoluyi/go-api-template/compare/v1.0.0...v1.0.1
[v1.0.0]: https://github.com/leoluyi/go-api-template/releases/tag/v1.0.0
