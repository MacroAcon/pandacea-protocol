package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log/slog"
	"net/http"
	"regexp"
	"time"

	"pandacea/agent-backend/internal/policy"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Server represents the HTTP API server
type Server struct {
	router   *chi.Mux
	policy   *policy.Engine
	logger   *slog.Logger
	products []DataProduct
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

// NewServer creates a new API server
func NewServer(policyEngine *policy.Engine, logger *slog.Logger) *Server {
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
		data, err = ioutil.ReadFile(path)
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
	// API v1 routes
	server.router.Route("/api/v1", func(r chi.Router) {
		r.Get("/products", server.handleGetProducts)
		r.Post("/leases", server.handleCreateLease)
	})

	// Health check
	server.router.Get("/health", server.handleHealth)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Perform strict input validation
	if err := server.validateLeaseRequest(&req); err != nil {
		server.logger.Error("lease request validation failed", "error", err)
		http.Error(w, fmt.Sprintf("Validation error: %s", err.Error()), http.StatusBadRequest)
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
		http.Error(w, fmt.Sprintf("Request rejected: %s", evaluation.Reason), http.StatusForbidden)
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
		http.Error(w, "Internal server error", http.StatusInternalServerError)
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
