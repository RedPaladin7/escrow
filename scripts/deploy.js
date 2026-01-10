const hre = require("hardhat");

async function main() {
  console.log("Starting deployment...\n");

  // Get deployer account
  const [deployer] = await hre.ethers.getSigners();
  console.log("Deploying contracts with account:", deployer.address);
  console.log("Account balance:", (await deployer.provider.getBalance(deployer.address)).toString(), "wei\n");

  // Deploy PlayerRegistry
  console.log("Deploying PlayerRegistry...");
  const PlayerRegistry = await hre.ethers.getContractFactory("PlayerRegistry");
  const playerRegistry = await PlayerRegistry.deploy();
  await playerRegistry.waitForDeployment();
  const playerRegistryAddress = await playerRegistry.getAddress();
  console.log("✓ PlayerRegistry deployed to:", playerRegistryAddress);

  // Deploy PotManager (with temporary poker table address)
  console.log("\nDeploying PotManager...");
  const PotManager = await hre.ethers.getContractFactory("PotManager");
  const potManager = await PotManager.deploy(deployer.address);
  await potManager.waitForDeployment();
  const potManagerAddress = await potManager.getAddress();
  console.log("✓ PotManager deployed to:", potManagerAddress);

  // Deploy PokerTable
  console.log("\nDeploying PokerTable...");
  const PokerTable = await hre.ethers.getContractFactory("PokerTable");
  const pokerTable = await PokerTable.deploy(
    potManagerAddress,
    playerRegistryAddress,
    deployer.address // Platform fee address
  );
  await pokerTable.waitForDeployment();
  const pokerTableAddress = await pokerTable.getAddress();
  console.log("✓ PokerTable deployed to:", pokerTableAddress);

  // Deploy DisputeResolver
  console.log("\nDeploying DisputeResolver...");
  const DisputeResolver = await hre.ethers.getContractFactory("DisputeResolver");
  const disputeResolver = await DisputeResolver.deploy();
  await disputeResolver.waitForDeployment();
  const disputeResolverAddress = await disputeResolver.getAddress();
  console.log("✓ DisputeResolver deployed to:", disputeResolverAddress);

  // Set PokerTable in PlayerRegistry
  console.log("\nSetting PokerTable in PlayerRegistry...");
  const tx = await playerRegistry.setPokerTable(pokerTableAddress);
  await tx.wait();
  console.log("✓ PokerTable set in PlayerRegistry");

  // Print deployment summary
  console.log("\n" + "=".repeat(70));
  console.log("DEPLOYMENT SUMMARY");
  console.log("=".repeat(70));
  console.log("Network:", hre.network.name);
  console.log("Chain ID:", (await hre.ethers.provider.getNetwork()).chainId.toString());
  console.log("\nContract Addresses:");
  console.log("-------------------");
  console.log("PokerTable:       ", pokerTableAddress);
  console.log("PotManager:       ", potManagerAddress);
  console.log("PlayerRegistry:   ", playerRegistryAddress);
  console.log("DisputeResolver:  ", disputeResolverAddress);
  console.log("\n" + "=".repeat(70));
  console.log("Add these to your .env file:");
  console.log("=".repeat(70));
  console.log(`CONTRACT_POKER_TABLE=${pokerTableAddress}`);
  console.log(`CONTRACT_POT_MANAGER=${potManagerAddress}`);
  console.log(`CONTRACT_PLAYER_REGISTRY=${playerRegistryAddress}`);
  console.log(`CONTRACT_DISPUTE_RESOLVER=${disputeResolverAddress}`);
  console.log("=".repeat(70) + "\n");

  // Verify deployment
  console.log("Verifying deployment...");
  const code = await hre.ethers.provider.getCode(pokerTableAddress);
  if (code === "0x") {
    console.log("❌ Deployment verification failed!");
  } else {
    console.log("✓ Deployment verified successfully!");
  }

  console.log("\n✅ All contracts deployed successfully!\n");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
