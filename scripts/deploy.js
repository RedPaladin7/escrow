// scripts/deploy.js
const hre = require("hardhat");

async function main() {
  console.log("Deploying PeerPoker contracts...");

  // Get deployer
  const [deployer] = await hre.ethers.getSigners();
  console.log("Deploying contracts with account:", deployer.address);

  const balance = await deployer.getBalance();
  console.log("Account balance:", hre.ethers.utils.formatEther(balance), "ETH");

  // Deploy PlayerRegistry
  console.log("\nDeploying PlayerRegistry...");
  const PlayerRegistry = await hre.ethers.getContractFactory("PlayerRegistry");
  const playerRegistry = await PlayerRegistry.deploy();
  await playerRegistry.deployed();
  console.log("PlayerRegistry deployed to:", playerRegistry.address);

  // Deploy PotManager (placeholder address for PokerTable)
  console.log("\nDeploying PotManager...");
  const PotManager = await hre.ethers.getContractFactory("PotManager");
  const potManager = await PotManager.deploy(deployer.address); // Temporary, will update
  await potManager.deployed();
  console.log("PotManager deployed to:", potManager.address);

  // Deploy PokerTable
  console.log("\nDeploying PokerTable...");
  const PokerTable = await hre.ethers.getContractFactory("PokerTable");
  const pokerTable = await PokerTable.deploy(
    potManager.address,
    playerRegistry.address,
    deployer.address // Platform fee address
  );
  await pokerTable.deployed();
  console.log("PokerTable deployed to:", pokerTable.address);

  // Update PotManager with PokerTable address
  console.log("\nUpdating PotManager with PokerTable address...");
  // Note: You'll need to add a function to update the pokerTable address in PotManager
  // or redeploy PotManager with the correct address

  // Deploy DisputeResolver
  console.log("\nDeploying DisputeResolver...");
  const DisputeResolver = await hre.ethers.getContractFactory("DisputeResolver");
  const disputeResolver = await DisputeResolver.deploy();
  await disputeResolver.deployed();
  console.log("DisputeResolver deployed to:", disputeResolver.address);

  // Set PokerTable in PlayerRegistry
  console.log("\nSetting PokerTable in PlayerRegistry...");
  await playerRegistry.setPokerTable(pokerTable.address);
  console.log("PokerTable set in PlayerRegistry");

  // Print summary
  console.log("\n=== Deployment Summary ===");
  console.log("PokerTable:", pokerTable.address);
  console.log("PotManager:", potManager.address);
  console.log("PlayerRegistry:", playerRegistry.address);
  console.log("DisputeResolver:", disputeResolver.address);
  console.log("\nAdd these addresses to your .env file:");
  console.log(`CONTRACT_POKER_TABLE=${pokerTable.address}`);
  console.log(`CONTRACT_POT_MANAGER=${potManager.address}`);
  console.log(`CONTRACT_PLAYER_REGISTRY=${playerRegistry.address}`);
  console.log(`CONTRACT_DISPUTE_RESOLVER=${disputeResolver.address}`);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
