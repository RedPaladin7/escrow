.PHONY: all build run clean test fmt lint deps help

# Variables
BINARY_NAME=poker-server
MAIN_PATH=cmd/server/main.go
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

all: deps fmt build

## build: Build the binary
build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## run: Run the server
run:
	@echo "Running server..."
	$(GOCMD) run $(MAIN_PATH)

## run-node1: Run first node on default ports
run-node1:
	@echo "Starting Node 1 (WS:3000, API:8080)..."
	$(GOCMD) run $(MAIN_PATH) -ws-port=3000 -api-port=8080

## run-node2: Run second node connecting to first
run-node2:
	@echo "Starting Node 2 (WS:3001, API:8081)..."
	$(GOCMD) run $(MAIN_PATH) -ws-port=3001 -api-port=8081 -connect=ws://localhost:3000/p2p

## run-node3: Run third node connecting to first
run-node3:
	@echo "Starting Node 3 (WS:3002, API:8082)..."
	$(GOCMD) run $(MAIN_PATH) -ws-port=3002 -api-port=8082 -connect=ws://localhost:3000/p2p

## clean: Clean build files
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf snapshots/
	rm -rf backups/
	@echo "Clean complete"

## test: Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

## lint: Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	golangci-lint run

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## install: Install the binary
install: build
	@echo "Installing..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

## dev: Run in development mode with hot reload (requires air)
dev:
	@echo "Running in development mode..."
	air

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t poker-server:latest .

## docker-run: Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 3000:3000 -p 8080:8080 poker-server:latest

## help: Display this help
help:
	@echo "Available targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'
