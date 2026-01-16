package core

import (
	"time"
)

type User struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	WalletAddress string    `gorm:"" json:"wallet_address"`
	Name          string    `json:"name"`
	CreatedAt     time.Time `json:"created_at"`
}

type Collection struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	CreatorUserID uint      `gorm:"not null" json:"creator_user_id"`
	Name          string    `gorm:"not null" json:"name"`
	Symbol        string    `gorm:"not null" json:"symbol"`
	CreatedAt     time.Time `json:"created_at"`
}

type NFT struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TokenID         string    `gorm:"not null" json:"token_id"`
	ContractAddress string    `gorm:"not null" json:"contract_address"`
	Chain           string    `gorm:"not null" json:"chain"`
	CollectionID    uint      `gorm:"not null" json:"collection_id"`
	OwnerUserID     uint      `gorm:"not null" json:"owner_user_id"`
	MetadataURL     string    `json:"metadata_url"`
	CreatedAt       time.Time `json:"created_at"`

	// Relations
	Collection Collection `gorm:"foreignKey:CollectionID" json:"collection"`
	Owner      User       `gorm:"foreignKey:OwnerUserID" json:"owner"`
}

type ListingStatus string

const (
	ListingActive    ListingStatus = "ACTIVE"
	ListingSold      ListingStatus = "SOLD"
	ListingCancelled ListingStatus = "CANCELLED"
)

type Listing struct {
	ID           uint          `gorm:"primaryKey" json:"id"`
	NFTID        uint          `gorm:"not null" json:"nft_id"`
	SellerUserID uint          `gorm:"not null" json:"seller_user_id"`
	PriceWei     string        `gorm:"not null" json:"price_wei"`
	Currency     string        `gorm:"default:'ETH'" json:"currency"`
	Status       ListingStatus `gorm:"default:'ACTIVE'" json:"status"`
	CreatedAt    time.Time     `json:"created_at"`

	// Relations
	NFT    NFT  `gorm:"foreignKey:NFTID" json:"nft"`
	Seller User `gorm:"foreignKey:SellerUserID" json:"seller"`
}

type OrderStatus string

const (
	OrderPending   OrderStatus = "PENDING"
	OrderConfirmed OrderStatus = "CONFIRMED"
	OrderFailed    OrderStatus = "FAILED"
)

type Order struct {
	ID          uint        `gorm:"primaryKey" json:"id"`
	ListingID   uint        `gorm:"not null" json:"listing_id"`
	BuyerUserID uint        `gorm:"not null" json:"buyer_user_id"`
	TxHash      *string     `json:"tx_hash"`
	Status      OrderStatus `gorm:"default:'PENDING'" json:"status"`
	CreatedAt   time.Time   `json:"created_at"`

	// Relations
	Listing Listing `gorm:"foreignKey:ListingID" json:"listing"`
	Buyer   User    `gorm:"foreignKey:BuyerUserID" json:"buyer"`
}
