ifneq ($(wildcard .env),)
include .env
export
else
$(warning WARNING: .env file not found! Using .env.example)
include .env.example
export
endif

BASE_STACK = docker compose -f docker-compose.yml
INTEGRATION_TEST_STACK = $(BASE_STACK) -f docker-compose-integration-test.yml
ALL_STACK = $(INTEGRATION_TEST_STACK)

# HELP =================================================================================================================
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## Display this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
compose-up: ### Run docker compose (core services: database)
	$(BASE_STACK) up --build -d
.PHONY: compose-up

compose-up-all: ### Run docker compose (full stack with backend and reverse proxy)
	$(BASE_STACK) up --build -d
.PHONY: compose-up-all

compose-up-integration-test: ### Run docker compose with integration test
	$(INTEGRATION_TEST_STACK) up --build --abort-on-container-exit --exit-code-from integration-test
.PHONY: compose-up-integration-test

compose-down: ### Stop docker compose
	$(ALL_STACK) down --remove-orphans
.PHONY: compose-down

swag-v1: ### Generate Swagger documentation
	swag init -g internal/controller/restapi/router.go
.PHONY: swag-v1

deps: ### Tidy and verify Go dependencies
	go mod tidy && go mod verify
.PHONY: deps

deps-audit: ### Check Go dependencies for vulnerabilities
	govulncheck ./...
.PHONY: deps-audit

format: ### Format Go code (gofumpt + gci)
	gofumpt -l -w .
	gci write . --skip-generated -s standard -s default
.PHONY: format

run: ### Start backend (deps + swagger)
	@$(MAKE) deps
	@$(MAKE) swag-v1
	go mod download && \
	CGO_ENABLED=0 go run -tags migrate ./cmd/app
.PHONY: run

##@ Frontend
web-dev: ### Start Vue frontend dev server
	cd web && npm install && npm run dev
.PHONY: web-dev

web-build: ### Build Vue frontend for production
	cd web && npm run build
.PHONY: web-build

##@ Docker
docker-rm-volume: ### Remove Docker volume ( Postgres data)
	docker volume rm litedock_pg-data
.PHONY: docker-rm-volume

##@ Linting
linter-golangci: ### Run golangci-lint
	golangci-lint run
.PHONY: linter-golangci

linter-hadolint: ### Run hadolint on Dockerfiles
	git ls-files --exclude='Dockerfile*' --ignored | xargs hadolint
.PHONY: linter-hadolint

linter-dotenv: ### Run dotenv-linter
	dotenv-linter
.PHONY: linter-dotenv

##@ Testing
test: ### Run unit tests
	go test -v -race -covermode atomic -coverprofile=coverage.txt ./internal/... ./pkg/...
.PHONY: test

integration-test: ### Run integration tests
	go clean -testcache && go test -v ./integration-test/...
.PHONY: integration-test

##@ Mocks
mock: ### Generate mocks with mockgen
	mockgen -source ./internal/repo/contracts.go -package usecase_test > ./internal/usecase/mocks_repo_test.go
	mockgen -source ./internal/usecase/contracts.go -package usecase_test > ./internal/usecase/mocks_usecase_test.go
.PHONY: mock

##@ Database
migrate-create:  ### Create new migration (usage: make migrate-create name=create_users)
	go run -tags 'postgres mysql' github.com/golang-migrate/migrate/v4/cmd/migrate create \
		-ext sql -dir migrations '$(word 2,$(MAKECMDGOALS))'
.PHONY: migrate-create

migrate-up: ### Apply migrations
	@if echo '$(DB_URL)' | grep -q 'mysql://'; then \
		go run -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate \
			-path migrations \
			-database '$(DB_URL)' up; \
	else \
		go run -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate \
			-path migrations \
			-database '$(DB_URL)?sslmode=disable' up; \
	fi
.PHONY: migrate-up

##@ Pre-commit
pre-commit: deps swag-v1 mock format linter-golangci test ### Run all pre-commit checks
.PHONY: pre-commit
