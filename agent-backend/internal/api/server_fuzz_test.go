//go:build go1.18
// +build go1.18

package api

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"log/slog"
	"pandacea/agent-backend/internal/p2p"
	"pandacea/agent-backend/internal/policy"
)

func FuzzHandleCreateLease(f *testing.F) {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	testConfig := createTestServerConfig()
	policyEngine, _ := policy.NewEngine(logger, testConfig)
	mockP2PNode := &p2p.Node{}
	server := NewServer(policyEngine, logger, mockP2PNode)

	// Seed with a few valid LeaseRequest payloads
	validPayloads := [][]byte{
		[]byte(`{"productId":"did:pandacea:earner:123/abc-456","maxPrice":"0.01","duration":"24h"}`),
		[]byte(`{"productId":"did:pandacea:earner:999/xyz-789","maxPrice":"1.00","duration":"12h"}`),
		[]byte(`{"productId":"did:pandacea:earner:555/def-321","maxPrice":"0.005","duration":"30m"}`),
	}
	for _, p := range validPayloads {
		f.Add(p)
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		req := httptest.NewRequest("POST", "/api/v1/leases", bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Defensive: If the data is not valid JSON, that's fine
		// The handler should never panic
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("handleCreateLease panicked with input: %q, panic: %v", string(data), r)
			}
		}()

		server.handleCreateLease(w, req)
		// Optionally, check that the response is a valid HTTP response
		_ = w.Result()
	})
}
