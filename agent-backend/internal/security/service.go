package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
	"gopkg.in/yaml.v3"
)

// SecurityConfig holds the security configuration
type SecurityConfig struct {
	RateLimits struct {
		PerIPRPS       int `yaml:"per_ip_rps"`
		PerIdentityRPS int `yaml:"per_identity_rps"`
		Burst          int `yaml:"burst"`
	} `yaml:"rate_limits"`
	Quotas struct {
		ConcurrentJobsPerIdentity int `yaml:"concurrent_jobs_per_identity"`
	} `yaml:"quotas"`
	Backpressure struct {
		CPUHighWatermark int `yaml:"cpu_high_watermark"`
		MemHighWatermark int `yaml:"mem_high_watermark_mb"`
	} `yaml:"backpressure"`
	Queue struct {
		MaxSize int `yaml:"max_size"`
	} `yaml:"queue"`
	Bans struct {
		GreylistSeconds int `yaml:"greylist_seconds"`
		TempBanSeconds  int `yaml:"temp_ban_seconds"`
	} `yaml:"bans"`
	RequestLimits struct {
		MaxBodySizeMB   int `yaml:"max_body_size_mb"`
		MaxHeaderSizeKB int `yaml:"max_header_size_kb"`
	} `yaml:"request_limits"`
	Auth struct {
		ChallengeTimeoutSeconds int `yaml:"challenge_timeout_seconds"`
		NonceLength             int `yaml:"nonce_length"`
	} `yaml:"auth"`
}

// TokenBucket implements a simple token bucket rate limiter
type TokenBucket struct {
	tokens     float64
	capacity   float64
	rate       float64
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity, rate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		rate:       rate,
		lastRefill: time.Now(),
	}
}

// Take attempts to take a token from the bucket
func (tb *TokenBucket) Take() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now

	// Refill tokens
	tb.tokens = min(tb.capacity, tb.tokens+elapsed*tb.rate)

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}
	return false
}

// Challenge represents an authentication challenge
type Challenge struct {
	Nonce     string    `json:"nonce"`
	Address   string    `json:"address"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// BoundedRequestQueue implements a bounded request queue for load shedding
type BoundedRequestQueue struct {
	queue    chan struct{}
	capacity int
	logger   *slog.Logger
}

// NewBoundedRequestQueue creates a new bounded request queue
func NewBoundedRequestQueue(capacity int, logger *slog.Logger) *BoundedRequestQueue {
	return &BoundedRequestQueue{
		queue:    make(chan struct{}, capacity),
		capacity: capacity,
		logger:   logger,
	}
}

// TryAcquire attempts to acquire a slot in the queue
func (bq *BoundedRequestQueue) TryAcquire() bool {
	select {
	case bq.queue <- struct{}{}:
		return true
	default:
		return false
	}
}

// Release releases a slot in the queue
func (bq *BoundedRequestQueue) Release() {
	select {
	case <-bq.queue:
		// Slot released
	default:
		// Queue was empty, nothing to release
	}
}

// GetQueueDepth returns the current queue depth
func (bq *BoundedRequestQueue) GetQueueDepth() int {
	return len(bq.queue)
}

// GetCapacity returns the queue capacity
func (bq *BoundedRequestQueue) GetCapacity() int {
	return bq.capacity
}

// SecurityService handles security controls
type SecurityService struct {
	config          *SecurityConfig
	logger          *slog.Logger
	ipBuckets       map[string]*TokenBucket
	identityBuckets map[string]*TokenBucket
	challenges      map[string]*Challenge
	concurrentJobs  map[string]int
	bannedIPs       map[string]time.Time
	greylistedIPs   map[string]time.Time
	requestQueue    *BoundedRequestQueue
	mu              sync.RWMutex
	cleanupTicker   *time.Ticker
	done            chan bool
}

// SecurityEvent represents a security event for logging
type SecurityEvent struct {
	Timestamp time.Time      `json:"ts"`
	Identity  string         `json:"identity,omitempty"`
	IP        string         `json:"ip"`
	Route     string         `json:"route"`
	Decision  string         `json:"decision"`
	Reason    string         `json:"reason"`
	Counters  map[string]int `json:"counters,omitempty"`
}

// RefusedRequestEvent represents a refused request event for logging
type RefusedRequestEvent struct {
	Timestamp     time.Time `json:"ts"`
	IP            string    `json:"ip"`
	Identity      string    `json:"identity,omitempty"`
	Route         string    `json:"route"`
	Method        string    `json:"method"`
	UserAgent     string    `json:"user_agent"`
	Reason        string    `json:"reason"`
	QueueDepth    int       `json:"queue_depth"`
	QueueCapacity int       `json:"queue_capacity"`
	RateLimited   bool      `json:"rate_limited"`
	Backpressure  bool      `json:"backpressure"`
	TraceID       string    `json:"trace_id,omitempty"`
}

// NewSecurityService creates a new security service
func NewSecurityService(configPath string, logger *slog.Logger) (*SecurityService, error) {
	config, err := loadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load security config: %w", err)
	}

	// Set queue size from environment variable or config
	queueSize := config.Queue.MaxSize
	if queueSize <= 0 {
		// Check environment variable
		if envSize := os.Getenv("BACKEND_QUEUE_SIZE"); envSize != "" {
			if parsed, err := strconv.Atoi(envSize); err == nil && parsed > 0 {
				queueSize = parsed
			}
		}
		// Fall back to default if still not set
		if queueSize <= 0 {
			queueSize = 100 // Default queue size
		}
	}

	service := &SecurityService{
		config:          config,
		logger:          logger,
		ipBuckets:       make(map[string]*TokenBucket),
		identityBuckets: make(map[string]*TokenBucket),
		challenges:      make(map[string]*Challenge),
		concurrentJobs:  make(map[string]int),
		bannedIPs:       make(map[string]time.Time),
		greylistedIPs:   make(map[string]time.Time),
		requestQueue:    NewBoundedRequestQueue(queueSize, logger),
		done:            make(chan bool),
	}

	// Start cleanup goroutine
	service.cleanupTicker = time.NewTicker(1 * time.Minute)
	go service.cleanupRoutine()

	return service, nil
}

// loadConfig loads the security configuration from file
func loadConfig(configPath string) (*SecurityConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config SecurityConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// cleanupRoutine periodically cleans up expired challenges and bans
func (s *SecurityService) cleanupRoutine() {
	for {
		select {
		case <-s.cleanupTicker.C:
			s.cleanup()
		case <-s.done:
			return
		}
	}
}

// cleanup removes expired challenges and bans
func (s *SecurityService) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	// Clean up expired challenges
	for nonce, challenge := range s.challenges {
		if now.After(challenge.ExpiresAt) {
			delete(s.challenges, nonce)
		}
	}

	// Clean up expired bans
	for ip, banTime := range s.bannedIPs {
		if now.After(banTime) {
			delete(s.bannedIPs, ip)
		}
	}

	// Clean up expired greylist entries
	for ip, greylistTime := range s.greylistedIPs {
		if now.After(greylistTime) {
			delete(s.greylistedIPs, ip)
		}
	}
}

// Shutdown stops the security service
func (s *SecurityService) Shutdown() {
	if s.cleanupTicker != nil {
		s.cleanupTicker.Stop()
	}
	close(s.done)
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check for forwarded headers
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

// CheckRateLimit checks if the request should be rate limited
func (s *SecurityService) CheckRateLimit(r *http.Request, identity string) (bool, time.Duration) {
	clientIP := getClientIP(r)

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if IP is banned
	if banTime, banned := s.bannedIPs[clientIP]; banned {
		if time.Now().Before(banTime) {
			s.logSecurityEvent(r, identity, "rate_limited", "IP banned", map[string]int{"banned_until": int(banTime.Sub(time.Now()).Seconds())})
			return false, banTime.Sub(time.Now())
		}
		delete(s.bannedIPs, clientIP)
	}

	// Check if IP is greylisted
	if greylistTime, greylisted := s.greylistedIPs[clientIP]; greylisted {
		if time.Now().Before(greylistTime) {
			s.logSecurityEvent(r, identity, "rate_limited", "IP greylisted", map[string]int{"greylisted_until": int(greylistTime.Sub(time.Now()).Seconds())})
			return false, greylistTime.Sub(time.Now())
		}
		delete(s.greylistedIPs, clientIP)
	}

	// Get or create IP bucket
	ipBucket, exists := s.ipBuckets[clientIP]
	if !exists {
		ipBucket = NewTokenBucket(float64(s.config.RateLimits.Burst), float64(s.config.RateLimits.PerIPRPS))
		s.ipBuckets[clientIP] = ipBucket
	}

	// Check IP rate limit
	if !ipBucket.Take() {
		s.greylistedIPs[clientIP] = time.Now().Add(time.Duration(s.config.Bans.GreylistSeconds) * time.Second)
		s.logSecurityEvent(r, identity, "rate_limited", "IP rate limit exceeded", map[string]int{"ip_rps": s.config.RateLimits.PerIPRPS})
		return false, time.Duration(s.config.Bans.GreylistSeconds) * time.Second
	}

	// Check identity rate limit if identity is provided
	if identity != "" {
		identityBucket, exists := s.identityBuckets[identity]
		if !exists {
			identityBucket = NewTokenBucket(float64(s.config.RateLimits.Burst), float64(s.config.RateLimits.PerIdentityRPS))
			s.identityBuckets[identity] = identityBucket
		}

		if !identityBucket.Take() {
			s.logSecurityEvent(r, identity, "rate_limited", "Identity rate limit exceeded", map[string]int{"identity_rps": s.config.RateLimits.PerIdentityRPS})
			return false, time.Duration(s.config.Bans.GreylistSeconds) * time.Second
		}
	}

	return true, 0
}

// CheckConcurrencyQuota checks if the identity has exceeded concurrent job limits
func (s *SecurityService) CheckConcurrencyQuota(identity string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentJobs := s.concurrentJobs[identity]
	if currentJobs >= s.config.Quotas.ConcurrentJobsPerIdentity {
		return false
	}

	s.concurrentJobs[identity] = currentJobs + 1
	return true
}

// ReleaseConcurrencyQuota releases a concurrent job slot
func (s *SecurityService) ReleaseConcurrencyQuota(identity string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if currentJobs := s.concurrentJobs[identity]; currentJobs > 0 {
		s.concurrentJobs[identity] = currentJobs - 1
	}
}

// CheckBackpressure checks if the system is under high load
func (s *SecurityService) CheckBackpressure() bool {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	memUsageMB := int(m.Alloc / 1024 / 1024)

	// Simple CPU check - in production you'd want more sophisticated monitoring
	// For now, we'll use a simple heuristic based on goroutine count
	cpuPressure := runtime.NumGoroutine() > 1000

	if memUsageMB > s.config.Backpressure.MemHighWatermark || cpuPressure {
		return true
	}

	return false
}

// CreateChallenge creates a new authentication challenge
func (s *SecurityService) CreateChallenge(address string) (*Challenge, error) {
	nonceBytes := make([]byte, s.config.Auth.NonceLength)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, err
	}

	nonce := hex.EncodeToString(nonceBytes)
	expiresAt := time.Now().Add(time.Duration(s.config.Auth.ChallengeTimeoutSeconds) * time.Second)

	challenge := &Challenge{
		Nonce:     nonce,
		Address:   address,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	s.challenges[nonce] = challenge
	s.mu.Unlock()

	return challenge, nil
}

// VerifyChallenge verifies an authentication challenge
func (s *SecurityService) VerifyChallenge(nonce, signature string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	challenge, exists := s.challenges[nonce]
	if !exists {
		return "", false
	}

	if time.Now().After(challenge.ExpiresAt) {
		delete(s.challenges, nonce)
		return "", false
	}

	// In a real implementation, you would verify the signature against the address
	// For now, we'll use a simple hash-based verification
	expectedHash := sha256.Sum256([]byte(nonce + challenge.Address))
	expectedSignature := hex.EncodeToString(expectedHash[:])

	if signature == expectedSignature {
		delete(s.challenges, nonce)
		return challenge.Address, true
	}

	return "", false
}

// logSecurityEvent logs a security event
func (s *SecurityService) logSecurityEvent(r *http.Request, identity, decision, reason string, counters map[string]int) {
	event := SecurityEvent{
		Timestamp: time.Now(),
		Identity:  identity,
		IP:        getClientIP(r),
		Route:     r.URL.Path,
		Decision:  decision,
		Reason:    reason,
		Counters:  counters,
	}

	eventJSON, _ := json.Marshal(event)
	s.logger.Info("security_event", "event", string(eventJSON))
}

// CheckRequestQueue checks if a request can be queued
func (s *SecurityService) CheckRequestQueue() bool {
	return s.requestQueue.TryAcquire()
}

// ReleaseRequestQueue releases a request slot
func (s *SecurityService) ReleaseRequestQueue() {
	s.requestQueue.Release()
}

// GetQueueStats returns current queue statistics
func (s *SecurityService) GetQueueStats() (depth, capacity int) {
	return s.requestQueue.GetQueueDepth(), s.requestQueue.GetCapacity()
}

// LogRefusedRequest logs a structured refused request event
func (s *SecurityService) LogRefusedRequest(r *http.Request, identity, reason string) {
	queueDepth, queueCapacity := s.GetQueueStats()

	// Check if request was rate limited
	rateLimited, _ := s.CheckRateLimit(r, identity)

	// Check if system is under backpressure
	backpressure := s.CheckBackpressure()

	// Extract trace ID if available
	traceID := ""
	if span := trace.SpanFromContext(r.Context()); span != nil {
		traceID = span.SpanContext().TraceID().String()
	}

	event := RefusedRequestEvent{
		Timestamp:     time.Now(),
		IP:            getClientIP(r),
		Identity:      identity,
		Route:         r.URL.Path,
		Method:        r.Method,
		UserAgent:     r.UserAgent(),
		Reason:        reason,
		QueueDepth:    queueDepth,
		QueueCapacity: queueCapacity,
		RateLimited:   !rateLimited,
		Backpressure:  backpressure,
		TraceID:       traceID,
	}

	s.logger.Warn("request_refused",
		"event", "refused_request",
		"ip", event.IP,
		"identity", event.Identity,
		"route", event.Route,
		"method", event.Method,
		"reason", event.Reason,
		"queue_depth", event.QueueDepth,
		"queue_capacity", event.QueueCapacity,
		"rate_limited", event.RateLimited,
		"backpressure", event.Backpressure,
		"trace_id", event.TraceID,
	)
}

// min returns the minimum of two integers
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
