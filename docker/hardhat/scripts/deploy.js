const { ethers } = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
    console.log("Starting deployment...");
    const [owner, seller, buyer] = await ethers.getSigners();

    console.log("--- DEPLOYING CONTRACTS ---");

    const NFT = await ethers.getContractFactory("NFT");
    const nft = await NFT.deploy();
    await nft.waitForDeployment();
    const nftAddress = await nft.getAddress();
    console.log(`NFT deployed to: ${nftAddress}`);

    const Marketplace = await ethers.getContractFactory("Marketplace");
    const marketplace = await Marketplace.deploy();
    await marketplace.waitForDeployment();
    const marketAddress = await marketplace.getAddress();
    console.log(`Marketplace deployed to: ${marketAddress}`);

    // Create .env content
    const envContent = `CHAIN_ID=1337
RPC_URL=http://localhost:8545
NFT_ADDRESS=${nftAddress}
MARKET_ADDRESS=${marketAddress}
OWNER_PRIVATE_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80
SELLER_PRIVATE_KEY=0x59c6995e998f97a5a0044966f0945389dc9e86dae88c7a8412f4603b6b78690d
BUYER_PRIVATE_KEY=0x5de4111afa1a4b94908f83103eb1f1706367c2e68ca870fc3fb9a804cdab365a
`;

    console.log("ENV_CONTENT_START");
    console.log(envContent);
    console.log("ENV_CONTENT_END");

    // Export ABIs for Go to read
    const artifactsDir = "/data/NFT/internal/abi";
    if (!fs.existsSync(artifactsDir)) {
        fs.mkdirSync(artifactsDir, { recursive: true });
    }

    const { artifacts } = require("hardhat");
    const nftArtifact = await artifacts.readArtifact("NFT");
    fs.writeFileSync(path.join(artifactsDir, "NFT.json"), JSON.stringify(nftArtifact.abi));

    const marketArtifact = await artifacts.readArtifact("Marketplace");
    fs.writeFileSync(path.join(artifactsDir, "Marketplace.json"), JSON.stringify(marketArtifact.abi));

    console.log(`Exported ABIs to ${artifactsDir}`);
}

main().catch((error) => {
    console.error(error);
    process.exitCode = 1;
});
