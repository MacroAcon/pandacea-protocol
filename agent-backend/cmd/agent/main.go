package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pandacea/agent-backend/internal/api"
	"pandacea/agent-backend/internal/config"
	"pandacea/agent-backend/internal/contracts"
	"pandacea/agent-backend/internal/p2p"
	"pandacea/agent-backend/internal/policy"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Set up structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	logger.Info("starting Pandacea agent backend")

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger.Info("configuration loaded",
		"http_port", cfg.Server.Port,
		"p2p_port", cfg.P2P.ListenPort,
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize policy engine
	policyEngine, err := policy.NewEngine(logger, cfg.Server) // Pass the whole ServerConfig
	if err != nil {
		logger.Error("failed to initialize policy engine", "error", err)
		os.Exit(1)
	}

	// Initialize P2P node
	p2pNode, err := p2p.NewNode(ctx, cfg.P2P.ListenPort, cfg.P2P.KeyFilePath, logger)
	if err != nil {
		logger.Error("failed to initialize P2P node", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := p2pNode.Close(); err != nil {
			logger.Error("failed to close P2P node", "error", err)
		}
	}()

	// Initialize API server
	apiServer := api.NewServer(policyEngine, logger, p2pNode)

	// Start API server in a goroutine
	go func() {
		if err := apiServer.Start(cfg.GetServerAddr()); err != nil {
			logger.Error("failed to start API server", "error", err)
			cancel() // Signal shutdown
		}
	}()

	// Start blockchain event listener if blockchain configuration is provided
	if cfg.Blockchain.RPCURL != "" && cfg.Blockchain.ContractAddress != "" {
		go func() {
			if err := startEventListener(ctx, cfg, apiServer, logger); err != nil {
				logger.Error("failed to start blockchain event listener", "error", err)
				cancel() // Signal shutdown
			}
		}()
	} else {
		logger.Warn("blockchain configuration not provided, skipping event listener")
	}

	// Log startup information
	logger.Info("agent backend started successfully",
		"peer_id", p2pNode.GetPeerID(),
		"http_addr", cfg.GetServerAddr(),
		"p2p_addrs", p2pNode.GetListenAddrs(),
	)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	select {
	case sig := <-sigChan:
		logger.Info("received shutdown signal", "signal", sig)
	case <-ctx.Done():
		logger.Info("shutdown requested via context")
	}

	// Perform graceful shutdown
	logger.Info("starting graceful shutdown")

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown API server
	if err := apiServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown API server", "error", err)
	}

	logger.Info("agent backend shutdown complete")
}

// startEventListener starts listening for blockchain events
func startEventListener(ctx context.Context, cfg *config.Config, apiServer *api.Server, logger *slog.Logger) error {
	logger.Info("connecting to blockchain", "rpc_url", cfg.Blockchain.RPCURL)

	// Connect to the Ethereum client
	client, err := ethclient.Dial(cfg.Blockchain.RPCURL)
	if err != nil {
		return err
	}
	defer client.Close()

	// Create contract instance
	contractAddress := common.HexToAddress(cfg.Blockchain.ContractAddress)
	contract, err := contracts.NewLeaseAgreement(contractAddress, client)
	if err != nil {
		return err
	}

	logger.Info("blockchain connection established",
		"contract_address", cfg.Blockchain.ContractAddress,
		"rpc_url", cfg.Blockchain.RPCURL,
	)

	// Subscribe to LeaseCreated events
	logs := make(chan *contracts.LeaseAgreementLeaseCreated)
	sub, err := contract.WatchLeaseCreated(nil, logs, nil, nil, nil)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	logger.Info("subscribed to LeaseCreated events")

	// Process events
	for {
		select {
		case err := <-sub.Err():
			logger.Error("subscription error", "error", err)
			return err
		case log := <-logs:
			handleLeaseCreatedEvent(log, apiServer, logger)
		case <-ctx.Done():
			logger.Info("shutting down event listener")
			return nil
		}
	}
}

// handleLeaseCreatedEvent processes a LeaseCreated event
func handleLeaseCreatedEvent(event *contracts.LeaseAgreementLeaseCreated, apiServer *api.Server, logger *slog.Logger) {
	logger.Info("received LeaseCreated event",
		"lease_id", event.LeaseId,
		"spender", event.Spender.Hex(),
		"earner", event.Earner.Hex(),
		"price", event.Price.String(),
	)

	// Convert price to string
	priceStr := event.Price.String()

	// Convert lease ID to uint64 for storage
	// Note: This is a simplified approach. In production, you might want to store the full bytes32
	leaseID := uint64(0) // We'll use 0 for now since we don't have a direct mapping

	// For now, we'll use a simple mapping from lease ID to proposal ID
	// In a real implementation, you might want to maintain a mapping table
	leaseProposalID := fmt.Sprintf("lease_prop_%x", event.LeaseId)

	// Update the lease status in the API server
	apiServer.UpdateLeaseStatus(
		leaseProposalID,
		"approved",
		&leaseID,
		event.Spender.Hex(),
		event.Earner.Hex(),
		&priceStr,
	)
}
