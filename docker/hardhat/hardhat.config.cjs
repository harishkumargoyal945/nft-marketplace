/** @type import('hardhat/config').HardhatUserConfig */
require("@nomicfoundation/hardhat-toolbox");
module.exports = {
    solidity: "0.8.20",
    networks: {
        hardhat: {
            chainId: 1337,
            mining: {
                auto: true,
                interval: 0
            }
        },
        localhost: {
            url: "http://127.0.0.1:8500",
            chainId: 1337
        }
    }
};
