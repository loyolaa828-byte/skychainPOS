require("@nomicfoundation/hardhat-ethers");

module.exports = {
  solidity: "0.8.24",
  networks: {
    hardhat: { chainId: 8888 },
    localhost: { url: "http://127.0.0.1:8545", chainId: 8888 }
  }
};
