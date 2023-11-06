.PHONY:
.SILENT:
.DEFAULT_GOAL := run

FRONT_END_BINARY=frontend
ORDER_SERVICE_BINARY=orderserv

up:
	@echo "Starting Docker images..."
	docker-compose up -d
	@echo "Docker images started!"

build:
	@echo "Building Docker images..."
	docker-compose up -d
	@echo "Docker images build!"

up_build: build_app build_front
	@echo "Stopping docker images (if running...)"
	docker-compose down
	@echo "Building (when required) and starting docker images..."
	docker-compose up --build -d
	@echo "Docker images built and started!"

down:
	@echo "Stopping docker compose..."
	docker-compose down
	@echo "Done!"

erase:
	@echo "Erasing docker compose..."
	docker-compose down -v --remove-orphans
	docker-compose rm -v -f
	@echo "Done!"

stop:
	@echo "Stopping docker compose..."
	docker-compose stop
	@echo "Done!"

build_app:
	@echo "Building app binary..."
	cd ./order-service && env GOOS=linux CGO_ENABLED=0 go build -o ./bin/${ORDER_SERVICE_BINARY} ./cmd/main.go
	@echo "Done!"

build_front:
	@echo "Building front end binary..."
	cd ./front-end && env GOOS=linux CGO_ENABLED=0 go build -o ./bin/${FRONT_END_BINARY} ./cmd/web/main.go
	@echo "Done!"
