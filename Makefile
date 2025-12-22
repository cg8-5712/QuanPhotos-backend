.PHONY: build run test clean fmt lint dev

# Build the application
build:
	go build -o bin/api ./cmd/api/

# Run the application
run: build
	./bin/api

# Run in development mode with hot reload (requires air)
dev:
	air -c .air.toml

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# Download dependencies
deps:
	go mod download
	go mod tidy

# Generate Swagger docs (requires swag)
swagger:
	swag init -g cmd/api/main.go -o api/

# Database migrations (requires golang-migrate)
migrate-up:
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up

migrate-down:
	migrate -path migrations -database "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

# Docker commands
docker-build:
	docker build -t quanphotos-api .

docker-run:
	docker-compose up -d

docker-stop:
	docker-compose down

# Help
help:
	@echo "Available commands:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Build and run the application"
	@echo "  make dev            - Run with hot reload (requires air)"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make fmt            - Format code"
	@echo "  make lint           - Lint code"
	@echo "  make deps           - Download dependencies"
	@echo "  make swagger        - Generate Swagger docs"
	@echo "  make migrate-up     - Run database migrations"
	@echo "  make migrate-down   - Rollback database migrations"
	@echo "  make docker-build   - Build Docker image"
	@echo "  make docker-run     - Run with Docker Compose"
