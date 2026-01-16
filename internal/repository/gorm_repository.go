package repository

import (
	"github.com/user/nft-marketplace/internal/core"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Ping() error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// User methods
func (r *Repository) CreateUser(user *core.User) error {
	return r.db.Where(core.User{WalletAddress: user.WalletAddress}).FirstOrCreate(user).Error
}

func (r *Repository) GetUserByID(id uint) (*core.User, error) {
	var user core.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) GetUserByWallet(wallet string) (*core.User, error) {
	var user core.User
	if err := r.db.Where("wallet_address = ?", wallet).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Collection methods
func (r *Repository) CreateCollection(collection *core.Collection) error {
	return r.db.Create(collection).Error
}

func (r *Repository) ListCollections() ([]core.Collection, error) {
	var collections []core.Collection
	if err := r.db.Find(&collections).Error; err != nil {
		return nil, err
	}
	return collections, nil
}

// NFT methods
func (r *Repository) CreateNFT(nft *core.NFT) error {
	return r.db.Create(nft).Error
}

func (r *Repository) GetNFTByID(id uint) (*core.NFT, error) {
	var nft core.NFT
	if err := r.db.First(&nft, id).Error; err != nil {
		return nil, err
	}
	return &nft, nil
}

func (r *Repository) ListNFTs(ownerID uint, collectionID uint, chain string) ([]core.NFT, error) {
	query := r.db.Model(&core.NFT{})
	if ownerID != 0 {
		query = query.Where("owner_user_id = ?", ownerID)
	}
	if collectionID != 0 {
		query = query.Where("collection_id = ?", collectionID)
	}
	if chain != "" {
		query = query.Where("chain = ?", chain)
	}
	var nfts []core.NFT
	if err := query.Find(&nfts).Error; err != nil {
		return nil, err
	}
	return nfts, nil
}

// Listing methods
func (r *Repository) CreateListing(listing *core.Listing) error {
	return r.db.Create(listing).Error
}

func (r *Repository) GetListingByID(id uint) (*core.Listing, error) {
	var listing core.Listing
	if err := r.db.First(&listing, id).Error; err != nil {
		return nil, err
	}
	return &listing, nil
}

func (r *Repository) ListActiveListings() ([]core.Listing, error) {
	var listings []core.Listing
	if err := r.db.Preload("NFT").Preload("Seller").Where("status = ?", core.ListingActive).Find(&listings).Error; err != nil {
		return nil, err
	}
	return listings, nil
}

func (r *Repository) UpdateListingStatus(id uint, status core.ListingStatus) error {
	return r.db.Model(&core.Listing{}).Where("id = ?", id).Update("status", status).Error
}

// Order methods
func (r *Repository) CreateOrder(order *core.Order) error {
	return r.db.Create(order).Error
}

func (r *Repository) GetOrderByID(id uint) (*core.Order, error) {
	var order core.Order
	if err := r.db.First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *Repository) ConfirmOrder(orderID uint, txHash string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var order core.Order
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&order, orderID).Error; err != nil {
			return err
		}

		if order.Status != core.OrderPending {
			return gorm.ErrInvalidData
		}

		var listing core.Listing
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&listing, order.ListingID).Error; err != nil {
			return err
		}

		if listing.Status != core.ListingActive {
			return gorm.ErrInvalidData
		}

		// Update order
		if err := tx.Model(&order).Updates(map[string]interface{}{
			"status":  core.OrderConfirmed,
			"tx_hash": txHash,
		}).Error; err != nil {
			return err
		}

		// Update listing
		if err := tx.Model(&listing).Update("status", core.ListingSold).Error; err != nil {
			return err
		}

		// Transfer NFT ownership
		if err := tx.Model(&core.NFT{}).Where("id = ?", listing.NFTID).Update("owner_user_id", order.BuyerUserID).Error; err != nil {
			return err
		}

		return nil
	})
}
