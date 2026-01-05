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
	go test -v -race ./...

test-coverage:
	go test -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
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

check: fmt vet lint test
