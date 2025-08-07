package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	P2P        P2PConfig        `yaml:"p2p"`
	Blockchain BlockchainConfig `yaml:"blockchain"`
	IPFS       IPFSConfig       `yaml:"ipfs"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Port     int    `yaml:"port"`
	MinPrice string `yaml:"min_price"`

	// Economic parameters based on simulation findings
	RoyaltyPercentage      float64 `yaml:"royalty_percentage"`
	SaboteurCooldown       int     `yaml:"saboteur_cooldown"`
	ReputationWeight       float64 `yaml:"reputation_weight"`
	ReputationDecayRate    float64 `yaml:"reputation_decay_rate"`
	CollusionSpendFraction float64 `yaml:"collusion_spend_fraction"`
	CollusionBonusDivisor  int     `yaml:"collusion_bonus_divisor"`
}

// P2PConfig contains P2P node configuration
type P2PConfig struct {
	ListenPort  int    `yaml:"listen_port"`
	KeyFilePath string `yaml:"key_file_path"`
}

// BlockchainConfig contains blockchain configuration
type BlockchainConfig struct {
	RPCURL          string `yaml:"rpc_url"`
	ContractAddress string `yaml:"contract_address"`
}

// IPFSConfig contains IPFS configuration
type IPFSConfig struct {
	APIURL string `yaml:"api_url"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	// Default configuration
	config := &Config{
		Server: ServerConfig{
			Port: 8080, // Default HTTP port
			// Default economic parameters based on simulation findings
			RoyaltyPercentage:      0.20,
			SaboteurCooldown:       20,
			ReputationWeight:       0.5,
			ReputationDecayRate:    0.0005,
			CollusionSpendFraction: 0.005,
			CollusionBonusDivisor:  200,
		},
		P2P: P2PConfig{
			ListenPort: 0, // Let libp2p choose a random port
		},
		Blockchain: BlockchainConfig{
			RPCURL:          "http://127.0.0.1:8545", // Default Anvil RPC URL
			ContractAddress: "",                      // Must be set via environment variable
		},
		IPFS: IPFSConfig{
			APIURL: "http://127.0.0.1:5001", // Default IPFS API URL
		},
	}

	// Load from config file if it exists
	if configPath != "" {
		if err := loadFromFile(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// Override with environment variables
	loadFromEnv(config)

	return config, nil
}

// loadFromFile loads configuration from a YAML file
func loadFromFile(config *Config, configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}

// loadFromEnv loads configuration from environment variables
func loadFromEnv(config *Config) {
	// Server configuration
	if portStr := os.Getenv("HTTP_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Server.Port = port
		}
	}

	// P2P configuration
	if portStr := os.Getenv("P2P_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			config.P2P.ListenPort = port
		}
	}

	if keyFilePath := os.Getenv("P2P_KEY_FILE"); keyFilePath != "" {
		config.P2P.KeyFilePath = keyFilePath
	}

	// Blockchain configuration
	if rpcURL := os.Getenv("RPC_URL"); rpcURL != "" {
		config.Blockchain.RPCURL = rpcURL
	}

	if contractAddress := os.Getenv("CONTRACT_ADDRESS"); contractAddress != "" {
		config.Blockchain.ContractAddress = contractAddress
	}
}

// GetServerAddr returns the server address string
func (c *Config) GetServerAddr() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}
