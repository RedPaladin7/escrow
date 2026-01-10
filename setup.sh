#!/bin/bash

echo "=================================================="
echo "PeerPoker Setup Script (Hardhat)"
echo "=================================================="

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

# Check prerequisites
echo -e "\n${BLUE}Checking prerequisites...${NC}"

if ! command -v go &> /dev/null; then
    echo -e "${RED}Go is not installed. Please install Go first.${NC}"
    exit 1
fi

if ! command -v node &> /dev/null; then
    echo -e "${RED}Node.js is not installed. Please install Node.js first.${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Prerequisites OK${NC}"

# Step 1: Create directories
echo -e "\n${BLUE}Step 1: Creating directory structure...${NC}"
mkdir -p cmd/server
mkdir -p internal/{api,blockchain,config,crypto,deck,game,persistence,protocol,server,transport}
mkdir -p contracts/interfaces
mkdir -p scripts test snapshots backups
echo -e "${GREEN}✓ Directories created${NC}"

# Step 2: Initialize Go module
echo -e "\n${BLUE}Step 2: Initializing Go module...${NC}"
if [ ! -f go.mod ]; then
    go mod init github.com/RedPaladin7/peerpoker
    echo -e "${GREEN}✓ Go module initialized${NC}"
else
    echo -e "${GREEN}✓ Go module already exists${NC}"
fi

# Step 3: Initialize Node.js
echo -e "\n${BLUE}Step 3: Initializing Node.js...${NC}"
if [ ! -f package.json ]; then
    npm init -y
    echo -e "${GREEN}✓ Node.js initialized${NC}"
else
    echo -e "${GREEN}✓ package.json already exists${NC}"
fi

# Step 4: Install Node dependencies
echo -e "\n${BLUE}Step 4: Installing Hardhat dependencies...${NC}"
npm install --save-dev hardhat @nomicfoundation/hardhat-toolbox
npm install dotenv
echo -e "${GREEN}✓ Hardhat installed${NC}"

# Step 5: Setup .env
echo -e "\n${BLUE}Step 5: Setting up environment file...${NC}"
if [ ! -f .env ]; then
    cp .env.example .env
    echo -e "${GREEN}✓ .env file created${NC}"
else
    echo -e "${GREEN}✓ .env file already exists${NC}"
fi

# Step 6: Install Go dependencies
echo -e "\n${BLUE}Step 6: Installing Go dependencies...${NC}"
go mod download
go mod tidy
echo -e "${GREEN}✓ Go dependencies installed${NC}"

# Step 7: Compile contracts
echo -e "\n${BLUE}Step 7: Compiling smart contracts...${NC}"
npx hardhat compile
echo -e "${GREEN}✓ Contracts compiled${NC}"

echo -e "\n${GREEN}=================================================="
echo "Setup Complete!"
echo "==================================================${NC}"
echo -e "\nNext steps:"
echo "1. Edit .env file with your configuration"
echo "2. Terminal 1: ${BLUE}npx hardhat node${NC}"
echo "3. Terminal 2: ${BLUE}npx hardhat run scripts/deploy.js --network localhost${NC}"
echo "4. Copy contract addresses to .env"
echo "5. Terminal 3: ${BLUE}go run cmd/server/main.go${NC}"
