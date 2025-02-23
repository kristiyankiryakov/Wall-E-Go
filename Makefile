BROKER_BINARY=brokerApp
AUTH_BINARY=authApp
WALLET_BINARY=walletApp

## up: starts all containers in the background without forcing build
up:
	@echo "Starting Docker images..."
	docker compose up -d
	@echo "Docker images started!"

## up_build: stops docker-compose (if running), builds all projects and starts docker compose
up_build: build_broker build_auth build_wallet
	@echo "Stopping docker images (if running...)"
	docker compose down
	@echo "Building (when required) and starting docker images..."
	docker compose up --build -d
	@echo "Docker images built and started!"

## down: stop docker compose
down:
	@echo "Stopping docker compose..."
	docker compose down
	@echo "Done!"

## build_broker: builds the broker binary as a linux executable
build_broker:
	@echo "Building broker binary..."
	cd ./broker-service && env GOOS=linux CGO_ENABLED=0 go build -o ${BROKER_BINARY} ./cmd
	@echo "Done!"

## build_auth: builds the auth binary as a linux executable
build_auth:
	@echo "Building broker binary..."
	cd ./auth-service && env GOOS=linux CGO_ENABLED=0 go build -o ${AUTH_BINARY} ./cmd
	@echo "Done!"

## build_wallet: builds the wallet binary as a linux executable
build_wallet:
	@echo "Building broker binary..."
	cd ./wallet-service && env GOOS=linux CGO_ENABLED=0 go build -o ${WALLET_BINARY} ./cmd
	@echo "Done!"