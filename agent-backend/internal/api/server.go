package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"pandacea/agent-backend/internal/p2p"
	"pandacea/agent-backend/internal/policy"
	"pandacea/agent-backend/internal/privacy"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/libp2p/go-libp2p/core/peer"
)

// LeaseProposalState represents the state of a lease proposal
type LeaseProposalState struct {
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	LeaseID     *uint64   `json:"leaseId,omitempty"`
	SpenderAddr string    `json:"spenderAddr,omitempty"`
	EarnerAddr  string    `json:"earnerAddr,omitempty"`
	Price       *string   `json:"price,omitempty"`
}

// Server represents the HTTP API server
type Server struct {
	router         *chi.Mux
	policy         *policy.Engine
	logger         *slog.Logger
	products       []DataProduct
	p2pNode        *p2p.Node
	pendingLeases  map[string]*LeaseProposalState
	leasesMutex    sync.RWMutex
	privacyService privacy.PrivacyService
}

// DataProduct represents a data product as per API specification
type DataProduct struct {
	ProductID string   `json:"productId"`
	Name      string   `json:"name"`
	DataType  string   `json:"dataType"`
	Keywords  []string `json:"keywords"`
}

// ProductsResponse represents the response for the products endpoint
type ProductsResponse struct {
	Data       []DataProduct `json:"data"`
	NextCursor string        `json:"nextCursor"`
}

// LeaseRequest represents a lease request as per API specification
type LeaseRequest struct {
	ProductID string `json:"productId"`
	MaxPrice  string `json:"maxPrice"`
	Duration  string `json:"duration"`
}

// LeaseResponse represents the response for the lease endpoint
type LeaseResponse struct {
	LeaseProposalID string `json:"leaseProposalId"`
}

// DisputeRequest represents a dispute request
type DisputeRequest struct {
	Reason      string `json:"reason"`
	StakeAmount string `json:"stakeAmount"` // PGT token amount to stake
}

// DisputeResponse represents the response for the dispute endpoint
type DisputeResponse struct {
	DisputeID string `json:"disputeId"`
	Status    string `json:"status"`
}

// ErrorResponse represents a standardized error response as per API specification
type ErrorResponse struct {
	Error struct {
		Code      string `json:"code"`
		Message   string `json:"message"`
		RequestID string `json:"requestId"`
	} `json:"error"`
}

// Error codes for standardized error responses
const (
	ErrorCodeValidationError = "VALIDATION_ERROR"
	ErrorCodePolicyRejection = "POLICY_REJECTION"
	ErrorCodeUnauthorized    = "UNAUTHORIZED"
	ErrorCodeForbidden       = "FORBIDDEN"
	ErrorCodeInternalError   = "INTERNAL_ERROR"
	ErrorCodeInvalidRequest  = "INVALID_REQUEST"
)

// sendErrorResponse sends a standardized error response
func (server *Server) sendErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, errorCode, message string) {
	requestID := middleware.GetReqID(r.Context())
	if requestID == "" {
		requestID = "unknown"
	}

	errorResp := ErrorResponse{}
	errorResp.Error.Code = errorCode
	errorResp.Error.Message = message
	errorResp.Error.RequestID = requestID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		server.logger.Error("failed to encode error response", "error", err)
		// Fallback to simple error if JSON encoding fails
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// NewServer creates a new API server
func NewServer(policyEngine *policy.Engine, logger *slog.Logger, p2pNode *p2p.Node, privacyService privacy.PrivacyService) *Server {
	router := chi.NewRouter()

	// Add middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	// Add structured logging middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
			next.ServeHTTP(w, r)
		})
	})

	server := &Server{
		router:         router,
		policy:         policyEngine,
		logger:         logger,
		products:       []DataProduct{},
		p2pNode:        p2pNode,
		pendingLeases:  make(map[string]*LeaseProposalState),
		privacyService: privacyService,
	}

	// Load products from JSON file
	server.loadProducts()

	// Set up routes
	server.setupRoutes()

	return server
}

// loadProducts loads products from the products.json file
func (server *Server) loadProducts() {
	// Try multiple paths for products.json
	paths := []string{
		"products.json",
		"../products.json",
		"../../products.json",
		"./products.json",
	}

	var data []byte
	var err error

	for _, path := range paths {
		data, err = os.ReadFile(path)
		if err == nil {
			server.logger.Info("found products.json at", "path", path)
			break
		}
	}

	if err != nil {
		server.logger.Warn("products.json not found in any expected location, starting with empty product list", "error", err)
		return
	}

	// Parse the JSON data
	var products []DataProduct
	if err := json.Unmarshal(data, &products); err != nil {
		server.logger.Error("failed to parse products.json", "error", err)
		return
	}

	server.products = products
	server.logger.Info("loaded products from file", "count", len(products))
}

// setupRoutes configures the API routes
func (server *Server) setupRoutes() {
	// API v1 routes with signature verification
	server.router.Route("/api/v1", func(r chi.Router) {
		// Add signature verification middleware to all API routes
		r.Use(server.verifySignatureMiddleware)
		r.Get("/products", server.handleGetProducts)
		r.Post("/leases", server.handleCreateLease)
		r.Get("/leases/{leaseProposalId}", server.handleGetLeaseStatus)
		r.Post("/leases/{leaseId}/dispute", server.handleRaiseDispute)
		r.Post("/privacy/execute", server.handleExecuteComputation)
		r.Get("/privacy/results/{computation_id}", server.handleGetComputationResult)
	})

	// Health check (no signature required)
	server.router.Get("/health", server.handleHealth)
}

// verifySignatureMiddleware verifies the cryptographic signature of incoming requests
func (server *Server) verifySignatureMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract signature from header
		signature := r.Header.Get("X-Pandacea-Signature")
		if signature == "" {
			server.logger.Error("missing signature header", "path", r.URL.Path)
			server.sendErrorResponse(w, r, http.StatusUnauthorized, ErrorCodeUnauthorized, "Missing signature header")
			return
		}

		// Extract peer ID from header
		peerIDStr := r.Header.Get("X-Pandacea-Peer-ID")
		if peerIDStr == "" {
			server.logger.Error("missing peer ID header", "path", r.URL.Path)
			server.sendErrorResponse(w, r, http.StatusUnauthorized, ErrorCodeUnauthorized, "Missing peer ID header")
			return
		}

		// Parse peer ID
		peerID, err := peer.Decode(peerIDStr)
		if err != nil {
			server.logger.Error("invalid peer ID format", "peer_id", peerIDStr, "error", err)
			server.sendErrorResponse(w, r, http.StatusUnauthorized, ErrorCodeUnauthorized, "Invalid peer ID format")
			return
		}

		// Get the public key from the peer ID
		// Note: In a real implementation, you would need to store/retrieve public keys
		// associated with peer IDs. For now, we'll use a simplified approach.
		// In production, you'd want to maintain a key registry or use DHT for key discovery.
		pubKey, err := peerID.ExtractPublicKey()
		if err != nil {
			server.logger.Error("failed to extract public key from peer ID", "peer_id", peerIDStr, "error", err)
			server.sendErrorResponse(w, r, http.StatusForbidden, ErrorCodeForbidden, "Unable to verify signature")
			return
		}

		// Read request body for signature verification
		body, err := io.ReadAll(r.Body)
		if err != nil {
			server.logger.Error("failed to read request body", "error", err)
			server.sendErrorResponse(w, r, http.StatusInternalServerError, ErrorCodeInternalError, "Failed to read request body")
			return
		}

		// Restore the body for the next handler
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		// Decode the signature
		signatureBytes, err := base64.StdEncoding.DecodeString(signature)
		if err != nil {
			server.logger.Error("invalid signature format", "error", err)
			server.sendErrorResponse(w, r, http.StatusForbidden, ErrorCodeForbidden, "Invalid signature format")
			return
		}

		// Verify the signature
		// For GET requests, we'll sign an empty string or a canonical representation
		// For POST requests, we'll sign the request body
		var dataToVerify []byte
		if r.Method == "GET" {
			// For GET requests, sign a canonical representation of the request
			dataToVerify = []byte(fmt.Sprintf("%s %s", r.Method, r.URL.Path))
		} else {
			// For POST requests, sign the request body
			dataToVerify = body
		}

		// Verify the signature using the public key
		verified, err := pubKey.Verify(dataToVerify, signatureBytes)
		if err != nil {
			server.logger.Error("signature verification failed", "error", err, "peer_id", peerIDStr)
			server.sendErrorResponse(w, r, http.StatusForbidden, ErrorCodeForbidden, "Signature verification failed")
			return
		}

		if !verified {
			server.logger.Error("signature verification failed", "peer_id", peerIDStr)
			server.sendErrorResponse(w, r, http.StatusForbidden, ErrorCodeForbidden, "Invalid signature")
			return
		}

		server.logger.Info("signature verified successfully", "peer_id", peerIDStr, "path", r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// handleGetProducts handles GET /api/v1/products
func (server *Server) handleGetProducts(w http.ResponseWriter, r *http.Request) {
	server.logger.Info("products request received")

	// Return products from the loaded list
	response := ProductsResponse{
		Data:       server.products,
		NextCursor: "cursor_def456",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		server.logger.Error("failed to encode products response", "error", err)
		server.sendErrorResponse(w, r, http.StatusInternalServerError, ErrorCodeInternalError, "Failed to encode response")
		return
	}

	server.logger.Info("products response sent", "count", len(server.products))
}

// handleCreateLease handles POST /api/v1/leases
func (server *Server) handleCreateLease(w http.ResponseWriter, r *http.Request) {
	server.logger.Info("lease request received")

	// Parse request body
	var req LeaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.logger.Error("failed to decode lease request", "error", err)
		server.sendErrorResponse(w, r, http.StatusBadRequest, ErrorCodeInvalidRequest, "Invalid request body")
		return
	}

	// Perform strict input validation
	if err := server.validateLeaseRequest(&req); err != nil {
		server.logger.Error("lease request validation failed", "error", err)
		server.sendErrorResponse(w, r, http.StatusBadRequest, ErrorCodeValidationError, err.Error())
		return
	}

	// Call policy engine for evaluation
	policyReq := &policy.Request{
		ProductID: req.ProductID,
		MaxPrice:  req.MaxPrice,
		Duration:  req.Duration,
	}

	evaluation := server.policy.EvaluateRequest(r.Context(), policyReq)
	if !evaluation.Allowed {
		server.logger.Error("lease request rejected by policy", "reason", evaluation.Reason)
		server.sendErrorResponse(w, r, http.StatusForbidden, ErrorCodePolicyRejection, evaluation.Reason)
		return
	}

	// Generate a lease proposal ID (in a real implementation, this would be more sophisticated)
	leaseProposalID := fmt.Sprintf("lease_prop_%d", time.Now().UnixNano())

	// Create initial lease state
	server.UpdateLeaseStatus(leaseProposalID, "pending", nil, "", "", nil)

	// Return success response
	response := LeaseResponse{
		LeaseProposalID: leaseProposalID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		server.logger.Error("failed to encode lease response", "error", err)
		server.sendErrorResponse(w, r, http.StatusInternalServerError, ErrorCodeInternalError, "Failed to encode response")
		return
	}

	server.logger.Info("lease response sent", "lease_proposal_id", response.LeaseProposalID)
}

// validateLeaseRequest performs strict schema-based input validation
func (server *Server) validateLeaseRequest(req *LeaseRequest) error {
	// Check for required fields
	if req.ProductID == "" {
		return fmt.Errorf("productId is required")
	}
	if req.MaxPrice == "" {
		return fmt.Errorf("maxPrice is required")
	}
	if req.Duration == "" {
		return fmt.Errorf("duration is required")
	}

	// Validate productId format (did:pandacea format)
	didPattern := regexp.MustCompile(`^did:pandacea:[^:]+:[^/]+/[^/]+$`)
	if !didPattern.MatchString(req.ProductID) {
		return fmt.Errorf("productId must conform to did:pandacea format")
	}

	// Validate maxPrice format (should be a valid decimal number)
	pricePattern := regexp.MustCompile(`^\d+(\.\d+)?$`)
	if !pricePattern.MatchString(req.MaxPrice) {
		return fmt.Errorf("maxPrice must be a valid decimal number")
	}

	// Validate duration format (should be a valid time duration)
	durationPattern := regexp.MustCompile(`^\d+[dhms]$`)
	if !durationPattern.MatchString(req.Duration) {
		return fmt.Errorf("duration must be in format: <number>[d|h|m|s] (e.g., 24h, 30m)")
	}

	return nil
}

// handleHealth handles GET /health
func (server *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// handleGetLeaseStatus handles requests to get the status of a lease proposal
func (server *Server) handleGetLeaseStatus(w http.ResponseWriter, r *http.Request) {
	leaseProposalID := chi.URLParam(r, "leaseProposalId")
	if leaseProposalID == "" {
		server.sendErrorResponse(w, r, http.StatusBadRequest, ErrorCodeInvalidRequest, "Missing lease proposal ID")
		return
	}

	server.leasesMutex.RLock()
	leaseState, exists := server.pendingLeases[leaseProposalID]
	server.leasesMutex.RUnlock()

	if !exists {
		server.sendErrorResponse(w, r, http.StatusNotFound, ErrorCodeInvalidRequest, "Lease proposal not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(leaseState)
}

// UpdateLeaseStatus updates the status of a lease proposal
func (server *Server) UpdateLeaseStatus(leaseProposalID string, status string, leaseID *uint64, spenderAddr, earnerAddr string, price *string) {
	server.leasesMutex.Lock()
	defer server.leasesMutex.Unlock()

	now := time.Now()

	if existingState, exists := server.pendingLeases[leaseProposalID]; exists {
		// Update existing state
		existingState.Status = status
		existingState.UpdatedAt = now
		if leaseID != nil {
			existingState.LeaseID = leaseID
		}
		if spenderAddr != "" {
			existingState.SpenderAddr = spenderAddr
		}
		if earnerAddr != "" {
			existingState.EarnerAddr = earnerAddr
		}
		if price != nil {
			existingState.Price = price
		}
	} else {
		// Create new state
		server.pendingLeases[leaseProposalID] = &LeaseProposalState{
			Status:      status,
			CreatedAt:   now,
			UpdatedAt:   now,
			LeaseID:     leaseID,
			SpenderAddr: spenderAddr,
			EarnerAddr:  earnerAddr,
			Price:       price,
		}
	}

	server.logger.Info("lease status updated",
		"lease_proposal_id", leaseProposalID,
		"status", status,
		"lease_id", leaseID,
	)
}

// Start starts the HTTP server
func (server *Server) Start(addr string) error {
	server.logger.Info("starting HTTP server", "addr", addr)
	return http.ListenAndServe(addr, server.router)
}

// Shutdown gracefully shuts down the server
func (server *Server) Shutdown(ctx context.Context) error {
	server.logger.Info("shutting down HTTP server")
	// For a simple server, we just return nil
	// In a production environment, you'd want to implement proper shutdown
	return nil
}

// handleExecuteComputation handles privacy-preserving computation requests
func (server *Server) handleExecuteComputation(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req privacy.ComputationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.sendErrorResponse(w, r, http.StatusBadRequest, ErrorCodeInvalidRequest, "Invalid request body")
		return
	}

	// Extract spender address from signature verification
	spenderAddr := r.Header.Get("X-Pandacea-Spender-Address")
	if spenderAddr == "" {
		server.sendErrorResponse(w, r, http.StatusUnauthorized, ErrorCodeUnauthorized, "Spender address not found in request")
		return
	}

	// Verify lease is valid and authorized
	if err := server.privacyService.VerifyLease(r.Context(), req.LeaseID, spenderAddr); err != nil {
		server.logger.Error("lease verification failed", "error", err, "lease_id", req.LeaseID, "spender", spenderAddr)
		server.sendErrorResponse(w, r, http.StatusForbidden, ErrorCodeForbidden, fmt.Sprintf("Lease verification failed: %v", err))
		return
	}

	// Start the asynchronous computation
	response, err := server.privacyService.ExecuteComputation(r.Context(), &req)
	if err != nil {
		server.logger.Error("computation execution failed", "error", err, "lease_id", req.LeaseID)
		server.sendErrorResponse(w, r, http.StatusInternalServerError, ErrorCodeInternalError, "Computation execution failed")
		return
	}

	// Return 202 Accepted with computation ID
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		server.logger.Error("failed to encode response", "error", err)
	}
}

// handleGetComputationResult handles requests to get computation results
func (server *Server) handleGetComputationResult(w http.ResponseWriter, r *http.Request) {
	// Extract computation ID from URL parameters
	computationID := chi.URLParam(r, "computation_id")
	if computationID == "" {
		server.sendErrorResponse(w, r, http.StatusBadRequest, ErrorCodeInvalidRequest, "Computation ID is required")
		return
	}

	// Get the computation result
	result, err := server.privacyService.GetComputationResult(r.Context(), computationID)
	if err != nil {
		server.logger.Error("failed to get computation result", "error", err, "computation_id", computationID)
		server.sendErrorResponse(w, r, http.StatusNotFound, ErrorCodeInvalidRequest, fmt.Sprintf("Computation result not found: %v", err))
		return
	}

	// Return the result
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		server.logger.Error("failed to encode response", "error", err)
	}
}

// handleRaiseDispute handles the dispute creation endpoint
func (server *Server) handleRaiseDispute(w http.ResponseWriter, r *http.Request) {
	leaseID := chi.URLParam(r, "leaseId")
	if leaseID == "" {
		server.sendErrorResponse(w, r, http.StatusBadRequest, ErrorCodeValidationError, "Missing lease ID")
		return
	}

	// Parse request body
	var req DisputeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.logger.Error("failed to decode dispute request", "error", err)
		server.sendErrorResponse(w, r, http.StatusBadRequest, ErrorCodeValidationError, "Invalid request body")
		return
	}

	// Validate request
	if req.Reason == "" {
		server.sendErrorResponse(w, r, http.StatusBadRequest, ErrorCodeValidationError, "Dispute reason is required")
		return
	}

	if req.StakeAmount == "" {
		server.sendErrorResponse(w, r, http.StatusBadRequest, ErrorCodeValidationError, "Stake amount is required")
		return
	}

	// TODO: Implement blockchain interaction to raise dispute with stake
	// This would involve:
	// 1. Verifying the spender has sufficient PGT tokens
	// 2. Checking PGT allowance for the LeaseAgreement contract
	// 3. Calling the raiseDispute function on the smart contract
	// For now, we'll return a mock response
	server.logger.Info("stake-based dispute raised", "lease_id", leaseID, "reason", req.Reason, "stake_amount", req.StakeAmount)

	response := DisputeResponse{
		DisputeID: fmt.Sprintf("dispute_%s_%d", leaseID, time.Now().Unix()),
		Status:    "pending",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
