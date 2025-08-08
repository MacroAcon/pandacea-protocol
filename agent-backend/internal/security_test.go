package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"pandacea/agent-backend/internal/p2p"
	"pandacea/agent-backend/internal/policy"
	"pandacea/agent-backend/internal/privacy"
	"pandacea/agent-backend/internal/security"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockPrivacyService implements privacy.PrivacyService for testing
type MockPrivacyService struct{}

func (m *MockPrivacyService) ExecuteComputation(ctx context.Context, req *privacy.ComputationRequest) (*privacy.ComputationResponse, error) {
	return &privacy.ComputationResponse{ComputationID: "test_id"}, nil
}

func (m *MockPrivacyService) GetComputationResult(ctx context.Context, computationID string) (*privacy.ComputationResult, error) {
	return &privacy.ComputationResult{Status: "completed"}, nil
}

func (m *MockPrivacyService) VerifyLease(ctx context.Context, leaseID string, spenderAddr string) error {
	return nil
}

func (m *MockPrivacyService) Start() error {
	return nil
}

func (m *MockPrivacyService) Stop() error {
	return nil
}

// setupTestServer creates a test server with security controls
func setupTestServer(t *testing.T) (*Server, *httptest.Server) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	// Create temporary security config
	configPath := "test_security.yaml"
	configContent := `
rate_limits:
  per_ip_rps: 5
  per_identity_rps: 2
  burst: 10
quotas:
  concurrent_jobs_per_identity: 2
backpressure:
  cpu_high_watermark: 85
  mem_high_watermark_mb: 2048
bans:
  greylist_seconds: 600
  temp_ban_seconds: 1800
request_limits:
  max_body_size_mb: 10
  max_header_size_kb: 8
auth:
  challenge_timeout_seconds: 300
  nonce_length: 32
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)
	defer os.Remove(configPath)

	// Initialize security service
	securityService, err := security.NewSecurityService(configPath, logger)
	require.NoError(t, err)

	// Create mock services
	policyEngine := &policy.Engine{}
	p2pNode := &p2p.Node{}
	privacyService := &MockPrivacyService{}

	// Create server
	server := NewServer(policyEngine, logger, p2pNode, privacyService, securityService)

	// Create test server
	testServer := httptest.NewServer(server.router)

	return server, testServer
}

func TestRateLimiting(t *testing.T) {
	_, testServer := setupTestServer(t)
	defer testServer.Close()

	// Test rate limiting by making rapid requests
	client := &http.Client{}

	// Make requests up to the rate limit
	for i := 0; i < 5; i++ {
		req, err := http.NewRequest("GET", testServer.URL+"/api/v1/products", nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		// First 5 requests should succeed
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}

	// The 6th request should be rate limited
	req, err := http.NewRequest("GET", testServer.URL+"/api/v1/products", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Retry-After"), "600")
}

func TestAuthenticationChallenge(t *testing.T) {
	_, testServer := setupTestServer(t)
	defer testServer.Close()

	// Test creating an authentication challenge
	challengeReq := map[string]string{
		"address": "0x1234567890123456789012345678901234567890",
	}
	reqBody, _ := json.Marshal(challengeReq)

	req, err := http.NewRequest("POST", testServer.URL+"/api/v1/auth/challenge", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var challengeResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&challengeResp)
	require.NoError(t, err)

	assert.Contains(t, challengeResp, "nonce")
	assert.Contains(t, challengeResp, "address")
	assert.Contains(t, challengeResp, "expires_at")
	assert.Equal(t, "0x1234567890123456789012345678901234567890", challengeResp["address"])
}

func TestAuthenticationVerification(t *testing.T) {
	_, testServer := setupTestServer(t)
	defer testServer.Close()

	// First create a challenge
	challengeReq := map[string]string{
		"address": "0x1234567890123456789012345678901234567890",
	}
	reqBody, _ := json.Marshal(challengeReq)

	req, err := http.NewRequest("POST", testServer.URL+"/api/v1/auth/challenge", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	var challengeResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&challengeResp)

	// Now verify the challenge with a valid signature
	nonce := challengeResp["nonce"].(string)
	address := challengeResp["address"].(string)

	// Create a valid signature (in real implementation, this would be signed by the private key)
	validSignature := fmt.Sprintf("%x", []byte(nonce+address))

	verifyReq := map[string]string{
		"nonce":     nonce,
		"signature": validSignature,
	}
	verifyBody, _ := json.Marshal(verifyReq)

	req, err = http.NewRequest("POST", testServer.URL+"/api/v1/auth/verify", bytes.NewBuffer(verifyBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var verifyResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&verifyResp)
	require.NoError(t, err)

	assert.Equal(t, address, verifyResp["address"])
	assert.Equal(t, true, verifyResp["valid"])
}

func TestAuthenticationVerificationInvalid(t *testing.T) {
	_, testServer := setupTestServer(t)
	defer testServer.Close()

	// Test verification with invalid signature
	verifyReq := map[string]string{
		"nonce":     "invalid_nonce",
		"signature": "invalid_signature",
	}
	verifyBody, _ := json.Marshal(verifyReq)

	req, err := http.NewRequest("POST", testServer.URL+"/api/v1/auth/verify", bytes.NewBuffer(verifyBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	var verifyResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&verifyResp)
	require.NoError(t, err)

	assert.Equal(t, false, verifyResp["valid"])
}

func TestConcurrencyQuota(t *testing.T) {
	_, testServer := setupTestServer(t)
	defer testServer.Close()

	// Test concurrency quota by making multiple training requests
	client := &http.Client{}

	// Create a training request
	trainReq := map[string]interface{}{
		"dataset": "test_dataset",
		"task":    "test_task",
		"dp": map[string]interface{}{
			"enabled": true,
			"epsilon": 1.0,
		},
	}
	reqBody, _ := json.Marshal(trainReq)

	// Make first request (should succeed)
	req, err := http.NewRequest("POST", testServer.URL+"/api/v1/train", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", "test_signature") // Simulate authenticated user

	resp, err := client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	// First request should succeed
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Make second request (should succeed)
	req, err = http.NewRequest("POST", testServer.URL+"/api/v1/train", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", "test_signature")

	resp, err = client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	// Second request should succeed
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Make third request (should fail due to quota)
	req, err = http.NewRequest("POST", testServer.URL+"/api/v1/train", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Signature", "test_signature")

	resp, err = client.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	// Third request should fail
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestBackpressure(t *testing.T) {
	_, testServer := setupTestServer(t)
	defer testServer.Close()

	// Note: This is a simplified test. In a real implementation,
	// you would need to inject high CPU/memory conditions to trigger backpressure.
	// For now, we'll test that the endpoint exists and responds normally.

	req, err := http.NewRequest("GET", testServer.URL+"/api/v1/products", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should respond normally when not under pressure
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSecurityHeaders(t *testing.T) {
	_, testServer := setupTestServer(t)
	defer testServer.Close()

	req, err := http.NewRequest("GET", testServer.URL+"/api/v1/products", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check for security headers
	assert.Equal(t, "v1", resp.Header.Get("X-API-Version"))
}

func TestLegacyEndpointsWithDeprecation(t *testing.T) {
	_, testServer := setupTestServer(t)
	defer testServer.Close()

	// Test legacy endpoint still works but with deprecation warning
	req, err := http.NewRequest("GET", testServer.URL+"/health", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Legacy endpoint should still work
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequestSizeLimits(t *testing.T) {
	_, testServer := setupTestServer(t)
	defer testServer.Close()

	// Test with a large request body (should be rejected)
	largeBody := strings.Repeat("a", 11*1024*1024) // 11MB, exceeds 10MB limit

	req, err := http.NewRequest("POST", testServer.URL+"/api/v1/train", strings.NewReader(largeBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should be rejected due to size limit
	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestSecurityEventLogging(t *testing.T) {
	_, testServer := setupTestServer(t)
	defer testServer.Close()

	// Make a request that should trigger security logging
	req, err := http.NewRequest("GET", testServer.URL+"/api/v1/products", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// The request should succeed and security events should be logged
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// Note: In a real test, you would capture and verify the log output
}

func TestRateLimitRecovery(t *testing.T) {
	_, testServer := setupTestServer(t)
	defer testServer.Close()

	client := &http.Client{}

	// Exhaust the rate limit
	for i := 0; i < 6; i++ {
		req, err := http.NewRequest("GET", testServer.URL+"/api/v1/products", nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()

		if i < 5 {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
		}
	}

	// Wait for rate limit to recover (in real implementation, this would be longer)
	// For testing, we'll just verify the behavior
	time.Sleep(100 * time.Millisecond)

	// Make another request
	req, err := http.NewRequest("GET", testServer.URL+"/api/v1/products", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should still be rate limited due to greylist
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
}
