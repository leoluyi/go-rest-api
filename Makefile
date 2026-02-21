SHELL := /bin/bash

MODULE = $(shell go list -m)
VERSION ?= $(shell git describe --tags --always --dirty --match=v* 2> /dev/null || echo "dev")
PACKAGES := $(shell go list ./... | grep -v /vendor/)
LDFLAGS := -ldflags "-s -w -X main.Version=${VERSION}"

CONFIG_FILE ?= ./config/local.yml
APP_DSN ?= $(shell sed -n 's/^dsn:[[:space:]]*"\(.*\)"/\1/p' $(CONFIG_FILE))
MIGRATE := docker run --rm --name migrate -v $(shell pwd)/migrations:/migrations --network host migrate/migrate:v4.19.1 -path=/migrations/ -database "$(APP_DSN)"

PID_FILE := './.pid'

.PHONY: default
default: help

# generate help info from comments: thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help: ## help information about make commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: test
test: ## run unit tests
	@echo "mode: atomic" > coverage-all.out
	@$(foreach pkg,$(PACKAGES), \
		go test -p=1 -race -cover -covermode=atomic -coverprofile=coverage.out ${pkg}; \
		tail -n +2 coverage.out >> coverage-all.out;)

.PHONY: test-cover
test-cover: test ## run unit tests and show test coverage information
	go tool cover -html=coverage-all.out

.PHONY: generate-docs
generate-docs: ## generate Swagger API documentation
	@swag init -g cmd/server/main.go

.PHONY: run
run: ## run the API server
	go run ${LDFLAGS} cmd/server/main.go

.PHONY: run-with-db
run-with-db: db-start ## run the API server with database server
	go run ${LDFLAGS} cmd/server/main.go

.PHONY: run-restart
run-restart: ## restart the API server
	@pkill -P $$(cat $(PID_FILE)) || true
	@printf '%*s\n' "80" '' | tr ' ' -
	@echo "Source file changed. Restarting server..."
	@go run ${LDFLAGS} cmd/server/main.go & echo $$! > $(PID_FILE)
	@printf '%*s\n' "80" '' | tr ' ' -

.PHONY: run-live
run-live: ## run the API server with live reload support (requires fswatch)
	@go run ${LDFLAGS} cmd/server/main.go & echo $$! > $(PID_FILE)
	@fswatch -x -o --event Created --event Updated --event Renamed -r internal pkg cmd config | xargs -n1 -I {} make run-restart

.PHONY: build
build:  ## build the API server binary
	CGO_ENABLED=0 go build -trimpath ${LDFLAGS} -a -o server $(MODULE)/cmd/server

.PHONY: build-docker
build-docker: ## build the API server as a docker image
	docker build -f cmd/server/Dockerfile -t server:latest -t server:$(VERSION) .

.PHONY: clean
clean: ## remove temporary files
	rm -rf server coverage.out coverage-all.out

.PHONY: version
version: ## display the version of the API server
	@echo $(VERSION)

.PHONY: db-start
db-start: ## start the database server
	docker rm -f postgres &>/dev/null || true
	docker run -d --rm \
	  --name postgres \
	  -e POSTGRES_DB=go_restful \
	  -e POSTGRES_PASSWORD=postgres \
	  -p 5432:5432 \
	  -v $(shell pwd)/testdata:/testdata \
	  postgres:17-alpine
	@until docker exec postgres pg_isready -U postgres -q 2>/dev/null; do sleep 1; done

.PHONY: db-stop
db-stop: ## stop the database server
	docker stop postgres

.PHONY: testdata
testdata: ## populate the database with test data
	$(MAKE) migrate-reset
	@echo "Populating test data..."
	@docker exec -i postgres psql "$(APP_DSN)" -f /testdata/testdata.sql

.PHONY: lint
lint: ## run golint on all Go package
	@golint $(PACKAGES)

.PHONY: fmt
fmt: ## run "go fmt" on all Go packages
	@go fmt $(PACKAGES)

.PHONY: migrate
migrate: ## run all new database migrations
	@echo "Running all new database migrations..."
	@$(MIGRATE) up

.PHONY: migrate-down
migrate-down: ## revert database to the last migration step
	@echo "Reverting database to the last migration step..."
	@$(MIGRATE) down 1

.PHONY: migrate-new
migrate-new: ## create a new database migration
	@read -p "Enter the name of the new migration: " name; \
	$(MIGRATE) create -ext sql -dir /migrations/ $${name// /_}

.PHONY: stress-test
stress-test: ## run stress test against the API server (requires ab)
	@scripts/stress-test.sh

.PHONY: migrate-reset
migrate-reset: ## reset database and re-run all migrations
	@echo "Resetting database..."
	@$(MIGRATE) drop -f
	@echo "Running all database migrations..."
	@$(MIGRATE) up

.PHONY: vendor
vendor: ## tidy go.mod and sync the vendor directory
	go mod tidy
	go mod vendor

.PHONY: vuln
vuln: ## scan for known vulnerabilities in dependencies (requires govulncheck)
	govulncheck ./...

.PHONY: install-tools
install-tools: ## install development tools (govulncheck, swag)
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: pre-commit
pre-commit: fmt lint test vuln ## run all checks before committing
