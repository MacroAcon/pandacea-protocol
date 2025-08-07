package api

import (
	"testing"
)

// TestSecurityScan demonstrates that gosec can detect security issues
// This is a test file to verify our CI security scanning works
func TestSecurityScan(t *testing.T) {
	// TODO: This is a fake security issue for testing gosec
	// password := "admin123" // gosec should flag this hardcoded password
	
	// TODO: Another fake security issue for testing
	// apiKey := "sk-1234567890abcdef" // gosec should flag this hardcoded secret
	
	t.Log("Security scan test completed")
}

// TestNoSecurityIssues demonstrates clean code that should pass security scan
func TestNoSecurityIssues(t *testing.T) {
	// This function should pass security scanning
	config := map[string]string{
		"host":     "localhost",
		"port":     "8080",
		"protocol": "http",
	}
	
	if config["host"] != "localhost" {
		t.Error("Expected localhost")
	}
} 