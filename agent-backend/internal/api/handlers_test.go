package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"pandacea/agent-backend/internal/policy"
	"pandacea/agent-backend/internal/privacy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockPrivacyService implements privacy.PrivacyService for testing
type MockPrivacyService struct{}

func (m *MockPrivacyService) ExecuteComputation(ctx context.Context, req *privacy.ComputationRequest) (*privacy.ComputationResponse, error) {
	return &privacy.ComputationResponse{
		ComputationID: "mock-computation-123",
	}, nil
}

func (m *MockPrivacyService) GetComputationResult(ctx context.Context, computationID string) (*privacy.ComputationResult, error) {
	return &privacy.ComputationResult{
		Status: "completed",
		Results: &privacy.ComputationResults{
			Output: "mock output",
			Artifacts: map[string]string{
				"result.json": "mock artifact",
			},
		},
	}, nil
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

// setupTestServer creates a test server with mock dependencies
func setupTestServer(t *testing.T) *Server {
	policyEngine := &policy.Engine{}
	privacyService := &MockPrivacyService{}

	server := NewServer(policyEngine, nil, nil, privacyService, nil)

	// Set MOCK_DP environment variable for testing
	os.Setenv("MOCK_DP", "1")

	return server
}

// TestAPIVersionHeader tests that API v1 endpoints set the correct version header
func TestAPIVersionHeader(t *testing.T) {
	server := setupTestServer(t)

	// Test /api/v1/train endpoint
	req := httptest.NewRequest("POST", "/api/v1/train", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	// Check that the version header is set
	assert.Equal(t, "v1", w.Header().Get("X-API-Version"))
}

// TestTrainEndpoint tests the /api/v1/train endpoint
func TestTrainEndpoint(t *testing.T) {
	server := setupTestServer(t)

	// Create a valid train request
	trainReq := TrainRequest{
		Dataset: "test_dataset",
		Task:    "classification",
		DP: struct {
			Enabled bool    `json:"enabled"`
			Epsilon float64 `json:"epsilon"`
		}{
			Enabled: true,
			Epsilon: 2.0,
		},
	}

	reqBody, err := json.Marshal(trainReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/train", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "v1", w.Header().Get("X-API-Version"))

	// Parse response
	var response TrainResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check that job ID is returned
	assert.NotEmpty(t, response.JobID)

	// Verify job was created
	server.jobsMutex.RLock()
	job, exists := server.jobs[response.JobID]
	server.jobsMutex.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, "pending", job.Status)
	assert.Equal(t, trainReq.Dataset, job.Dataset)
	assert.Equal(t, trainReq.Task, job.Task)
	assert.Equal(t, trainReq.DP.Epsilon, job.Epsilon)
}

// TestAggregateEndpoint tests the /api/v1/aggregate/:id endpoint
func TestAggregateEndpoint(t *testing.T) {
	server := setupTestServer(t)

	// Create a test job
	jobID := "test-job-123"
	job := &TrainingJob{
		JobID:        jobID,
		Status:       "complete",
		Dataset:      "test_dataset",
		Task:         "classification",
		Epsilon:      2.0,
		ArtifactPath: "./data/products/test-job-123/aggregate.json",
		CreatedAt:    time.Now(),
		CompletedAt:  &time.Time{},
	}

	server.jobsMutex.Lock()
	server.jobs[jobID] = job
	server.jobsMutex.Unlock()

	// Create the output directory and file
	outputDir := "./data/products/test-job-123"
	err := os.MkdirAll(outputDir, 0755)
	require.NoError(t, err)

	// Create a mock aggregate result
	aggregateResult := map[string]interface{}{
		"job_id":                jobID,
		"dataset":               "test_dataset",
		"task":                  "classification",
		"epsilon_used":          2.0,
		"model_accuracy":        0.85,
		"samples_processed":     1000,
		"training_time_seconds": 30.5,
		"dp_noise_scale":        0.5,
		"timestamp":             time.Now().Format(time.RFC3339),
	}

	resultBytes, err := json.MarshalIndent(aggregateResult, "", "  ")
	require.NoError(t, err)

	err = os.WriteFile(job.ArtifactPath, resultBytes, 0644)
	require.NoError(t, err)

	// Test the aggregate endpoint
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/aggregate/%s", jobID), nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "v1", w.Header().Get("X-API-Version"))

	// Parse response and validate schema
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check required fields exist
	assert.Contains(t, response, "job_id")
	assert.Contains(t, response, "dataset")
	assert.Contains(t, response, "task")
	assert.Contains(t, response, "epsilon_used")
	assert.Contains(t, response, "model_accuracy")
	assert.Contains(t, response, "samples_processed")
	assert.Contains(t, response, "training_time_seconds")
	assert.Contains(t, response, "dp_noise_scale")
	assert.Contains(t, response, "timestamp")

	// Clean up
	os.RemoveAll(outputDir)
}

// TestAggregateEndpointNotFound tests the aggregate endpoint for non-existent jobs
func TestAggregateEndpointNotFound(t *testing.T) {
	server := setupTestServer(t)

	req := httptest.NewRequest("GET", "/api/v1/aggregate/non-existent-job", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestTrainEndpointInvalidRequest tests the train endpoint with invalid requests
func TestTrainEndpointInvalidRequest(t *testing.T) {
	server := setupTestServer(t)

	testCases := []struct {
		name         string
		request      TrainRequest
		expectedCode int
	}{
		{
			name: "empty dataset",
			request: TrainRequest{
				Dataset: "",
				Task:    "classification",
				DP: struct {
					Enabled bool    `json:"enabled"`
					Epsilon float64 `json:"epsilon"`
				}{
					Enabled: true,
					Epsilon: 2.0,
				},
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "empty task",
			request: TrainRequest{
				Dataset: "test_dataset",
				Task:    "",
				DP: struct {
					Enabled bool    `json:"enabled"`
					Epsilon float64 `json:"epsilon"`
				}{
					Enabled: true,
					Epsilon: 2.0,
				},
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name: "negative epsilon",
			request: TrainRequest{
				Dataset: "test_dataset",
				Task:    "classification",
				DP: struct {
					Enabled bool    `json:"enabled"`
					Epsilon float64 `json:"epsilon"`
				}{
					Enabled: true,
					Epsilon: -1.0,
				},
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody, err := json.Marshal(tc.request)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/v1/train", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedCode, w.Code)
		})
	}
}

// TestLegacyEndpoints tests that legacy unversioned endpoints still work
func TestLegacyEndpoints(t *testing.T) {
	server := setupTestServer(t)

	// Test legacy /train endpoint
	trainReq := TrainRequest{
		Dataset: "test_dataset",
		Task:    "classification",
		DP: struct {
			Enabled bool    `json:"enabled"`
			Epsilon float64 `json:"epsilon"`
		}{
			Enabled: true,
			Epsilon: 2.0,
		},
	}

	reqBody, err := json.Marshal(trainReq)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/train", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	// Legacy endpoint should still work but may log deprecation warning
	assert.Equal(t, http.StatusCreated, w.Code)

	// Parse response
	var response TrainResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.JobID)
}

// TestHealthEndpoint tests the health check endpoint
func TestHealthEndpoint(t *testing.T) {
	server := setupTestServer(t)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
}

func TestHealthzAndReadyz(t *testing.T) {
	server := setupTestServer(t)

	// /healthz should always succeed
	req := httptest.NewRequest("GET", "/healthz", nil)
	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// /readyz returns 200 or 503 based on environment, but JSON must include keys
	req = httptest.NewRequest("GET", "/readyz", nil)
	w = httptest.NewRecorder()
	server.router.ServeHTTP(w, req)
	assert.Contains(t, []int{http.StatusOK, http.StatusServiceUnavailable}, w.Code)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &payload))
	_, ok := payload["checks"]
	assert.True(t, ok)
}
