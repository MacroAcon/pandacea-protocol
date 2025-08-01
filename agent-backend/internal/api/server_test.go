package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"
	"pandacea/agent-backend/internal/config"
	"pandacea/agent-backend/internal/p2p"
	"pandacea/agent-backend/internal/policy"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

// createTestServerConfig creates a ServerConfig for testing
func createTestServerConfig() config.ServerConfig {
	return config.ServerConfig{
		Port:                   8080,
		MinPrice:               "0.001",
		RoyaltyPercentage:      0.20,
		SaboteurCooldown:       20,
		ReputationWeight:       0.5,
		ReputationDecayRate:    0.0005,
		CollusionSpendFraction: 0.005,
		CollusionBonusDivisor:  200,
	}
}

func TestServer_handleGetProducts(t *testing.T) {
	// Create test logger
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	// Create policy engine with test config
	testConfig := createTestServerConfig()
	policyEngine, err := policy.NewEngine(logger, testConfig)
	assert.NoError(t, err)

	// Create mock P2P node
	mockP2PNode := &p2p.Node{}

	// Create server
	server := NewServer(policyEngine, logger, mockP2PNode)

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()

	// Call handler
	server.handleGetProducts(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response
	var response ProductsResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Assert response structure
	assert.Len(t, response.Data, 1)
	assert.Equal(t, "did:pandacea:earner:123/abc-456", response.Data[0].ProductID)
	assert.Equal(t, "Novel Package 3D Scans - Warehouse A", response.Data[0].Name)
	assert.Equal(t, "RoboticSensorData", response.Data[0].DataType)
	assert.Equal(t, []string{"robotics", "3d-scan", "lidar"}, response.Data[0].Keywords)
	assert.Equal(t, "cursor_def456", response.NextCursor)
}

func TestServer_handleCreateLease_ValidRequest(t *testing.T) {
	// Create test logger
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	// Create policy engine with test config
	testConfig := createTestServerConfig()
	policyEngine, err := policy.NewEngine(logger, testConfig)
	assert.NoError(t, err)

	// Create mock P2P node
	mockP2PNode := &p2p.Node{}

	// Create server
	server := NewServer(policyEngine, logger, mockP2PNode)

	// Create valid request
	leaseReq := LeaseRequest{
		ProductID: "did:pandacea:earner:123/abc-456",
		MaxPrice:  "0.01",
		Duration:  "24h",
	}

	reqBody, _ := json.Marshal(leaseReq)
	req := httptest.NewRequest("POST", "/api/v1/leases", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	server.handleCreateLease(w, req)

	// Assert response
	assert.Equal(t, http.StatusAccepted, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response
	var response LeaseResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Assert response
	assert.NotEmpty(t, response.LeaseProposalID)
}

func TestServer_handleCreateLease_InvalidProductID(t *testing.T) {
	// Create test logger
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	// Create policy engine with test config
	testConfig := createTestServerConfig()
	policyEngine, err := policy.NewEngine(logger, testConfig)
	assert.NoError(t, err)

	// Create mock P2P node
	mockP2PNode := &p2p.Node{}

	// Create server
	server := NewServer(policyEngine, logger, mockP2PNode)

	// Create invalid request
	leaseReq := LeaseRequest{
		ProductID: "invalid-product-id",
		MaxPrice:  "0.01",
		Duration:  "24h",
	}

	reqBody, _ := json.Marshal(leaseReq)
	req := httptest.NewRequest("POST", "/api/v1/leases", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	server.handleCreateLease(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "productId must conform to did:pandacea format")
}

func TestServer_handleCreateLease_InvalidMaxPrice(t *testing.T) {
	// Create test logger
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	// Create policy engine with test config
	testConfig := createTestServerConfig()
	policyEngine, err := policy.NewEngine(logger, testConfig)
	assert.NoError(t, err)

	// Create mock P2P node
	mockP2PNode := &p2p.Node{}

	// Create server
	server := NewServer(policyEngine, logger, mockP2PNode)

	// Create invalid request
	leaseReq := LeaseRequest{
		ProductID: "did:pandacea:earner:123/abc-456",
		MaxPrice:  "invalid-price",
		Duration:  "24h",
	}

	reqBody, _ := json.Marshal(leaseReq)
	req := httptest.NewRequest("POST", "/api/v1/leases", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	server.handleCreateLease(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "maxPrice must be a valid decimal number")
}

func TestServer_handleCreateLease_InvalidDuration(t *testing.T) {
	// Create test logger
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	// Create policy engine with test config
	testConfig := createTestServerConfig()
	policyEngine, err := policy.NewEngine(logger, testConfig)
	assert.NoError(t, err)

	// Create mock P2P node
	mockP2PNode := &p2p.Node{}

	// Create server
	server := NewServer(policyEngine, logger, mockP2PNode)

	// Create invalid request
	leaseReq := LeaseRequest{
		ProductID: "did:pandacea:earner:123/abc-456",
		MaxPrice:  "0.01",
		Duration:  "invalid",
	}

	reqBody, _ := json.Marshal(leaseReq)
	req := httptest.NewRequest("POST", "/api/v1/leases", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	server.handleCreateLease(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "duration must be in format")
}

func TestServer_handleCreateLease_MissingFields(t *testing.T) {
	// Create test logger
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	// Create policy engine with test config
	testConfig := createTestServerConfig()
	policyEngine, err := policy.NewEngine(logger, testConfig)
	assert.NoError(t, err)

	// Create mock P2P node
	mockP2PNode := &p2p.Node{}

	// Create server
	server := NewServer(policyEngine, logger, mockP2PNode)

	// Create invalid request with missing fields
	leaseReq := LeaseRequest{
		ProductID: "did:pandacea:earner:123/abc-456",
		// Missing MaxPrice and Duration
	}

	reqBody, _ := json.Marshal(leaseReq)
	req := httptest.NewRequest("POST", "/api/v1/leases", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	server.handleCreateLease(w, req)

	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "is required")
}

func TestServer_handleHealth(t *testing.T) {
	// Create test logger
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	// Create policy engine with test config
	testConfig := createTestServerConfig()
	policyEngine, err := policy.NewEngine(logger, testConfig)
	assert.NoError(t, err)

	// Create mock P2P node
	mockP2PNode := &p2p.Node{}

	// Create server
	server := NewServer(policyEngine, logger, mockP2PNode)

	// Create request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	server.handleHealth(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response
	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Assert response
	assert.Equal(t, "healthy", response["status"])
}

func TestServer_validateLeaseRequest(t *testing.T) {
	// Create test logger
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	// Create policy engine with test config
	testConfig := createTestServerConfig()
	policyEngine, err := policy.NewEngine(logger, testConfig)
	assert.NoError(t, err)

	// Create mock P2P node
	mockP2PNode := &p2p.Node{}

	// Create server
	server := NewServer(policyEngine, logger, mockP2PNode)

	tests := []struct {
		name    string
		request LeaseRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: LeaseRequest{
				ProductID: "did:pandacea:earner:123/abc-456",
				MaxPrice:  "0.01",
				Duration:  "24h",
			},
			wantErr: false,
		},
		{
			name: "valid request with different duration formats",
			request: LeaseRequest{
				ProductID: "did:pandacea:earner:123/abc-456",
				MaxPrice:  "0.01",
				Duration:  "30m",
			},
			wantErr: false,
		},
		{
			name: "invalid product ID format",
			request: LeaseRequest{
				ProductID: "invalid-format",
				MaxPrice:  "0.01",
				Duration:  "24h",
			},
			wantErr: true,
		},
		{
			name: "invalid price format",
			request: LeaseRequest{
				ProductID: "did:pandacea:earner:123/abc-456",
				MaxPrice:  "not-a-number",
				Duration:  "24h",
			},
			wantErr: true,
		},
		{
			name: "invalid duration format",
			request: LeaseRequest{
				ProductID: "did:pandacea:earner:123/abc-456",
				MaxPrice:  "0.01",
				Duration:  "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := server.validateLeaseRequest(&tt.request)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestServer_handleGetLeaseStatus(t *testing.T) {
	// Create test logger
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	// Create policy engine with test config
	testConfig := createTestServerConfig()
	policyEngine, err := policy.NewEngine(logger, testConfig)
	assert.NoError(t, err)

	// Create mock P2P node
	mockP2PNode := &p2p.Node{}

	// Create server
	server := NewServer(policyEngine, logger, mockP2PNode)

	// Test case 1: Lease proposal exists
	t.Run("lease proposal exists", func(t *testing.T) {
		// Create a test lease proposal
		leaseProposalID := "test_lease_prop_123"
		server.UpdateLeaseStatus(leaseProposalID, "pending", nil, "", "", nil)

		// Create request
		req := httptest.NewRequest("GET", "/api/v1/leases/"+leaseProposalID, nil)
		w := httptest.NewRecorder()

		// Set up chi context with URL parameter
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("leaseProposalId", leaseProposalID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Call handler
		server.handleGetLeaseStatus(w, req)

		// Assert response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		// Parse response
		var response LeaseProposalState
		err = json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Assert response
		assert.Equal(t, "pending", response.Status)
	})

	// Test case 2: Lease proposal does not exist
	t.Run("lease proposal not found", func(t *testing.T) {
		// Create request for non-existent lease proposal
		req := httptest.NewRequest("GET", "/api/v1/leases/non_existent_lease", nil)
		w := httptest.NewRecorder()

		// Set up chi context with URL parameter
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("leaseProposalId", "non_existent_lease")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Call handler
		server.handleGetLeaseStatus(w, req)

		// Assert response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test case 3: Missing lease proposal ID
	t.Run("missing lease proposal ID", func(t *testing.T) {
		// Create request without lease proposal ID
		req := httptest.NewRequest("GET", "/api/v1/leases/", nil)
		w := httptest.NewRecorder()

		// Set up chi context without URL parameter
		rctx := chi.NewRouteContext()
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		// Call handler
		server.handleGetLeaseStatus(w, req)

		// Assert response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestServer_UpdateLeaseStatus(t *testing.T) {
	// Create test logger
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	// Create policy engine with test config
	testConfig := createTestServerConfig()
	policyEngine, err := policy.NewEngine(logger, testConfig)
	assert.NoError(t, err)

	// Create mock P2P node
	mockP2PNode := &p2p.Node{}

	// Create server
	server := NewServer(policyEngine, logger, mockP2PNode)

	// Test case 1: Create new lease status
	t.Run("create new lease status", func(t *testing.T) {
		leaseProposalID := "test_lease_prop_new"
		leaseID := uint64(123)
		spenderAddr := "0x1234567890123456789012345678901234567890"
		earnerAddr := "0x0987654321098765432109876543210987654321"
		price := "1000000000000000000"

		server.UpdateLeaseStatus(leaseProposalID, "approved", &leaseID, spenderAddr, earnerAddr, &price)

		// Verify the lease status was created
		server.leasesMutex.RLock()
		leaseState, exists := server.pendingLeases[leaseProposalID]
		server.leasesMutex.RUnlock()

		assert.True(t, exists)
		assert.Equal(t, "approved", leaseState.Status)
		assert.Equal(t, &leaseID, leaseState.LeaseID)
		assert.Equal(t, spenderAddr, leaseState.SpenderAddr)
		assert.Equal(t, earnerAddr, leaseState.EarnerAddr)
		assert.Equal(t, &price, leaseState.Price)
	})

	// Test case 2: Update existing lease status
	t.Run("update existing lease status", func(t *testing.T) {
		leaseProposalID := "test_lease_prop_update"
		initialStatus := "pending"
		updatedStatus := "approved"

		// Create initial status
		server.UpdateLeaseStatus(leaseProposalID, initialStatus, nil, "", "", nil)

		// Update the status
		server.UpdateLeaseStatus(leaseProposalID, updatedStatus, nil, "", "", nil)

		// Verify the lease status was updated
		server.leasesMutex.RLock()
		leaseState, exists := server.pendingLeases[leaseProposalID]
		server.leasesMutex.RUnlock()

		assert.True(t, exists)
		assert.Equal(t, updatedStatus, leaseState.Status)
	})
}
