package eth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/user/nft-marketplace/internal/config"
)

type Client struct {
	cfg config.EthConfig
	rpc *ethclient.Client

	nftABI    abi.ABI
	marketABI abi.ABI

	nftAddr    common.Address
	marketAddr common.Address
}

func NewClient(cfg config.EthConfig) (*Client, error) {
	rpc, err := ethclient.Dial(cfg.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("dial rpc: %w", err)
	}

	// Read ABIs from filesystem
	nftABIBytes, err := os.ReadFile("internal/abi/NFT.abi.json")
	if err != nil {
		return nil, fmt.Errorf("read nft abi: %w", err)
	}
	nftABI, err := abi.JSON(strings.NewReader(string(nftABIBytes)))
	if err != nil {
		return nil, fmt.Errorf("parse nft abi: %w", err)
	}

	marketABIBytes, err := os.ReadFile("internal/abi/Marketplace.json")
	if err != nil {
		return nil, fmt.Errorf("read market abi: %w", err)
	}
	marketABI, err := abi.JSON(strings.NewReader(string(marketABIBytes)))
	if err != nil {
		return nil, fmt.Errorf("parse market abi: %w", err)
	}

	// Detect ChainID from RPC
	detectedChainID, err := rpc.NetworkID(context.Background())
	if err != nil {
		log.Printf("Warning: Could not detect ChainID from RPC: %v", err)
	} else {
		log.Printf("Connected to RPC: %s, Detected ChainID: %v, Configured ChainID: %v", cfg.RPCURL, detectedChainID, cfg.ChainID)
	}

	return &Client{
		cfg:        cfg,
		rpc:        rpc,
		nftABI:     nftABI,
		marketABI:  marketABI,
		nftAddr:    common.HexToAddress(cfg.NFTAddress),
		marketAddr: common.HexToAddress(cfg.MarketAddress),
	}, nil
}


func (c *Client) GetNFTAddress() string {
	return c.nftAddr.Hex()
}

func (c *Client) AddressFromPriv(privHex string) (string, error) {
	privHex = strings.TrimPrefix(privHex, "0x")
	if privHex == "" {
		return "", errors.New("empty private key")
	}
	pk, err := crypto.HexToECDSA(privHex)
	if err != nil {
		return "", err
	}
	addr := crypto.PubkeyToAddress(pk.PublicKey)
	return addr.Hex(), nil
}

func (c *Client) Mint(to string, tokenURI string) (string, string, error) {
	auth, err := c.txOpts(c.cfg.OwnerPrivateKey)
	if err != nil {
		return "", "", err
	}

	nft := bind.NewBoundContract(c.nftAddr, c.nftABI, c.rpc, c.rpc, c.rpc)
	tx, err := nft.Transact(auth, "mint", common.HexToAddress(to), tokenURI)
	if err != nil {
		return "", "", fmt.Errorf("mint tx: %w", err)
	}

	if err := c.waitMined(tx.Hash()); err != nil {
		return tx.Hash().Hex(), "", err
	}

	// Get tokenId (nextTokenId - 1)
	var next *big.Int
	// output must be a pointer to a slice of empty interfaces, or variadic
	// But BoundContract.Call takes (opts, *results, method, params...)
	// The results argument must be a pointer data type that matches the output.
	// However, for single return values, go-ethereum binding requires []interface{} usually?
	// Actually, NewBoundContract Call signature: func (c *BoundContract) Call(opts *CallOpts, results *[]interface{}, method string, params ...interface{}) error
	// Wait, standard bind.BoundContract Call signature is: Call(opts *CallOpts, result *[]interface{}, method string, params ...interface{})
	// Let's use the generated bindings style or handle []interface{}.

	results := []interface{}{&next}
	if err := nft.Call(&bind.CallOpts{}, &results, "nextTokenId"); err != nil {
		return tx.Hash().Hex(), "", fmt.Errorf("call nextTokenId: %w", err)
	}
	
	// nextTokenId is the *next* one, so minted was next - 1
	tokenId := new(big.Int).Sub(next, big.NewInt(1))

	return tx.Hash().Hex(), tokenId.String(), nil
}

func (c *Client) Approve(tokenId string) (string, error) {
	auth, err := c.txOpts(c.cfg.SellerPrivateKey)
	if err != nil {
		return "", err
	}

	tid, ok := new(big.Int).SetString(tokenId, 10)
	if !ok {
		return "", errors.New("invalid token id")
	}

	nft := bind.NewBoundContract(c.nftAddr, c.nftABI, c.rpc, c.rpc, c.rpc)
	tx, err := nft.Transact(auth, "approve", c.marketAddr, tid)
	if err != nil {
		return "", fmt.Errorf("approve tx: %w", err)
	}
	if err := c.waitMined(tx.Hash()); err != nil {
		return tx.Hash().Hex(), err
	}
	return tx.Hash().Hex(), nil
}

func (c *Client) List(tokenId, priceWei string) (string, error) {
	auth, err := c.txOpts(c.cfg.SellerPrivateKey)
	if err != nil {
		return "", err
	}

	tid, ok := new(big.Int).SetString(tokenId, 10)
	if !ok {
		return "", errors.New("invalid token id")
	}
	price, ok := new(big.Int).SetString(priceWei, 10)
	if !ok {
		return "", errors.New("invalid price")
	}

	market := bind.NewBoundContract(c.marketAddr, c.marketABI, c.rpc, c.rpc, c.rpc)
	tx, err := market.Transact(auth, "list", c.nftAddr, tid, price)
	if err != nil {
		return "", fmt.Errorf("list tx: %w", err)
	}
	if err := c.waitMined(tx.Hash()); err != nil {
		return tx.Hash().Hex(), err
	}
	return tx.Hash().Hex(), nil
}

func (c *Client) Buy(tokenId, priceWei string) (string, error) {
	auth, err := c.txOpts(c.cfg.BuyerPrivateKey)
	if err != nil {
		return "", err
	}

	tid, ok := new(big.Int).SetString(tokenId, 10)
	if !ok {
		return "", errors.New("invalid token id")
	}
	price, ok := new(big.Int).SetString(priceWei, 10)
	if !ok {
		return "", errors.New("invalid price")
	}
	
	// Value must match price
	auth.Value = price

	market := bind.NewBoundContract(c.marketAddr, c.marketABI, c.rpc, c.rpc, c.rpc)
	tx, err := market.Transact(auth, "buy", c.nftAddr, tid)
	if err != nil {
		return "", fmt.Errorf("buy tx: %w", err)
	}
	if err := c.waitMined(tx.Hash()); err != nil {
		return tx.Hash().Hex(), err
	}
	return tx.Hash().Hex(), nil
}

func (c *Client) OwnerOf(tokenId string) (string, error) {
	tid, ok := new(big.Int).SetString(tokenId, 10)
	if !ok {
		return "", errors.New("invalid token id")
	}

	nft := bind.NewBoundContract(c.nftAddr, c.nftABI, c.rpc, c.rpc, c.rpc)
	var out common.Address
	results := []interface{}{&out}
	if err := nft.Call(&bind.CallOpts{}, &results, "ownerOf", tid); err != nil {
		return "", err
	}
	return out.Hex(), nil
}

func (c *Client) TokenURI(tokenId string) (string, error) {
	tid, ok := new(big.Int).SetString(tokenId, 10)
	if !ok {
		return "", errors.New("invalid token id")
	}

	nft := bind.NewBoundContract(c.nftAddr, c.nftABI, c.rpc, c.rpc, c.rpc)
	var out string
	results := []interface{}{&out}
	// Note: string returns are tricky in bound contracts sometimes, but basic types usually ok with binding
	if err := nft.Call(&bind.CallOpts{}, &results, "tokenURI", tid); err != nil {
		return "", err
	}
	return out, nil
}

func (c *Client) txOpts(privHex string) (*bind.TransactOpts, error) {
	privHex = strings.TrimPrefix(privHex, "0x")
	if privHex == "" {
		return nil, errors.New("missing private key")
	}
	pk, err := crypto.HexToECDSA(privHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}

	// Get Nonce
	addr := crypto.PubkeyToAddress(pk.PublicKey)
	nonce, err := c.rpc.PendingNonceAt(context.Background(), addr)
	if err != nil {
		return nil, fmt.Errorf("nonce: %w", err)
	}

	// Suggest Gas Price
	gasPrice, err := c.rpc.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("gas price: %w", err)
	}


	log.Printf("Creating transaction with ChainID: %d for address: %s", c.cfg.ChainID, addr.Hex())
	log.Printf("Using ChainID as big.Int: %s", big.NewInt(c.cfg.ChainID).String())
	
	auth, err := bind.NewKeyedTransactorWithChainID(pk, big.NewInt(c.cfg.ChainID))
	if err != nil {
		return nil, fmt.Errorf("transactor: %w", err)
	}
	
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000) // Increase Gas Limit for safety
	auth.GasPrice = gasPrice
	auth.Context = context.Background()

	return auth, nil
}

func (c *Client) waitMined(txHash common.Hash) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
			receipt, err := c.rpc.TransactionReceipt(ctx, txHash)
			if err != nil {
				// un-mined tx might return "not found" error, depending on backend
				if err == ethereum.NotFound {
					continue
				}
				// some clients return nil receipt while pending?
				continue
			}
			if receipt != nil {
				if receipt.Status == 0 {
					return fmt.Errorf("transaction failed (status 0)")
				}
				return nil
			}
		}
	}
}

func (c *Client) Burn(tokenId, ownerPrivateKey string) (string, error) {
	auth, err := c.txOpts(ownerPrivateKey)
	if err != nil {
		return "", err
	}

	tid, ok := new(big.Int).SetString(tokenId, 10)
	if !ok {
		return "", errors.New("invalid token id")
	}

	nft := bind.NewBoundContract(c.nftAddr, c.nftABI, c.rpc, c.rpc, c.rpc)
	tx, err := nft.Transact(auth, "burn", tid)
	if err != nil {
		return "", fmt.Errorf("burn tx: %w", err)
	}
	if err := c.waitMined(tx.Hash()); err != nil {
		return tx.Hash().Hex(), err
	}
	return tx.Hash().Hex(), nil
}

func (c *Client) Delist(tokenId string) (string, error) {
	auth, err := c.txOpts(c.cfg.SellerPrivateKey)
	if err != nil {
		return "", err
	}

	tid, ok := new(big.Int).SetString(tokenId, 10)
	if !ok {
		return "", errors.New("invalid token id")
	}

	market := bind.NewBoundContract(c.marketAddr, c.marketABI, c.rpc, c.rpc, c.rpc)
	tx, err := market.Transact(auth, "delist", c.nftAddr, tid)
	if err != nil {
		return "", fmt.Errorf("delist tx: %w", err)
	}
	if err := c.waitMined(tx.Hash()); err != nil {
		return tx.Hash().Hex(), err
	}
	return tx.Hash().Hex(), nil
}

func (c *Client) GetListing(tokenId string) (price string, seller string, active bool, err error) {
	tid, ok := new(big.Int).SetString(tokenId, 10)
	if !ok {
		return "", "", false, errors.New("invalid token id")
	}

	market := bind.NewBoundContract(c.marketAddr, c.marketABI, c.rpc, c.rpc, c.rpc)
	
	// The struct returns (price, seller, active)
	var priceOut *big.Int
	var sellerOut common.Address
	var activeOut bool
	
	results := []interface{}{&priceOut, &sellerOut, &activeOut}
	if err := market.Call(&bind.CallOpts{}, &results, "getListing", c.nftAddr, tid); err != nil {
		return "", "", false, err
	}
	
	return priceOut.String(), sellerOut.Hex(), activeOut, nil
}
