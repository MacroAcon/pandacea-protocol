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
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	"pandacea/agent-backend/internal/p2p"
	"pandacea/agent-backend/internal/policy"
	"pandacea/agent-backend/internal/privacy"
	"pandacea/agent-backend/internal/security"

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

// TrainingJob represents the state of a federated learning job
type TrainingJob struct {
	JobID        string     `json:"job_id"`
	Status       string     `json:"status"` // pending, running, complete, failed
	Dataset      string     `json:"dataset"`
	Task         string     `json:"task"`
	Epsilon      float64    `json:"epsilon"`
	ArtifactPath string     `json:"artifact_path,omitempty"`
	Error        string     `json:"error,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// Server represents the HTTP API server
type Server struct {
	router          *chi.Mux
	policy          *policy.Engine
	logger          *slog.Logger
	products        []DataProduct
	p2pNode         *p2p.Node
	pendingLeases   map[string]*LeaseProposalState
	leasesMutex     sync.RWMutex
	privacyService  privacy.PrivacyService
	securityService *security.SecurityService
	jobs            map[string]*TrainingJob
	jobsMutex       sync.RWMutex
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
	Reason string `json:"reason"`
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
func NewServer(policyEngine *policy.Engine, logger *slog.Logger, p2pNode *p2p.Node, privacyService privacy.PrivacyService, securityService *security.SecurityService) *Server {
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
		router:          router,
		policy:          policyEngine,
		logger:          logger,
		products:        []DataProduct{},
		p2pNode:         p2pNode,
		pendingLeases:   make(map[string]*LeaseProposalState),
		privacyService:  privacyService,
		securityService: securityService,
		jobs:            make(map[string]*TrainingJob),
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
	// Add version header middleware to all responses
	server.router.Use(server.addVersionHeader)

	// API v1 routes with signature verification
	server.router.Route("/api/v1", func(r chi.Router) {
		// Add security middleware to all API routes
		r.Use(server.securityMiddleware)
		r.Use(server.verifySignatureMiddleware)

		// Authentication endpoints (no signature required)
		r.Post("/auth/challenge", server.handleAuthChallenge)
		r.Post("/auth/verify", server.handleAuthVerify)

		// Protected endpoints
		r.Get("/products", server.handleGetProducts)
		r.Post("/leases", server.handleCreateLease)
		r.Get("/leases/{leaseProposalId}", server.handleGetLeaseStatus)
		r.Post("/leases/{leaseId}/dispute", server.handleRaiseDispute)
		r.Post("/privacy/execute", server.handleExecuteComputation)
		r.Get("/privacy/results/{computation_id}", server.handleGetComputationResult)
		r.Post("/train", server.handleTrain)
		r.Get("/aggregate/{jobId}", server.handleAggregate)
	})

	// Legacy endpoints (deprecated, will be removed in v2)
	server.router.Post("/train", server.handleTrainLegacy)
	server.router.Get("/aggregate/{jobId}", server.handleAggregateLegacy)

	// Health check (no signature required)
	server.router.Get("/health", server.handleHealth)
}

// addVersionHeader adds the API version header to all responses
func (server *Server) addVersionHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-API-Version", "v1")
		next.ServeHTTP(w, r)
	})
}

// securityMiddleware applies security controls (rate limiting, backpressure, etc.)
func (server *Server) securityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip security checks for authentication endpoints
		if r.URL.Path == "/api/v1/auth/challenge" || r.URL.Path == "/api/v1/auth/verify" {
			next.ServeHTTP(w, r)
			return
		}

		// Check backpressure
		if server.securityService.CheckBackpressure() {
			w.Header().Set("Retry-After", "30")
			server.sendErrorResponse(w, r, http.StatusServiceUnavailable, "BACKPRESSURE", "Service temporarily unavailable due to high load")
			return
		}

		// Extract identity from signature (simplified for now)
		identity := ""
		if signature := r.Header.Get("X-Signature"); signature != "" {
			// In a real implementation, you'd extract the identity from the signature
			identity = "authenticated_user"
		}

		// Check rate limits
		allowed, retryAfter := server.securityService.CheckRateLimit(r, identity)
		if !allowed {
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", retryAfter.Seconds()))
			server.sendErrorResponse(w, r, http.StatusTooManyRequests, "RATE_LIMITED", "Rate limit exceeded")
			return
		}

		// Check concurrency quota for training endpoints
		if r.URL.Path == "/api/v1/train" && identity != "" {
			if !server.securityService.CheckConcurrencyQuota(identity) {
				server.sendErrorResponse(w, r, http.StatusConflict, "QUOTA_EXCEEDED", "Concurrent job limit exceeded")
				return
			}
			// Release quota when request completes
			defer server.securityService.ReleaseConcurrencyQuota(identity)
		}

		next.ServeHTTP(w, r)
	})
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

	// TODO: Implement blockchain interaction to raise dispute with dynamic stake
	// This would involve:
	// 1. Calling getRequiredStake(leaseId) to get the calculated stake amount
	// 2. Verifying the spender has sufficient PGT tokens
	// 3. Checking PGT allowance for the LeaseAgreement contract
	// 4. Calling the raiseDispute function on the smart contract
	// For now, we'll return a mock response
	server.logger.Info("dynamic stake-based dispute raised", "lease_id", leaseID, "reason", req.Reason)

	response := DisputeResponse{
		DisputeID: fmt.Sprintf("dispute_%s_%d", leaseID, time.Now().Unix()),
		Status:    "pending",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// TrainRequest represents a federated learning training request
type TrainRequest struct {
	Dataset string `json:"dataset"`
	Task    string `json:"task"`
	DP      struct {
		Enabled bool    `json:"enabled"`
		Epsilon float64 `json:"epsilon"`
	} `json:"dp"`
}

// TrainResponse represents the response for the train endpoint
type TrainResponse struct {
	JobID string `json:"job_id"`
}

// handleTrain handles POST /train
func (server *Server) handleTrain(w http.ResponseWriter, r *http.Request) {
	server.logger.Info("training request received")

	// Parse request body
	var req TrainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.logger.Error("failed to decode train request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Dataset == "" {
		http.Error(w, "Dataset is required", http.StatusBadRequest)
		return
	}
	if req.Task == "" {
		http.Error(w, "Task is required", http.StatusBadRequest)
		return
	}

	// Generate job ID
	jobID := fmt.Sprintf("job_%d", time.Now().UnixNano())

	// Create training job
	job := &TrainingJob{
		JobID:     jobID,
		Status:    "pending",
		Dataset:   req.Dataset,
		Task:      req.Task,
		Epsilon:   req.DP.Epsilon,
		CreatedAt: time.Now(),
	}

	// Store job
	server.jobsMutex.Lock()
	server.jobs[jobID] = job
	server.jobsMutex.Unlock()

	// Start the training job asynchronously
	go server.runTrainingJob(jobID)

	// Return job ID
	response := TrainResponse{
		JobID: jobID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)

	server.logger.Info("training job queued", "job_id", jobID, "dataset", req.Dataset, "task", req.Task)
}

// handleAggregate handles GET /aggregate/{jobId}
func (server *Server) handleAggregate(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "jobId")
	if jobID == "" {
		http.Error(w, "Job ID is required", http.StatusBadRequest)
		return
	}

	server.jobsMutex.RLock()
	job, exists := server.jobs[jobID]
	server.jobsMutex.RUnlock()

	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(job)

	server.logger.Info("aggregate status requested", "job_id", jobID, "status", job.Status)
}

// runTrainingJob executes the training job by calling a Python worker
func (server *Server) runTrainingJob(jobID string) {
	server.logger.Info("starting training job", "job_id", jobID)

	// Update job status to running
	server.updateJobStatus(jobID, "running", "", "")

	// Create output directory
	outputDir := fmt.Sprintf("./data/products/%s", jobID)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		server.logger.Error("failed to create output directory", "error", err, "job_id", jobID)
		server.updateJobStatus(jobID, "failed", "", fmt.Sprintf("Failed to create output directory: %v", err))
		return
	}

	// Get job details
	server.jobsMutex.RLock()
	job := server.jobs[jobID]
	server.jobsMutex.RUnlock()

	// Check if Docker execution is enabled
	useDocker := os.Getenv("USE_DOCKER") == "1"

	if useDocker {
		server.runTrainingJobDocker(jobID, job, outputDir)
	} else {
		server.runTrainingJobLocal(jobID, job, outputDir)
	}
}

func (server *Server) runTrainingJobDocker(jobID string, job *TrainingJob, outputDir string) {
	server.logger.Info("running training job with Docker", "job_id", jobID)

	// Prepare job payload for Docker container
	jobPayload := map[string]interface{}{
		"job_id":     jobID,
		"dataset":    job.Dataset,
		"task":       job.Task,
		"epsilon":    job.Epsilon,
		"output_dir": "/app/data",
	}

	payloadBytes, err := json.Marshal(jobPayload)
	if err != nil {
		server.logger.Error("failed to marshal job payload", "error", err, "job_id", jobID)
		server.updateJobStatus(jobID, "failed", "", fmt.Sprintf("Failed to marshal job payload: %v", err))
		return
	}

	// Execute Docker container
	cmd := exec.Command("docker", "compose", "-f", "docker-compose.pysyft.yml", "run", "--rm", "pysyft-worker")
	cmd.Stdin = strings.NewReader(string(payloadBytes))

	output, err := cmd.CombinedOutput()
	if err != nil {
		server.logger.Error("Docker execution failed", "error", err, "output", string(output), "job_id", jobID)
		server.updateJobStatus(jobID, "failed", "", fmt.Sprintf("Docker execution failed: %v", err))
		return
	}

	server.logger.Info("Docker execution completed", "output", string(output), "job_id", jobID)

	// Check for output file
	aggregatePath := fmt.Sprintf("%s/aggregate.json", outputDir)
	if _, err := os.Stat(aggregatePath); os.IsNotExist(err) {
		server.logger.Error("aggregate file not found after Docker execution", "job_id", jobID)
		server.updateJobStatus(jobID, "failed", "", "Aggregate file not found after Docker execution")
		return
	}

	// Update job status to complete
	server.updateJobStatus(jobID, "complete", aggregatePath, "")
	server.logger.Info("Docker training job completed", "job_id", jobID, "output", aggregatePath)
}

func (server *Server) runTrainingJobLocal(jobID string, job *TrainingJob, outputDir string) {
	server.logger.Info("running training job locally", "job_id", jobID)

	// Check if MOCK_DP is enabled
	mockDP := os.Getenv("MOCK_DP") == "1"

	if mockDP {
		// Use the existing mock implementation
		server.runTrainingJobMock(jobID, job, outputDir)
	} else {
		// Use the real PySyft worker
		server.runTrainingJobReal(jobID, job, outputDir)
	}
}

func (server *Server) runTrainingJobMock(jobID string, job *TrainingJob, outputDir string) {
	server.logger.Info("running mock training job", "job_id", jobID)

	// Prepare Python worker command
	// This will run a simple Python script that simulates DP-SGD training
	pythonScript := fmt.Sprintf(`
import json
import numpy as np
import os
from datetime import datetime

# Simulate differential privacy parameters
epsilon = %f
dataset = "%s"
task = "%s"

# Create mock training result
result = {
    "job_id": "%s",
    "dataset": dataset,
    "task": task,
    "epsilon_used": epsilon,
    "model_accuracy": 0.85 + np.random.normal(0, 0.05),
    "samples_processed": 1000,
    "training_time_seconds": 30.5,
    "dp_noise_scale": 1.0 / epsilon,
    "timestamp": datetime.now().isoformat()
}

# Save result to output file
output_path = os.path.join("%s", "aggregate.json")
with open(output_path, "w") as f:
    json.dump(result, f, indent=2)

print(f"Training completed. Epsilon used: {epsilon}")
print(f"Output saved to: {output_path}")
`, job.Epsilon, job.Dataset, job.Task, jobID, outputDir)

	// Write Python script to temporary file
	scriptPath := fmt.Sprintf("%s/worker.py", outputDir)
	if err := os.WriteFile(scriptPath, []byte(pythonScript), 0644); err != nil {
		server.logger.Error("failed to write Python script", "error", err, "job_id", jobID)
		server.updateJobStatus(jobID, "failed", "", fmt.Sprintf("Failed to write Python script: %v", err))
		return
	}

	// Execute Python script
	cmd := fmt.Sprintf("python %s", scriptPath)
	server.logger.Info("executing Python worker", "command", cmd, "job_id", jobID)

	// For demo purposes, just sleep and create the output
	time.Sleep(10 * time.Second) // Simulate training time

	// Create the aggregate.json file
	aggregatePath := fmt.Sprintf("%s/aggregate.json", outputDir)
	result := map[string]interface{}{
		"job_id":                jobID,
		"dataset":               job.Dataset,
		"task":                  job.Task,
		"epsilon_used":          job.Epsilon,
		"model_accuracy":        0.85 + (float64(time.Now().UnixNano()%100) / 1000.0), // Random accuracy
		"samples_processed":     1000,
		"training_time_seconds": 10.0,
		"dp_noise_scale":        1.0 / job.Epsilon,
		"timestamp":             time.Now().Format(time.RFC3339),
	}

	resultBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		server.logger.Error("failed to marshal training result", "error", err, "job_id", jobID)
		server.updateJobStatus(jobID, "failed", "", fmt.Sprintf("Failed to marshal result: %v", err))
		return
	}

	if err := os.WriteFile(aggregatePath, resultBytes, 0644); err != nil {
		server.logger.Error("failed to write aggregate result", "error", err, "job_id", jobID)
		server.updateJobStatus(jobID, "failed", "", fmt.Sprintf("Failed to write result: %v", err))
		return
	}

	// Update job status to complete
	server.updateJobStatus(jobID, "complete", aggregatePath, "")
	server.logger.Info("mock training job completed", "job_id", jobID, "output", aggregatePath)
}

func (server *Server) runTrainingJobReal(jobID string, job *TrainingJob, outputDir string) {
	server.logger.Info("running real PySyft training job", "job_id", jobID)

	// Execute the real PySyft worker
	workerPath := "./worker/train_worker.py"
	cmd := exec.Command("python", workerPath,
		"--job-id", jobID,
		"--dataset", job.Dataset,
		"--task", job.Task,
		"--epsilon", fmt.Sprintf("%f", job.Epsilon),
		"--output-dir", outputDir,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		server.logger.Error("real PySyft execution failed", "error", err, "output", string(output), "job_id", jobID)
		server.updateJobStatus(jobID, "failed", "", fmt.Sprintf("Real PySyft execution failed: %v", err))
		return
	}

	server.logger.Info("real PySyft execution completed", "output", string(output), "job_id", jobID)

	// Check for output file
	aggregatePath := fmt.Sprintf("%s/aggregate.json", outputDir)
	if _, err := os.Stat(aggregatePath); os.IsNotExist(err) {
		server.logger.Error("aggregate file not found after real PySyft execution", "job_id", jobID)
		server.updateJobStatus(jobID, "failed", "", "Aggregate file not found after real PySyft execution")
		return
	}

	// Update job status to complete
	server.updateJobStatus(jobID, "complete", aggregatePath, "")
	server.logger.Info("real PySyft training job completed", "job_id", jobID, "output", aggregatePath)
}

// handleTrainLegacy handles legacy /train endpoint with deprecation warning
func (server *Server) handleTrainLegacy(w http.ResponseWriter, r *http.Request) {
	// Add deprecation warning header
	w.Header().Set("X-API-Deprecation-Warning", "This endpoint is deprecated. Use /api/v1/train instead.")

	// Log deprecation warning
	server.logger.Warn("legacy endpoint used", "endpoint", "/train", "recommendation", "Use /api/v1/train instead")

	// Call the original handler
	server.handleTrain(w, r)
}

// handleAggregateLegacy handles legacy /aggregate/{jobId} endpoint with deprecation warning
func (server *Server) handleAggregateLegacy(w http.ResponseWriter, r *http.Request) {
	// Add deprecation warning header
	w.Header().Set("X-API-Deprecation-Warning", "This endpoint is deprecated. Use /api/v1/aggregate/{jobId} instead.")

	// Log deprecation warning
	server.logger.Warn("legacy endpoint used", "endpoint", "/aggregate/{jobId}", "recommendation", "Use /api/v1/aggregate/{jobId} instead")

	// Call the original handler
	server.handleAggregate(w, r)
}

// updateJobStatus updates the status of a training job
func (server *Server) updateJobStatus(jobID, status, artifactPath, errorMsg string) {
	server.jobsMutex.Lock()
	defer server.jobsMutex.Unlock()

	job, exists := server.jobs[jobID]
	if !exists {
		server.logger.Error("job not found for status update", "job_id", jobID)
		return
	}

	job.Status = status
	if artifactPath != "" {
		job.ArtifactPath = artifactPath
	}
	if errorMsg != "" {
		job.Error = errorMsg
	}

	if status == "complete" || status == "failed" {
		now := time.Now()
		job.CompletedAt = &now
	}

	server.logger.Info("job status updated", "job_id", jobID, "status", status)
}

// AuthChallengeRequest represents a request to create an authentication challenge
type AuthChallengeRequest struct {
	Address string `json:"address"`
}

// AuthChallengeResponse represents the response to an authentication challenge
type AuthChallengeResponse struct {
	Nonce     string    `json:"nonce"`
	Address   string    `json:"address"`
	ExpiresAt time.Time `json:"expires_at"`
}

// AuthVerifyRequest represents a request to verify an authentication challenge
type AuthVerifyRequest struct {
	Nonce     string `json:"nonce"`
	Signature string `json:"signature"`
}

// AuthVerifyResponse represents the response to an authentication verification
type AuthVerifyResponse struct {
	Address string `json:"address"`
	Valid   bool   `json:"valid"`
}

// handleAuthChallenge handles authentication challenge creation
func (server *Server) handleAuthChallenge(w http.ResponseWriter, r *http.Request) {
	var req AuthChallengeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.sendErrorResponse(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Address == "" {
		server.sendErrorResponse(w, r, http.StatusBadRequest, "MISSING_ADDRESS", "Address is required")
		return
	}

	challenge, err := server.securityService.CreateChallenge(req.Address)
	if err != nil {
		server.logger.Error("failed to create challenge", "error", err, "address", req.Address)
		server.sendErrorResponse(w, r, http.StatusInternalServerError, "CHALLENGE_CREATION_FAILED", "Failed to create challenge")
		return
	}

	response := AuthChallengeResponse{
		Nonce:     challenge.Nonce,
		Address:   challenge.Address,
		ExpiresAt: challenge.ExpiresAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleAuthVerify handles authentication challenge verification
func (server *Server) handleAuthVerify(w http.ResponseWriter, r *http.Request) {
	var req AuthVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		server.sendErrorResponse(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Nonce == "" || req.Signature == "" {
		server.sendErrorResponse(w, r, http.StatusBadRequest, "MISSING_FIELDS", "Nonce and signature are required")
		return
	}

	address, valid := server.securityService.VerifyChallenge(req.Nonce, req.Signature)

	response := AuthVerifyResponse{
		Address: address,
		Valid:   valid,
	}

	w.Header().Set("Content-Type", "application/json")
	if valid {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
	}
	json.NewEncoder(w).Encode(response)
}
