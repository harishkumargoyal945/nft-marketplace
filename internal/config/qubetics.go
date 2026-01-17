package config

import (
	"strconv"
)

type EthConfig struct {
	RPCURL          string
	NFTAddress      string
	MarketAddress   string
	OwnerPrivateKey string
	SellerPrivateKey string
	BuyerPrivateKey  string
	ChainID         int64
}

func LoadEthConfig() *EthConfig {
	chainID, _ := strconv.ParseInt(getEnv("CHAIN_ID", "1337"), 10, 64)
	return &EthConfig{
		RPCURL:          getEnv("RPC_URL", "http://localhost:8500"),
		NFTAddress:      getEnv("NFT_ADDRESS", ""),
		MarketAddress:   getEnv("MARKET_ADDRESS", ""),
		OwnerPrivateKey: getEnv("OWNER_PRIVATE_KEY", ""),
		SellerPrivateKey: getEnv("SELLER_PRIVATE_KEY", ""),
		BuyerPrivateKey:  getEnv("BUYER_PRIVATE_KEY", ""),
		ChainID:         chainID,
	}
}