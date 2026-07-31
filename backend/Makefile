.PHONY: build test lint migrate

build:
	@echo "Building Go backend services..."
	go build ./...

test:
	go test ./...

lint:
	golangci-lint run ./...

migrate:
	@echo "Running database migrations..."
