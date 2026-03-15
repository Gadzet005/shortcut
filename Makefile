.PHONY: build run test test-coverage lint fmt vet clean install-tools mock

APP_NAME=shortcut
BUILD_DIR=bin
MAIN_PATH=./cmd/shortcut
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html
GOBIN=$(shell go env GOPATH)/bin

install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/vektra/mockery/v3@latest

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)/main.go

test:
	go test -v -race -short ./...

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

clean:
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)

check: fmt vet lint test
