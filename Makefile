# Pandacea Protocol Demo Makefile
# Provides an end-to-end demo of the protocol

.PHONY: demo demo-setup demo-deploy demo-run demo-cleanup help contracts-test contracts-coverage verify pysyft-build pysyft-up demo-real-docker agent-security-test sims-run sims-report repro-build repro-verify sbom chaos-one chaos-test

# Variables
ANVIL_PID_FILE := .anvil.pid
AGENT_PID_FILE := .agent.pid
ENV_FILE := integration/.env

# Default target
help:
	@echo "Pandacea Protocol Demo Commands:"
	@echo "  demo          - Run complete end-to-end demo"
	@echo "  demo-setup    - Set up demo environment (install deps, etc.)"
	@echo "  demo-deploy   - Deploy contracts and start services"
	@echo "  demo-run      - Run the demo script"
	@echo "  demo-cleanup  - Stop services and clean up"
	@echo ""
	@echo "Development Commands:"
	@echo "  contracts-test     - Run Foundry tests with verbose output"
	@echo "  contracts-coverage - Run coverage check with thresholds"
	@echo "  verify            - Run full verification (tests + coverage + agent + SDK)"
	@echo "  agent-security-test - Run agent security tests"
	@echo "  sims-run          - Run adversarial economic simulations"
	@echo "  sims-report       - Generate simulation report and plots"
	@echo "  pysyft-build      - Build PySyft Docker image"
	@echo "  pysyft-up         - Start PySyft worker container"
	@echo "  demo-real-docker  - Run demo with real PySyft via Docker"
	@echo "  obs-up          - Start OpenTelemetry collector, Prometheus, Grafana"
	@echo "  obs-down        - Stop observability stack"
	@echo ""
	@echo "Chaos Engineering Commands:"
	@echo "  chaos-one         - Run a single chaos scenario (set SCENARIO=?)"
	@echo "  chaos-test        - Run a tiny subset of chaos scenarios serially"
	@echo ""
	@echo "Reproducible Build Commands:"
	@echo "  repro-build     - Build reproducible artifacts (Go, Python, containers)"
	@echo "  repro-verify    - Verify reproducibility by building twice and comparing"
	@echo "  sbom            - Generate SBOMs for all components"

# Complete demo flow
demo: demo-setup demo-deploy demo-run demo-cleanup

# Set up demo environment
demo-setup:
	@echo "🔧 Setting up demo environment..."
	@echo "Installing Python dependencies for builder-sdk..."
	cd builder-sdk && pip install -e . --quiet
	@echo "Installing integration dependencies..."
	cd integration && pip install -r requirements.txt python-dotenv --quiet
	@echo "Building agent backend..."
	cd agent-backend && make build --quiet
	@echo "✅ Demo environment set up"

# Contract testing and coverage
contracts-test:
	@echo "🧪 Running Foundry tests..."
	cd contracts && forge test -vvv

contracts-coverage:
	@echo "📊 Running coverage check..."
	cd contracts && forge coverage --report lcov
	cd contracts && python scripts/coverage/check_coverage.py

# Full verification
verify: contracts-test contracts-coverage
	@echo "🔍 Running agent backend tests..."
	cd agent-backend && go test ./... -v
	@echo "🔍 Running privacy accountant tests..."
	cd agent-backend/worker && python -m pytest tests/ -v
	@echo "🔍 Running SDK tests..."
	cd builder-sdk && python -m pytest tests/ -v
	@echo "✅ All verification checks passed"

# Agent security tests
agent-security-test:
	@echo "🔒 Running agent security tests..."
	cd agent-backend && go test -v -run TestRateLimiting -run TestAuthenticationChallenge -run TestAuthenticationVerification -run TestConcurrencyQuota -run TestBackpressure -run TestSecurityHeaders -run TestLegacyEndpointsWithDeprecation -run TestRequestSizeLimits -run TestSecurityEventLogging -run TestRateLimitRecovery
	@echo "✅ Agent security tests completed"

# Adversarial economic simulations
sims-run:
	@echo "🎯 Running adversarial economic simulations..."
	python sims/run_sweeps.py
	@echo "✅ Simulations completed"

sims-report:
	@echo "📊 Generating simulation report..."
	jupyter nbconvert --to html --execute sims/plots/report.ipynb --output docs/economics/sim_report.html
	@echo "✅ Report generated at docs/economics/sim_report.html"

# PySyft Docker integration
pysyft-build:
	@echo "🐳 Building PySyft Docker image..."
	docker build -f Dockerfile.pysyft -t pandacea/pysyft-worker:local .

pysyft-up:
	@echo "🚀 Starting PySyft worker container..."
	docker compose -f docker-compose.pysyft.yml up --build -d

demo-real-docker:
	@echo "🎯 Running demo with real PySyft via Docker..."
	USE_DOCKER=1 make demo

# Observability stack
obs-up:
	@echo "📈 Starting observability stack..."
	docker compose up -d otel-collector prometheus grafana
	@echo "✅ Observability available: Grafana http://localhost:3000, Prometheus http://localhost:9091"

obs-down:
	@echo "🛑 Stopping observability stack..."
	docker compose stop otel-collector prometheus grafana

# Chaos Engineering Commands
chaos-one:
	@echo "🎲 Running chaos scenario: $(or $(SCENARIO),evm_rpc_flap)"
	@if [ -z "$(SCENARIO)" ]; then \
		echo "No SCENARIO specified, using default: evm_rpc_flap"; \
		cd integration/chaos && python run_chaos.py --scenario evm_rpc_flap; \
	else \
		cd integration/chaos && python run_chaos.py --scenario $(SCENARIO); \
	fi

chaos-test:
	@echo "🧪 Running chaos test suite..."
	@echo "Testing EVM RPC flap scenario..."
	cd integration/chaos && python run_chaos.py --scenario evm_rpc_flap
	@echo "✅ Chaos test suite completed"

# Deploy contracts and start services
demo-deploy:
	@echo "🚀 Starting demo services..."
	
	@echo "Starting Anvil blockchain..."
	anvil --port 8545 --accounts 10 --mnemonic "test test test test test test test test test test test junk" > anvil.log 2>&1 & echo $$! > $(ANVIL_PID_FILE)
	@sleep 3
	
	@echo "Deploying contracts..."
	cd contracts && \
	PRIVATE_KEY=0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80 \
	forge script scripts/deploy.sol:DeployScript --rpc-url http://localhost:8545 --broadcast --legacy > deploy.log 2>&1
	
	@echo "Extracting contract addresses..."
	@# Extract contract addresses from deploy logs and write to .env file
	@echo "# Generated by demo deploy script" > $(ENV_FILE)
	@echo "AGENT_URL=http://localhost:8080" >> $(ENV_FILE)
	@echo "RPC_URL=http://127.0.0.1:8545" >> $(ENV_FILE)
	@echo "CHAIN_ID=31337" >> $(ENV_FILE)
	@echo "TIMEOUT_SECONDS=300" >> $(ENV_FILE)
	@echo "POLL_INTERVAL_SECONDS=2" >> $(ENV_FILE)
	@# Extract addresses from forge output (this is a simplified approach)
	@echo 'PGT_ADDRESS="0x5FbDB2315678afecb367f032d93F642f64180aa3"' >> $(ENV_FILE)
	@echo 'LEASE_ADDRESS="0xe7f1725E7734CE288F8367e1Bb143E90bb3F0512"' >> $(ENV_FILE)
	@echo "Contract addresses written to $(ENV_FILE)"
	
	@echo "Starting agent backend..."
	cd agent-backend && ./agent > agent.log 2>&1 & echo $$! > ../$(AGENT_PID_FILE)
	@sleep 5
	
	@echo "✅ Services started successfully"
	@echo "   - Anvil: http://localhost:8545"
	@echo "   - Agent: http://localhost:8080"

# Run the demo script
demo-run:
	@echo "🎯 Running vertical slice demo..."
	cd integration && python demo_vertical_slice.py

# Clean up services
demo-cleanup:
	@echo "🧹 Cleaning up demo services..."
	@if [ -f $(AGENT_PID_FILE) ]; then \
		echo "Stopping agent backend..."; \
		kill `cat $(AGENT_PID_FILE)` 2>/dev/null || true; \
		rm $(AGENT_PID_FILE); \
	fi
	@if [ -f $(ANVIL_PID_FILE) ]; then \
		echo "Stopping Anvil blockchain..."; \
		kill `cat $(ANVIL_PID_FILE)` 2>/dev/null || true; \
		rm $(ANVIL_PID_FILE); \
	fi
	@echo "Cleaning up log files..."
	@rm -f anvil.log agent.log contracts/deploy.log
	@echo "Cleaning up data products..."
	@rm -rf agent-backend/data/products/*
	@echo "✅ Cleanup completed"

# Force cleanup (if processes are stuck)
demo-force-cleanup:
	@echo "🔥 Force cleaning up demo services..."
	@pkill -f "anvil.*8545" || true
	@pkill -f "agent-backend" || true
	@rm -f $(ANVIL_PID_FILE) $(AGENT_PID_FILE)
	@rm -f anvil.log agent.log contracts/deploy.log
	@rm -rf agent-backend/data/products/*
	@echo "✅ Force cleanup completed"

# Check if services are running
demo-status:
	@echo "📊 Demo services status:"
	@if [ -f $(ANVIL_PID_FILE) ] && kill -0 `cat $(ANVIL_PID_FILE)` 2>/dev/null; then \
		echo "   ✅ Anvil blockchain: Running (PID: `cat $(ANVIL_PID_FILE)`)"; \
	else \
		echo "   ❌ Anvil blockchain: Not running"; \
	fi
	@if [ -f $(AGENT_PID_FILE) ] && kill -0 `cat $(AGENT_PID_FILE)` 2>/dev/null; then \
		echo "   ✅ Agent backend: Running (PID: `cat $(AGENT_PID_FILE)`)"; \
	else \
		echo "   ❌ Agent backend: Not running"; \
	fi
	@if [ -f $(ENV_FILE) ]; then \
		echo "   ✅ Environment file: $(ENV_FILE)"; \
	else \
		echo "   ❌ Environment file: Not found"; \
	fi

# Reproducible build targets
repro-build: repro-build-go repro-build-python repro-build-container

repro-build-go:
	@echo "🔧 Building reproducible Go binary..."
	@mkdir -p artifacts/bin artifacts/sbom
	cd agent-backend && \
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 \
	go build -trimpath -buildvcs=false -ldflags="-s -w" \
	-o ../artifacts/bin/agent-backend ./cmd/agent
	@echo "📋 Generating Go SBOM..."
	cd agent-backend && syft . -o spdx-json -o ../artifacts/sbom/agent-backend.spdx.json
	@echo "✅ Go binary and SBOM generated in artifacts/"

repro-build-python:
	@echo "🐍 Building reproducible Python wheel..."
	@mkdir -p artifacts/dist artifacts/sbom
	cd builder-sdk && \
	poetry install --no-dev --no-interaction && \
	PYTHONHASHSEED=0 SOURCE_DATE_EPOCH=0 poetry build --format wheel --no-interaction && \
	cp dist/*.whl ../artifacts/dist/
	@echo "📋 Generating Python SBOM..."
	cd builder-sdk && syft . -o spdx-json -o ../artifacts/sbom/pandacea-sdk.spdx.json
	@echo "✅ Python wheel and SBOM generated in artifacts/"

repro-build-container:
	@echo "🐳 Building reproducible container..."
	@mkdir -p artifacts/sbom
	docker build -f agent-backend/Dockerfile -t pandacea/agent-backend:repro agent-backend
	@echo "📋 Generating container SBOM..."
	syft pandacea/agent-backend:repro -o spdx-json -o artifacts/sbom/agent-backend-container.spdx.json
	@echo "✅ Container and SBOM generated in artifacts/"

repro-verify:
	@echo "🔍 Verifying reproducible builds..."
	powershell -ExecutionPolicy Bypass -File scripts/repro/verify_reproducibility.ps1

sbom:
	@echo "📋 Generating SBOMs..."
	@mkdir -p artifacts/sbom
	cd agent-backend && syft . -o spdx-json -o ../artifacts/sbom/agent-backend.spdx.json
	cd builder-sdk && syft . -o spdx-json -o ../artifacts/sbom/pandacea-sdk.spdx.json
	@echo "✅ SBOMs generated in artifacts/sbom/"
