.PHONY: up down deploy run demo clean test

APP_NAME = nft-marketplace

up:
	docker-compose -f docker/docker-compose.yml up -d --build hardhat
	@echo "Waiting for Hardhat to start..."
	@sleep 5

down:
	docker-compose -f docker/docker-compose.yml down

deploy:
	docker-compose -f docker/docker-compose.yml exec hardhat npx hardhat run scripts/deploy.js --network localhost

run:
	go run cmd/api/main.go

demo:
	chmod +x scripts/demo.sh
	./scripts/demo.sh

clean:
	rm -rf bin pkg/abi
