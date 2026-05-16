.PHONY: build run test test/unit test/e2e test-coverage lint fmt vet clean install-tools mock gen-dashboards

APP_NAME=shortcut
BUILD_DIR=bin
MAIN_PATH=./cmd/shortcut
COVERAGE_FILE=coverage.out
COVERAGE_HTML=coverage.html
GOBIN=$(shell go env GOPATH)/bin
DASHBOARDS_CONFIGS_DIR ?= demo/configs
DASHBOARDS_OUT_DIR ?= k8s/dashboards

install:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/vektra/mockery/v3@latest

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

run:
	go run $(MAIN_PATH)

run/mock:
	CONFIGS_DIR=./tests/mock-service/configs go run ./tests/mock-service

test: install mock
	go test -v ./...

test/unit: install mock
	go test -race -count=1 $$(go list ./... | grep -v /tests/e2e)

test/e2e: install mock
	go test -v -count=1 ./tests/e2e/...

test-coverage: install mock
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

gen-dashboards:
	@python3 scripts/gen_dashboards.py --configs-dir $(DASHBOARDS_CONFIGS_DIR) --out-dir $(DASHBOARDS_OUT_DIR)
