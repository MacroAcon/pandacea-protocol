package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pandacea/agent-backend/internal/p2p"
	"pandacea/agent-backend/internal/policy"

	"log/slog"

	"github.com/stretchr/testify/assert"
)

func TestServer_handleGetProducts(t *testing.T) {
	// Create test logger
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))

	// Create policy engine
	policyEngine, err := policy.NewEngine(logger, "0.001") // Use the same min_price as config.yaml
	assert.NoError(t, err)                                 // Add this assertion to handle the error return

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

	// Create policy engine
	policyEngine, err := policy.NewEngine(logger, "0.001") // Use the same min_price as config.yaml
	assert.NoError(t, err)                                 // Add this assertion to handle the error return

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

	// Create policy engine
	policyEngine, err := policy.NewEngine(logger, "0.001") // Use the same min_price as config.yaml
	assert.NoError(t, err)                                 // Add this assertion to handle the error return

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

	// Create policy engine
	policyEngine, err := policy.NewEngine(logger, "0.001") // Use the same min_price as config.yaml
	assert.NoError(t, err)                                 // Add this assertion to handle the error return

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

	// Create policy engine
	policyEngine, err := policy.NewEngine(logger, "0.001")
	assert.NoError(t, err)

	// Create mock P2P node
	mockP2PNode := &p2p.Node{}

	// Create server
	server := NewServer(policyEngine, logger, mockP2PNode)

	// Create invalid request
	leaseReq := LeaseRequest{
		ProductID: "did:pandacea:earner:123/abc-456",
		MaxPrice:  "0.01",
		Duration:  "invalid-duration",
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

	// Create policy engine
	policyEngine, err := policy.NewEngine(logger, "0.001")
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

	// Create policy engine
	policyEngine, err := policy.NewEngine(logger, "0.001")
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

	// Create policy engine
	policyEngine, err := policy.NewEngine(logger, "0.001")
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
