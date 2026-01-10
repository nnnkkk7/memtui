.PHONY: build test lint run clean docker-up docker-down

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
BINARY_NAME=memtui

# Build
build:
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/memtui

# Run
run: build
	./$(BINARY_NAME)

# Test
test:
	$(GOTEST) -v ./...

test-coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Lint
lint:
	golangci-lint run ./...

# Clean
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Dependencies
tidy:
	$(GOMOD) tidy

# Docker for integration tests
docker-up:
	docker compose up -d

docker-down:
	docker compose down

# Integration tests (requires docker-up)
test-integration: docker-up
	sleep 1
	$(GOTEST) -v -tags=integration ./tests/integration/...

# Run all tests including integration
test-all: test test-integration
