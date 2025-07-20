# Pandacea Agent Backend Makefile
# Provides convenient commands for building, testing, and running the agent

.PHONY: build test lint run clean

# Build the Go application
build:
	go build -o agent ./cmd/agent

# Run all unit tests
test:
	go test -v ./...

# Run the linter
lint:
	golangci-lint run

# Run the application
run:
	go run ./cmd/agent

# Clean up compiled binary
clean:
	rm -f agent

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build and run in one command
dev: build
	./agent

# Run tests with coverage
test-coverage:
	go test -v -cover ./...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./... 