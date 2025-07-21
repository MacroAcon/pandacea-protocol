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
	"time"

	"pandacea/agent-backend/internal/p2p"
	"pandacea/agent-backend/internal/policy"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/libp2p/go-libp2p/core/peer"
)

// Server represents the HTTP API server
type Server struct {
	router   *chi.Mux
	policy   *policy.Engine
	logger   *slog.Logger
	products []DataProduct
	p2pNode  *p2p.Node
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
func NewServer(policyEngine *policy.Engine, logger *slog.Logger, p2pNode *p2p.Node) *Server {
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
		router:   router,
		policy:   policyEngine,
		logger:   logger,
		products: []DataProduct{},
		p2pNode:  p2pNode,
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

	// Return success response with hardcoded lease proposal ID
	response := LeaseResponse{
		LeaseProposalID: "lease_prop_123456789",
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
