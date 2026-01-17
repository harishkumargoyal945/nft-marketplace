package service

import (
	"errors"
	"fmt"
	"log"

	"github.com/user/nft-marketplace/internal/core"
	"github.com/user/nft-marketplace/internal/platform/eth"
	"github.com/user/nft-marketplace/internal/repository"
)

type MarketplaceService struct {
	repo *repository.Repository
	eth  *eth.Client
}

func NewMarketplaceService(repo *repository.Repository, ethClient *eth.Client) *MarketplaceService {
	return &MarketplaceService{repo: repo, eth: ethClient}
}

func (s *MarketplaceService) Health() error {
	return s.repo.Ping()
}

func (s *MarketplaceService) CreateUser(wallet string, name string) (*core.User, error) {
	user := &core.User{
		WalletAddress: wallet,
		Name:          name,
	}
	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *MarketplaceService) GetUser(id uint) (*core.User, error) {
	return s.repo.GetUserByID(id)
}

func (s *MarketplaceService) CreateCollection(creatorID uint, name string, symbol string) (*core.Collection, error) {
	collection := &core.Collection{
		CreatorUserID: creatorID,
		Name:          name,
		Symbol:        symbol,
	}
	if err := s.repo.CreateCollection(collection); err != nil {
		return nil, err
	}
	return collection, nil
}

func (s *MarketplaceService) ListCollections() ([]core.Collection, error) {
	return s.repo.ListCollections()
}

func (s *MarketplaceService) MintNFT(ownerID uint, name, symbol, desc, imageURL, collectionName string) (*core.NFT, error) {
	user, err := s.repo.GetUserByID(ownerID)
	if err != nil {
		return nil, err
	}

	// Find or Create Collection
	var collection *core.Collection
	if collectionName != "" {
		collection, err = s.repo.FindCollectionByName(ownerID, collectionName)
		if err != nil {
			// Not found, create it
			newCol := &core.Collection{
				CreatorUserID: ownerID,
				Name:          collectionName,
				Symbol:        "NFT", // Default symbol, logic could be better
			}
			if err := s.repo.CreateCollection(newCol); err != nil {
				return nil, fmt.Errorf("failed to auto-create collection: %w", err)
			}
			collection = newCol
		}
	} else {
		// Use default if no name provided ? Or fail?
		// User scenario implies we use name if given. If not given, we can fall back to "Default Collection" or fail.
		// Let's fallback to finding *any* collection or "Default Collection" to be safe.
		collection, err = s.repo.FindCollectionByOwner(ownerID)
		if err != nil {
			// Auto create default
			defaultName := "Default Collection"
			collection, err = s.repo.FindCollectionByName(ownerID, defaultName)
			if err != nil {
				newCol := &core.Collection{
					CreatorUserID: ownerID,
					Name:          defaultName,
					Symbol:        "DEF",
				}
				if err := s.repo.CreateCollection(newCol); err != nil {
					return nil, fmt.Errorf("failed to create default collection: %w", err)
				}
				collection = newCol
			}
		}
	}

	// 1. Mint on blockchain
	txHash, tokenID, err := s.eth.Mint(user.WalletAddress, imageURL)
	if err != nil {
		return nil, fmt.Errorf("blockchain mint failure: %w", err)
	}
	log.Printf("Minted NFT: TokenID=%s, TxHandle=%s", tokenID, txHash)

	// 2. Register in DB
	nft := &core.NFT{
		TokenID:         tokenID,
		ContractAddress: s.eth.GetNFTAddress(),
		Chain:           "Qubetics",
		CollectionID:    collection.ID,
		OwnerUserID:     ownerID,
		MetadataURL:     imageURL,
	}
	if err := s.repo.CreateNFT(nft); err != nil {
		return nil, err
	}
	return nft, nil
}

func (s *MarketplaceService) RegisterNFT(tokenID, contract, chain string, collectionID, ownerID uint, metadataURL string) (*core.NFT, error) {
	nft := &core.NFT{
		TokenID:         tokenID,
		ContractAddress: contract,
		Chain:           chain,
		CollectionID:    collectionID,
		OwnerUserID:     ownerID,
		MetadataURL:     metadataURL,
	}
	if err := s.repo.CreateNFT(nft); err != nil {
		return nil, err
	}
	return nft, nil
}

func (s *MarketplaceService) ListNFTs(ownerID, collectionID uint, chain string) ([]core.NFT, error) {
	return s.repo.ListNFTs(ownerID, collectionID, chain)
}

func (s *MarketplaceService) CreateListing(nftID, sellerID uint, priceWei, currency string) (*core.Listing, error) {
	// Check if seller owns NFT
	nft, err := s.repo.GetNFTByID(nftID)
	if err != nil {
		return nil, fmt.Errorf("nft not found: %w", err)
	}
	if nft.OwnerUserID != sellerID {
		return nil, errors.New("seller does not own this nft")
	}

	listing := &core.Listing{
		NFTID:        nftID,
		SellerUserID: sellerID,
		PriceWei:     priceWei,
		Currency:     currency,
		Status:       core.ListingActive,
	}
	if err := s.repo.CreateListing(listing); err != nil {
		return nil, err
	}
	return listing, nil
}

func (s *MarketplaceService) ListActiveListings() ([]core.Listing, error) {
	return s.repo.ListActiveListings()
}

func (s *MarketplaceService) CancelListing(listingID, userID uint) error {
	listing, err := s.repo.GetListingByID(listingID)
	if err != nil {
		return err
	}
	if listing.SellerUserID != userID {
		return errors.New("only seller can cancel listing")
	}
	if listing.Status != core.ListingActive {
		return errors.New("listing is not active")
	}
	return s.repo.UpdateListingStatus(listingID, core.ListingCancelled)
}

func (s *MarketplaceService) CreateOrder(listingID, buyerID uint) (*core.Order, error) {
	listing, err := s.repo.GetListingByID(listingID)
	if err != nil {
		return nil, err
	}
	if listing.Status != core.ListingActive {
		return nil, errors.New("listing is not active")
	}
	if listing.SellerUserID == buyerID {
		return nil, errors.New("seller cannot buy their own listing")
	}

	order := &core.Order{
		ListingID:   listingID,
		BuyerUserID: buyerID,
		Status:      core.OrderPending,
	}
	if err := s.repo.CreateOrder(order); err != nil {
		return nil, err
	}
	return order, nil
}

func (s *MarketplaceService) ConfirmOrder(orderID uint, txHash string) error {
	return s.repo.ConfirmOrder(orderID, txHash)
}
