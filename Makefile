.PHONY: build run test clean dev debug dev-verbose run-rest

# Generate mock
generate-mock:
	go generate ./...

# Generate swagger docs
swagger:
	go install github.com/swaggo/swag/cmd/swag@latest
	swag init -g cmd/rest.go

# Build the application, run swagger before build
build: swagger generate-mock
	go build -o moneytransfer

# Run the application
run: build
	./moneytransfer rest

# Run tests
test:
	TESTCONTAINERS_RYUK_DISABLED=false go test ./... -cover -v

# Clean build artifacts
clean:
	rm -f moneytransfer
	rm -rf tmp

# Run the application in development mode using Air
dev:
	air

# Run Air in verbose mode for debugging
dev-verbose:
	air -v

# Run the application directly for debugging
run-debug:
	go build -o ./tmp/moneytransfer
	./tmp/moneytransfer rest

# Debug target to check environment and commands
debug:
	@echo "Current directory:"
	@pwd
	@echo "\nContents of current directory:"
	@ls -la
	@echo "\nChecking if moneytransfer binary exists:"
	@if [ -f "./tmp/moneytransfer" ]; then echo "Binary exists"; else echo "Binary not found"; fi
	@echo "\nTrying to run moneytransfer rest manually:"
	@./tmp/moneytransfer rest || echo "Failed to run moneytransfer rest"

# Install dependencies (including Air)
deps:
	go mod tidy
	go install github.com/cosmtrek/air@latest

# Default target
all: build

# Run the REST API server directly
run-rest: build
	./moneytransfer rest
