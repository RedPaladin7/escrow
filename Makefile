.PHONY: help build run test clean deploy compile install

# Default target
help:
	@echo "PeerPoker Makefile"
	@echo ""
	@echo "Available targets:"
	@echo "  install        - Install all dependencies (Go + Node.js)"
	@echo "  compile        - Compile smart contracts"
	@echo "  build          - Build Go binary"
	@echo "  run            - Run the server"
	@echo "  test           - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo "  node           - Start Hardhat node"
	@echo "  deploy         - Deploy contracts to localhost"
	@echo "  deploy-sepolia - Deploy contracts to Sepolia"
	@echo "  all            - Install, compile, and build"

# Install dependencies
install:
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy
	@echo "Installing Node.js dependencies..."
	npm install
	@echo "✓ Dependencies installed"

# Compile smart contracts
compile:
	@echo "Compiling smart contracts..."
	npx hardhat compile
	@echo "✓ Contracts compiled"

# Build Go binary
build:
	@echo "Building Go binary..."
	go build -o bin/peerpoker cmd/server/main.go
	@echo "✓ Binary built: bin/peerpoker"

# Run the server
run:
	@echo "Starting PeerPoker server..."
	go run cmd/server/main.go

# Run tests
test:
	@echo "Running Go tests..."
	go test ./... -v
	@echo "Running Hardhat tests..."
	npx hardhat test

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf artifacts/
	rm -rf cache/
	rm -rf node_modules/
	go clean
	@echo "✓ Cleaned"

# Start Hardhat node
node:
	@echo "Starting Hardhat node..."
	npx hardhat node

# Deploy to localhost
deploy:
	@echo "Deploying to localhost..."
	npx hardhat run scripts/deploy.js --network localhost

# Deploy to Sepolia
deploy-sepolia:
	@echo "Deploying to Sepolia..."
	npx hardhat run scripts/deploy.js --network sepolia

# All-in-one
all: install compile build
	@echo "✓ Setup complete!"
