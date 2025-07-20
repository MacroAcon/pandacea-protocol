package p2p

import (
	"context"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
)

// Node represents a P2P node
type Node struct {
	host   host.Host
	dht    *dht.IpfsDHT
	logger *slog.Logger
}

// NewNode creates and initializes a new P2P node
func NewNode(ctx context.Context, listenPort int, keyFilePath string, logger *slog.Logger) (*Node, error) {
	var priv crypto.PrivKey
	var err error

	// Expand tilde in file path if present
	if keyFilePath != "" {
		if keyFilePath[0] == '~' {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %w", err)
			}
			keyFilePath = filepath.Join(homeDir, keyFilePath[1:])
		}
	}

	// Try to load existing key from file
	if keyFilePath != "" {
		if _, err := os.Stat(keyFilePath); err == nil {
			// File exists, try to load the key
			keyData, err := ioutil.ReadFile(keyFilePath)
			if err != nil {
				logger.Warn("failed to read key file, generating new key", "error", err)
			} else {
				priv, err = crypto.UnmarshalPrivateKey(keyData)
				if err != nil {
					logger.Warn("failed to unmarshal key from file, generating new key", "error", err)
				} else {
					logger.Info("loaded existing private key from file", "path", keyFilePath)
				}
			}
		}
	}

	// Generate new key if we don't have one
	if priv == nil {
		priv, _, err = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate key pair: %w", err)
		}

		// Save the new key to file if path is specified
		if keyFilePath != "" {
			// Ensure directory exists
			keyDir := filepath.Dir(keyFilePath)
			if err := os.MkdirAll(keyDir, 0700); err != nil {
				logger.Warn("failed to create key directory", "error", err, "path", keyDir)
			} else {
				// Marshal and save the key
				keyData, err := crypto.MarshalPrivateKey(priv)
				if err != nil {
					logger.Warn("failed to marshal private key", "error", err)
				} else {
					if err := ioutil.WriteFile(keyFilePath, keyData, 0600); err != nil {
						logger.Warn("failed to save private key to file", "error", err, "path", keyFilePath)
					} else {
						logger.Info("saved new private key to file", "path", keyFilePath)
					}
				}
			}
		}
	}

	// Create libp2p host
	var opts []libp2p.Option

	opts = append(opts, libp2p.Identity(priv))

	if listenPort > 0 {
		listenAddr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", listenPort))
		if err != nil {
			return nil, fmt.Errorf("failed to create listen address: %w", err)
		}
		opts = append(opts, libp2p.ListenAddrs(listenAddr))
	}

	opts = append(opts,
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
		libp2p.NATPortMap(),
	)

	host, err := libp2p.New(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p host: %w", err)
	}

	// Create KAD-DHT
	kadDHT, err := dht.New(ctx, host, dht.Mode(dht.ModeServer))
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	// Bootstrap the DHT
	if err := kadDHT.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	// Set up local peer discovery
	mdns.NewMdnsService(host, "pandacea-agent", &discoveryNotifee{host: host})

	node := &Node{
		host:   host,
		dht:    kadDHT,
		logger: logger,
	}

	// Log the peer ID for discovery
	logger.Info("P2P node initialized",
		"peer_id", host.ID().String(),
		"listen_addrs", host.Addrs(),
	)

	return node, nil
}

// GetPeerID returns the peer ID of this node
func (n *Node) GetPeerID() string {
	return n.host.ID().String()
}

// GetListenAddrs returns the listen addresses of this node
func (n *Node) GetListenAddrs() []multiaddr.Multiaddr {
	return n.host.Addrs()
}

// Close gracefully shuts down the P2P node
func (n *Node) Close() error {
	n.logger.Info("shutting down P2P node")

	if err := n.dht.Close(); err != nil {
		n.logger.Error("failed to close DHT", "error", err)
	}

	if err := n.host.Close(); err != nil {
		n.logger.Error("failed to close host", "error", err)
		return err
	}

	return nil
}

// discoveryNotifee handles peer discovery events
type discoveryNotifee struct {
	host host.Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	// Connect to discovered peers
	if err := n.host.Connect(context.Background(), pi); err != nil {
		slog.Error("failed to connect to discovered peer",
			"peer_id", pi.ID.String(),
			"error", err)
	} else {
		slog.Info("connected to discovered peer", "peer_id", pi.ID.String())
	}
}
