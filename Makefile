BUILD_FLAGS     :=
GOBIN           := $(shell go env GOPATH)/bin
THRESHOLD       := 15

BINARY_NAME=hw-4
MAIN_FILE=cmd/main.go
GO_FILES=$(wildcard *.go)

DOCKER_IMAGE=hw-4
DOCKER_TAG=latest

CONFIG_FILE=config/config.yaml
MIGRATIONS_FOLDER=migrations
CONFIG_EXAMPLE=config.yaml.example

PROTO_PATH = proto
PROTO_SRC = $(PROTO_PATH)/order_service.proto
PROTO_DEST = internal/api/grpc/generated

ifeq ($(POSTGRES_SETUP_TEST),)
    POSTGRES_SETUP_TEST = user=test password=test dbname=test host=localhost port=5433 sslmode=disable
endif

ifeq ($(POSTGRES_SETUP_MAIN),)
    POSTGRES_SETUP_MAIN = user=test password=test dbname=test host=localhost port=5432 sslmode=disable
endif

MIGRATION_FOLDER := ./migrations

all: lint build

docker-build:
	@echo "Building from Docker Compose file..."
	@docker-compose build

docker-run: docker-build
	docker-compose up -d

docker-down:
	@echo "Stopping Docker container..."
	@docker-compose down

build:
	@echo "Building..."
	@go build -o $(BINARY_NAME) $(MAIN_FILE)
	@docker-compose build memcached
	sleep 3
	@docker-compose build db

run-background:
	@docker-compose up memcached -d
	@docker-compose up db -d
	@docker-compose up zookeeper -d
	@docker-compose up kafka -d
	sleep 3
	@make migration-up

down-background: docker-down

run: run-background
	./$(BINARY_NAME)

down:
	@make migration-down
	@docker-compose down

deps:
	go mod tidy
	go mod download

lint: install-linters
	golangci-lint run --config=.golangci.yml

install-linters:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
	go install github.com/uudashr/gocognit/cmd/gocognit@latest
	go install github.com/kisielk/errcheck@latest
	go install github.com/jgautheron/goconst/cmd/goconst@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install github.com/gordonklaus/ineffassign@latest
	go install github.com/mgechev/revive@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest

clean:
	rm -f $(BINARY_NAME)

migration-create:
	goose -dir $(MIGRATION_FOLDER) create $(name) sql

test-migration-up:
	goose -dir $(MIGRATION_FOLDER) postgres "$(POSTGRES_SETUP_TEST)" up

migration-up:
	goose -dir $(MIGRATION_FOLDER) postgres "$(POSTGRES_SETUP_MAIN)" up

test-migration-down:
	goose -dir $(MIGRATION_FOLDER) postgres "$(POSTGRES_SETUP_TEST)" down

migration-down:
	goose -dir $(MIGRATION_FOLDER) postgres "$(POSTGRES_SETUP_MAIN)" down

test-docker-up:
	docker-compose up test_db -d

integration-tests:
	go test -tags=integration ./tests/handlers_test/... -v

unit-tests:
	go test ./internal/... -v -cover

run-migrations: test-migration-up

clean-test-db:
	echo "DELETE FROM users WHERE email LIKE '%test%';" | psql "$(POSTGRES_SETUP_TEST)"
	echo "TRUNCATE TABLE test_data CASCADE;" | psql "$(POSTGRES_SETUP_TEST)"

proto-generate:
	protoc --proto_path=$(PROTO_PATH) --go_out=$(PROTO_DEST) --go-grpc_out=$(PROTO_DEST) $(PROTO_SRC)