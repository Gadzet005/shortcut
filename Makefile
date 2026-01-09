.PHONY: build run test test-coverage lint fmt vet clean install-tools mock

APP_NAME=shortcut
BUILD_DIR=bin
MAIN_PATH=./shortcut/cmd/server
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html
GOBIN=$(shell go env GOPATH)/bin

install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/vektra/mockery/v3@latest

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

dev:
	go run $(MAIN_PATH)/main.go

prod:
	go run $(MAIN_PATH)/main.go -c "./shortcut/configs/base.yaml,./shortcut/configs/prod.yaml"

test:
	go test -v -race -short ./...

test-e2e: podman-up
	go test -v -race ./tests/e2e/...
	$(MAKE) podman-down

test-coverage:
	go test -v -race -short -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	go tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)

lint:
	$(GOBIN)/golangci-lint run ./...

fmt:
	go fmt ./...
	gofmt -s -w .

vet:
	go vet ./...

mock:
	$(GOBIN)/mockery

tidy:
	go mod tidy
	go mod verify

clean:
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)

deps:
	go mod download

podman-build:
	podman-compose -f tests/infra/docker-compose.yaml build

podman-up:
	podman-compose -f tests/infra/docker-compose.yaml up -d

podman-down:
	podman-compose -f tests/infra/docker-compose.yaml down -v

podman-logs:
	podman-compose -f tests/infra/docker-compose.yaml logs -f shortcut

podman-logs-shortcut:
	podman-compose -f tests/infra/docker-compose.yaml logs -f shortcut

podman-logs-mock-8081:
	podman-compose -f tests/infra/docker-compose.yaml logs -f mock-service-8081

podman-logs-mock-8082:
	podman-compose -f tests/infra/docker-compose.yaml logs -f mock-service-8082

podman-logs-mock-8083:
	podman-compose -f tests/infra/docker-compose.yaml logs -f mock-service-8083

podman-logs-all-mocks:
	podman-compose -f tests/infra/docker-compose.yaml logs -f mock-service-8081 mock-service-8082 mock-service-8083

podman-clean:
	podman-compose -f tests/infra/docker-compose.yaml down -v
	podman system prune -af

check: fmt vet lint test
