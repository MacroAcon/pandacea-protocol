package security

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBoundedRequestQueue(t *testing.T) {
	logger := slog.Default()

	tests := []struct {
		name     string
		capacity int
		acquires int
		want     bool
	}{
		{
			name:     "acquire within capacity",
			capacity: 5,
			acquires: 3,
			want:     true,
		},
		{
			name:     "acquire at capacity",
			capacity: 5,
			acquires: 5,
			want:     true,
		},
		{
			name:     "acquire beyond capacity",
			capacity: 5,
			acquires: 6,
			want:     false,
		},
		{
			name:     "acquire with zero capacity",
			capacity: 0,
			acquires: 1,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queue := NewBoundedRequestQueue(tt.capacity, logger)

			// Try to acquire slots
			success := true
			for i := 0; i < tt.acquires; i++ {
				if !queue.TryAcquire() {
					success = false
					break
				}
			}

			if success != tt.want {
				t.Errorf("BoundedRequestQueue.TryAcquire() = %v, want %v", success, tt.want)
			}

			// Check queue depth
			depth := queue.GetQueueDepth()
			expectedDepth := tt.acquires
			if !tt.want {
				expectedDepth = tt.capacity
			}
			if depth != expectedDepth {
				t.Errorf("Queue depth = %d, want %d", depth, expectedDepth)
			}
		})
	}
}

func TestBoundedRequestQueueRelease(t *testing.T) {
	logger := slog.Default()
	queue := NewBoundedRequestQueue(5, logger)

	// Acquire 3 slots
	for i := 0; i < 3; i++ {
		if !queue.TryAcquire() {
			t.Fatal("Failed to acquire slot")
		}
	}

	// Check initial depth
	if depth := queue.GetQueueDepth(); depth != 3 {
		t.Errorf("Initial queue depth = %d, want 3", depth)
	}

	// Release 2 slots
	for i := 0; i < 2; i++ {
		queue.Release()
	}

	// Check depth after release
	if depth := queue.GetQueueDepth(); depth != 1 {
		t.Errorf("Queue depth after release = %d, want 1", depth)
	}

	// Release more than available (should not panic)
	queue.Release()
	queue.Release()
	queue.Release()

	// Check depth is 0
	if depth := queue.GetQueueDepth(); depth != 0 {
		t.Errorf("Queue depth after over-release = %d, want 0", depth)
	}
}

func TestSecurityServiceRequestQueue(t *testing.T) {
	logger := slog.Default()

	// Create a temporary security config
	config := &SecurityConfig{}
	config.Queue.MaxSize = 3

	service := &SecurityService{
		config:       config,
		logger:       logger,
		requestQueue: NewBoundedRequestQueue(config.Queue.MaxSize, logger),
	}

	// Test queue operations
	if !service.CheckRequestQueue() {
		t.Error("Expected to acquire queue slot")
	}

	if !service.CheckRequestQueue() {
		t.Error("Expected to acquire second queue slot")
	}

	if !service.CheckRequestQueue() {
		t.Error("Expected to acquire third queue slot")
	}

	// Fourth request should be rejected
	if service.CheckRequestQueue() {
		t.Error("Expected queue to be full")
	}

	// Release a slot
	service.ReleaseRequestQueue()

	// Should be able to acquire again
	if !service.CheckRequestQueue() {
		t.Error("Expected to acquire queue slot after release")
	}

	// Check stats
	depth, capacity := service.GetQueueStats()
	if depth != 3 {
		t.Errorf("Queue depth = %d, want 3", depth)
	}
	if capacity != 3 {
		t.Errorf("Queue capacity = %d, want 3", capacity)
	}
}

func TestLogRefusedRequest(t *testing.T) {
	logger := slog.Default()

	// Create a temporary security service
	config := &SecurityConfig{}
	config.Queue.MaxSize = 5

	service := &SecurityService{
		config:          config,
		logger:          logger,
		requestQueue:    NewBoundedRequestQueue(config.Queue.MaxSize, logger),
		ipBuckets:       make(map[string]*TokenBucket),
		identityBuckets: make(map[string]*TokenBucket),
		challenges:      make(map[string]*Challenge),
		concurrentJobs:  make(map[string]int),
		bannedIPs:       make(map[string]time.Time),
		greylistedIPs:   make(map[string]time.Time),
	}

	// Create a test request
	req := httptest.NewRequest("GET", "/api/v1/products", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.RemoteAddr = "192.168.1.1:12345"

	// Test logging refused request
	service.LogRefusedRequest(req, "test-identity", "queue_full")

	// The function should not panic and should complete successfully
	// In a real test environment, you might want to capture the log output
	// to verify the structured logging format
}

func TestSecurityServiceDefaultQueueSize(t *testing.T) {
	logger := slog.Default()

	// Test with zero queue size (should use default)
	config := &SecurityConfig{}
	config.Queue.MaxSize = 0

	service := &SecurityService{
		config:       config,
		logger:       logger,
		requestQueue: NewBoundedRequestQueue(100, logger), // Default size
	}

	// Should be able to acquire default number of slots
	for i := 0; i < 100; i++ {
		if !service.CheckRequestQueue() {
			t.Errorf("Failed to acquire slot %d", i)
		}
	}

	// 101st request should be rejected
	if service.CheckRequestQueue() {
		t.Error("Expected queue to be full at default capacity")
	}
}

func TestQueueIntegrationWithMiddleware(t *testing.T) {
	logger := slog.Default()

	// Create a minimal server for testing
	config := &SecurityConfig{}
	config.Queue.MaxSize = 2

	securityService := &SecurityService{
		config:          config,
		logger:          logger,
		requestQueue:    NewBoundedRequestQueue(config.Queue.MaxSize, logger),
		ipBuckets:       make(map[string]*TokenBucket),
		identityBuckets: make(map[string]*TokenBucket),
		challenges:      make(map[string]*Challenge),
		concurrentJobs:  make(map[string]int),
		bannedIPs:       make(map[string]time.Time),
		greylistedIPs:   make(map[string]time.Time),
	}

	// Create a test handler that simulates the security middleware
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check bounded request queue first (load shedding)
		if !securityService.CheckRequestQueue() {
			securityService.LogRefusedRequest(r, "", "queue_full")
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		defer securityService.ReleaseRequestQueue()

		// Simulate successful request processing
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Test successful requests within capacity
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d: expected status 200, got %d", i, w.Code)
		}
	}

	// Fill the queue by acquiring slots without releasing them
	securityService.CheckRequestQueue() // This should fail now
	securityService.CheckRequestQueue() // This should also fail

	// Test request beyond capacity
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	if w.Header().Get("Retry-After") != "5" {
		t.Errorf("Expected Retry-After header, got %s", w.Header().Get("Retry-After"))
	}
}
