package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pandacea/agent-backend/internal/api"
	"pandacea/agent-backend/internal/config"
	"pandacea/agent-backend/internal/p2p"
	"pandacea/agent-backend/internal/policy"
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
