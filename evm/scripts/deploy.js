const hre = require("hardhat");
const fs = require("fs");

async function main() {
  const WSKY = await hre.ethers.getContractFactory("WSKY");
  const wsky = await WSKY.deploy();
  await wsky.waitForDeployment();

  const SkyUSD = await hre.ethers.getContractFactory("SkyUSD");
  const usdc = await SkyUSD.deploy(1_000_000n * 10n ** 6n);
  await usdc.waitForDeployment();

  const SkySwapAMM = await hre.ethers.getContractFactory("SkySwapAMM");
  const dex = await SkySwapAMM.deploy(await wsky.getAddress(), await usdc.getAddress());
  await dex.waitForDeployment();

  const SkyStakingVault = await hre.ethers.getContractFactory("SkyStakingVault");
  const vault = await SkyStakingVault.deploy(await wsky.getAddress());
  await vault.waitForDeployment();

  await (await wsky.deposit({ value: hre.ethers.parseEther("500") })).wait();
  await (await wsky.approve(await dex.getAddress(), hre.ethers.parseEther("100"))).wait();
  await (await usdc.approve(await dex.getAddress(), 250n * 10n ** 6n)).wait();
  await (await dex.addLiquidity(hre.ethers.parseEther("100"), 250n * 10n ** 6n)).wait();

  const addrs = {
    wsky: await wsky.getAddress(),
    usdc: await usdc.getAddress(),
    dex: await dex.getAddress(),
    vault: await vault.getAddress()
  };
  fs.writeFileSync("/workspaces/skychainPOS/explorer-lite/contracts.json", JSON.stringify(addrs, null, 2));
  console.log("Deployed contracts:", addrs);
}

main().catch((err) => { console.error(err); process.exit(1); });
