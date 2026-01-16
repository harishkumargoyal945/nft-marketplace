# NFT Marketplace MVP Backend

A high-performance, optimized MVP backend for an NFT Marketplace built with Go, Gin, GORM, and PostgreSQL.

## Features
- User & Wallet management
- Collection and NFT registration
- Marketplace listings (Active/Sold/Cancelled)
- Transactional Order fulfillment (Escrow-like logic)
- environment-based configuration
- Docker support with PostgreSQL
- Auto-migrations

## Running Locally

### Prerequisites
- Go 1.24+
- PostgreSQL

### Steps
1. Clone the repo
2. Run `go mod tidy`
3. Set environment variables (see `.env.example`)
4. Run the API:
   ```bash
   go run ./cmd/api
   ```

## Running with Docker Compose

```bash
docker compose up --build
```

## API Endpoints

### Health Check
- `GET /health`

### Users
- `POST /v1/users` - Create user
  ```json
  { "wallet_address": "0x123...", "name": "Alice" }
  ```
- `GET /v1/users/:id` - Get user

### Collections
- `POST /v1/collections` - Create collection
  ```json
  { "creator_user_id": 1, "name": "Bored Apes", "symbol": "BAYC" }
  ```
- `GET /v1/collections` - List collections

### NFTs
- `POST /v1/nfts` - Register NFT
  ```json
  { "token_id": "1", "contract_address": "0xABC...", "chain": "ethereum", "collection_id": 1, "owner_user_id": 1, "metadata_url": "ipfs://..." }
  ```
- `GET /v1/nfts?owner_id=1&collection_id=1&chain=ethereum` - Filter NFTs

### Listings
- `POST /v1/listings` - Create listing
  ```json
  { "nft_id": 1, "seller_user_id": 1, "price_wei": "1000000000000000000", "currency": "ETH" }
  ```
- `GET /v1/listings` - List active listings
- `POST /v1/listings/:id/cancel` - Cancel listing
  ```json
  { "user_id": 1 }
  ```

### Orders
- `POST /v1/orders` - Create order
  ```json
  { "listing_id": 1, "buyer_user_id": 2 }
  ```
- `POST /v1/orders/:id/confirm` - Confirm order (marks SOLD, transfers NFT)
  ```json
  { "tx_hash": "0xTXHASH..." }
  ```

## Sample Curl Commands

```bash
# Health check
curl http://localhost:8080/health

# Create user
curl -X POST http://localhost:8080/v1/users -d '{"wallet_address": "0x123", "name": "Alice"}' -H "Content-Type: application/json"

# Create collection
curl -X POST http://localhost:8080/v1/collections -d '{"creator_user_id": 1, "name": "Bored Apes", "symbol": "BAYC"}' -H "Content-Type: application/json"

# List active listings
curl http://localhost:8080/v1/listings
```
