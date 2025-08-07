package api

import (
	"crypto/md5"
	"fmt"
	"os"
)

// This file contains intentional security issues for testing gosec
// DO NOT USE IN PRODUCTION - FOR TESTING ONLY

func demoSecurityIssues() {
	// SECURITY ISSUE: Hardcoded password (gosec G101)
	password := "admin123"
	
	// SECURITY ISSUE: Hardcoded API key (gosec G101)
	apiKey := "sk-1234567890abcdef"
	
	// SECURITY ISSUE: Weak crypto algorithm (gosec G401)
	hash := md5.Sum([]byte(password))
	
	// SECURITY ISSUE: Potential command injection (gosec G204)
	cmd := fmt.Sprintf("echo %s", password)
	
	// SECURITY ISSUE: Potential path traversal (gosec G304)
	filePath := "/tmp/" + password + ".txt"
	
	// This would normally be a security issue, but we're just demonstrating
	_ = hash
	_ = cmd
	_ = filePath
	_ = apiKey
}

func demoWeakRandom() {
	// SECURITY ISSUE: Weak random number generation (gosec G404)
	// This is just for demonstration - don't use in real code
	weakRandom := os.Getpid() // Not cryptographically secure
	_ = weakRandom
} 