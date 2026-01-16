package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/user/nft-marketplace/internal/service"
)

type Handler struct {
	service *service.MarketplaceService
}

func NewHandler(service *service.MarketplaceService) *Handler {
	return &Handler{service: service}
}

// Responses
type errorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) Health(c *gin.Context) {
	if err := h.service.Health(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "message": "database unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// User Handlers
func (h *Handler) CreateUser(c *gin.Context) {
	var req struct {
		Wallet string `json:"wallet_address" binding:"required"`
		Name   string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	user, err := h.service.CreateUser(req.Wallet, req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *Handler) GetUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user, err := h.service.GetUser(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, errorResponse{Error: "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

// Collection Handlers
func (h *Handler) CreateCollection(c *gin.Context) {
	var req struct {
		CreatorID uint   `json:"creator_user_id" binding:"required"`
		Name      string `json:"name" binding:"required"`
		Symbol    string `json:"symbol" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	col, err := h.service.CreateCollection(req.CreatorID, req.Name, req.Symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, col)
}

func (h *Handler) ListCollections(c *gin.Context) {
	cols, err := h.service.ListCollections()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, cols)
}

// NFT Handlers
func (h *Handler) RegisterNFT(c *gin.Context) {
	var req struct {
		TokenID      string `json:"token_id" binding:"required"`
		Contract     string `json:"contract_address" binding:"required"`
		Chain        string `json:"chain" binding:"required"`
		CollectionID uint   `json:"collection_id" binding:"required"`
		OwnerID      uint   `json:"owner_user_id" binding:"required"`
		MetadataURL  string `json:"metadata_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	nft, err := h.service.RegisterNFT(req.TokenID, req.Contract, req.Chain, req.CollectionID, req.OwnerID, req.MetadataURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, nft)
}

func (h *Handler) ListNFTs(c *gin.Context) {
	ownerID, _ := strconv.Atoi(c.Query("owner_id"))
	collectionID, _ := strconv.Atoi(c.Query("collection_id"))
	chain := c.Query("chain")

	nfts, err := h.service.ListNFTs(uint(ownerID), uint(collectionID), chain)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, nfts)
}

// Listing Handlers
func (h *Handler) MintNFT(c *gin.Context) {
	var req struct {
		OwnerID  uint   `json:"owner_id" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Symbol   string `json:"symbol" binding:"required"`
		Desc     string `json:"description"`
		ImageURL string `json:"image_url" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	nft, err := h.service.MintNFT(req.OwnerID, req.Name, req.Symbol, req.Desc, req.ImageURL)
	if err != nil {
		log.Printf("MintNFT Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, nft)
}

func (h *Handler) CreateListing(c *gin.Context) {
	var req struct {
		NFTID    uint   `json:"nft_id" binding:"required"`
		SellerID uint   `json:"seller_user_id" binding:"required"`
		Price    string `json:"price_wei" binding:"required"`
		Currency string `json:"currency"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	if req.Currency == "" {
		req.Currency = "ETH"
	}

	listing, err := h.service.CreateListing(req.NFTID, req.SellerID, req.Price, req.Currency)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, listing)
}

func (h *Handler) ListListings(c *gin.Context) {
	listings, err := h.service.ListActiveListings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, listings)
}

func (h *Handler) CancelListing(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	if err := h.service.CancelListing(uint(id), req.UserID); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

// Order Handlers
func (h *Handler) CreateOrder(c *gin.Context) {
	var req struct {
		ListingID uint `json:"listing_id" binding:"required"`
		BuyerID   uint `json:"buyer_user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	order, err := h.service.CreateOrder(req.ListingID, req.BuyerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

func (h *Handler) ConfirmOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		TxHash string `json:"tx_hash" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}

	if err := h.service.ConfirmOrder(uint(id), req.TxHash); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "confirmed"})
}
