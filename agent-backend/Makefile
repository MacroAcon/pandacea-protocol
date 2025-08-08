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

# Run PySyft training worker demo
demo-real:
	@echo "Running PySyft training worker demo..."
	@echo '{"job_id": "demo_123", "dataset": "synthetic", "task": "classification", "dp": {"enabled": true, "epsilon": 1.0, "clip": 1.0, "noise_multiplier": 0.5}, "epochs": 5, "batch_size": 32, "learning_rate": 0.01}' | python worker/train_worker.py --user-id demo_user

# Run mock training demo
demo-mock:
	@echo "Running mock training demo..."
	@echo '{"job_id": "mock_123", "dataset": "synthetic", "task": "classification", "dp": {"enabled": true, "epsilon": 1.0, "clip": 1.0, "noise_multiplier": 0.5}, "epochs": 5, "batch_size": 32, "learning_rate": 0.01}' | python worker/train_worker.py --mock --user-id demo_user

# Test privacy accountant
test-accountant:
	@echo "Testing privacy accountant..."
	@python -c "from worker.privacy_accountant import PrivacyAccountant; a = PrivacyAccountant(); print('Privacy accountant initialized successfully')"

# Validate schemas
validate-schemas:
	@echo "Validating JSON schemas..."
	@python -c "import json; json.load(open('docs/schemas/artifact.schema.json')); print('Artifact schema valid')"
	@python -c "import json; json.load(open('docs/schemas/consent.schema.json')); print('Consent schema valid')"

# Run all Sprint 1 tests
test-sprint1: test-accountant validate-schemas
	@echo "All Sprint 1 tests passed!" 