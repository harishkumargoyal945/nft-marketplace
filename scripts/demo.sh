#!/bin/bash
set -e

API_URL="http://localhost:8081"

echo "=== NFT Marketplace Demo ==="

echo "1. Getting Deployed Addresses..."
ADDRESSES=$(curl -s $API_URL/demo/addresses)
echo $ADDRESSES | jq .
OWNER=$(echo $ADDRESSES | jq -r .owner)
SELLER=$(echo $ADDRESSES | jq -r .seller)
BUYER=$(echo $ADDRESSES | jq -r .buyer)

echo -e "\n2. Minting NFT to Seller ($SELLER)..."
MINT_RESP=$(curl -s -X POST $API_URL/mint \
  -H "Content-Type: application/json" \
  -d "{\"to\": \"$SELLER\", \"token_uri\": \"http://example.com/nft/1\"}")
echo $MINT_RESP | jq .
TOKEN_ID=$(echo $MINT_RESP | jq -r .token_id)
echo "Minted Token ID: $TOKEN_ID"

echo -e "\n3. Checking Initial Owner..."
curl -s $API_URL/nft/$TOKEN_ID/owner | jq .

echo -e "\n4. Approving Marketplace..."
APPROVE_RESP=$(curl -s -X POST $API_URL/approve \
  -H "Content-Type: application/json" \
  -d "{\"token_id\": \"$TOKEN_ID\"}")
echo $APPROVE_RESP | jq .

echo -e "\n5. Listing NFT for 1 ETH..."
PRICE_WEI="1000000000000000000"
LIST_RESP=$(curl -s -X POST $API_URL/list \
  -H "Content-Type: application/json" \
  -d "{\"token_id\": \"$TOKEN_ID\", \"price_wei\": \"$PRICE_WEI\"}")
echo $LIST_RESP | jq .

echo -e "\n6. Buying NFT as Buyer ($BUYER)..."
BUY_RESP=$(curl -s -X POST $API_URL/buy \
  -H "Content-Type: application/json" \
  -d "{\"token_id\": \"$TOKEN_ID\", \"price_wei\": \"$PRICE_WEI\"}")
echo $BUY_RESP | jq .

echo -e "\n7. Checking New Owner..."
curl -s $API_URL/nft/$TOKEN_ID/owner | jq .

echo -e "\n=== Demo Complete ==="
