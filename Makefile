# ====================================================================================
# Variables
# ====================================================================================
BINARY_NAME=gopowder
MODULE_PATH=://github.com

# ====================================================================================
# Development Targets
# ====================================================================================

.PHONY: help build run test test-cover lint tidy clean

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build       Compile the application binary"
	@echo "  run         Compile and execute the application locally"
	@echo "  test        Run all unit tests"
	@echo "  test-cover  Run all tests and open the visual coverage report in your browser"
	@echo "  lint        Run the Go linter to check for code style issues"
	@echo "  tidy        Clean up and download go.mod dependencies"
	@echo "  clean       Remove compiled binaries and test artifacts"

build:
	@echo "🔨 Building the binary..."
	go build -o bin/$(BINARY_NAME) ./cmd/gopowder

run:
	@echo "🚀 Running the application..."
	go run ./cmd/gopowder

test:
	@echo "🧪 Running unit tests..."
	go test -v -race ./...

test-cover:
	@echo "📊 Generating test coverage report..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

tidy:
	@echo "🧹 Tidying up module dependencies..."
	go mod tidy

lint:
	@echo "🔍 Linting code..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint is not installed. Run: go vet ./..."; \
		go vet ./...; \
	fi

clean:
	@echo "🧹 Cleaning workspace..."
	rm -rf bin/
	rm -f coverage.out
