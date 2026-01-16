package core

type MintRequest struct {
	To          string `json:"to" binding:"required"`
	TokenURI    string `json:"token_uri" binding:"required"`
	Name        string `json:"name"`
	Ticker      string `json:"ticker"`
	Description string `json:"description"`
}

type ApproveRequest struct {
	TokenID string `json:"token_id" binding:"required"`
}

type ListRequest struct {
	TokenID  string `json:"token_id" binding:"required"`
	PriceWei string `json:"price_wei" binding:"required"`
}

type BuyRequest struct {
	TokenID  string `json:"token_id" binding:"required"`
	PriceWei string `json:"price_wei" binding:"required"`
}

type BurnRequest struct {
	TokenID string `json:"token_id" binding:"required"`
}

type DelistRequest struct {
	TokenID string `json:"token_id" binding:"required"`
}

type TxResponse struct {
	TxHash  string `json:"tx_hash"`
	TokenID string `json:"token_id,omitempty"`
}

type AddressResponse struct {
	Owner          string `json:"owner"`
	Seller         string `json:"seller"`
	Buyer          string `json:"buyer"`
	NFTContract    string `json:"nft_contract"`
	MarketContract string `json:"market_contract"`
}

type ListingInfo struct {
	TokenID string `json:"token_id"`
	Price   string `json:"price"`
	Seller  string `json:"seller"`
	Active  bool   `json:"active"`
}

type NFTOwnerResponse struct {
	TokenIDs []string `json:"token_ids"`
}
