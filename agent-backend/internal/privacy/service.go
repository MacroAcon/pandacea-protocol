package privacy

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"pandacea/agent-backend/internal/contracts"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// PrivacyService defines the interface for privacy-preserving computations
type PrivacyService interface {
	ExecuteComputation(ctx context.Context, req *ComputationRequest) (*ComputationResponse, error)
	GetComputationResult(ctx context.Context, computationID string) (*ComputationResult, error)
	VerifyLease(ctx context.Context, leaseID string, spenderAddr string) error
	Start() error
	Stop() error
}

// privacyService implements the PrivacyService interface
type privacyService struct {
	logger          *slog.Logger
	ethClient       *ethclient.Client
	contractAddress common.Address
	contract        *contracts.LeaseAgreement
	dataDir         string

	// IPFS client configuration
	ipfsAPIURL string
	httpClient *http.Client

	// Asynchronous job management
	jobs      map[string]*ComputationJob
	jobsMutex sync.RWMutex

	// Container pool
	containerPool chan *DockerContainer
	poolSize      int
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// ComputationJob represents an asynchronous computation job
type ComputationJob struct {
	ID        string              `json:"id"`
	Status    string              `json:"status"` // "pending", "completed", "failed"
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	Request   *ComputationRequest `json:"request,omitempty"`
	Results   *ComputationResults `json:"results,omitempty"`
	Error     string              `json:"error,omitempty"`
}

// ComputationResult represents the result of a computation job
type ComputationResult struct {
	Status  string              `json:"status"`
	Results *ComputationResults `json:"results,omitempty"`
	Error   string              `json:"error,omitempty"`
}

// DockerContainer represents a container in the pool
type DockerContainer struct {
	ID       string
	IsActive bool
}

// ComputationRequest represents a request to execute privacy-preserving computation
type ComputationRequest struct {
	LeaseID        string      `json:"lease_id"`
	ComputationCid string      `json:"computationCid"` // IPFS Content ID pointing to the computation script
	Inputs         []DataInput `json:"inputs"`
}

// DataInput represents a data asset input for computation
type DataInput struct {
	AssetID      string `json:"asset_id"`
	VariableName string `json:"variable_name"`
}

// ComputationResponse represents the response for starting a computation
type ComputationResponse struct {
	ComputationID string `json:"computation_id"`
}

// ComputationResults contains the output and artifacts from computation
type ComputationResults struct {
	Output    string            `json:"output"`
	Artifacts map[string]string `json:"artifacts"`
}

// NewPrivacyService creates a new PrivacyService instance
func NewPrivacyService(
	logger *slog.Logger,
	ethClient *ethclient.Client,
	contractAddress common.Address,
	dataDir string,
	poolSize int,
	ipfsAPIURL string,
) (PrivacyService, error) {
	if poolSize <= 0 {
		poolSize = 3 // Default pool size
	}

	contract, err := contracts.NewLeaseAgreement(contractAddress, ethClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract instance: %w", err)
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Set default IPFS API URL if not provided
	if ipfsAPIURL == "" {
		ipfsAPIURL = "http://127.0.0.1:5001"
	}

	service := &privacyService{
		logger:          logger,
		ethClient:       ethClient,
		contractAddress: contractAddress,
		contract:        contract,
		dataDir:         dataDir,
		ipfsAPIURL:      ipfsAPIURL,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
		jobs:            make(map[string]*ComputationJob),
		containerPool:   make(chan *DockerContainer, poolSize),
		poolSize:        poolSize,
		stopChan:        make(chan struct{}),
	}

	return service, nil
}

// Start initializes the container pool and starts background workers
func (ps *privacyService) Start() error {
	ps.logger.Info("starting privacy service", "pool_size", ps.poolSize)

	// Initialize container pool
	for i := 0; i < ps.poolSize; i++ {
		container, err := ps.createContainer()
		if err != nil {
			ps.logger.Error("failed to create container for pool", "error", err, "index", i)
			continue
		}
		ps.containerPool <- container
	}

	ps.logger.Info("privacy service started successfully", "containers_initialized", len(ps.containerPool))
	return nil
}

// Stop gracefully shuts down the privacy service
func (ps *privacyService) Stop() error {
	ps.logger.Info("stopping privacy service")

	// Signal all workers to stop
	close(ps.stopChan)

	// Wait for all workers to finish
	ps.wg.Wait()

	// Clean up containers
	close(ps.containerPool)
	for container := range ps.containerPool {
		ps.destroyContainer(container)
	}

	ps.logger.Info("privacy service stopped")
	return nil
}

// ExecuteComputation starts an asynchronous computation job
func (ps *privacyService) ExecuteComputation(ctx context.Context, req *ComputationRequest) (*ComputationResponse, error) {
	ps.logger.Info("starting asynchronous computation",
		"lease_id", req.LeaseID,
		"inputs_count", len(req.Inputs))

	// Validate request
	if err := ps.validateComputationRequest(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Generate unique computation ID
	computationID := ps.generateComputationID()

	// Create job record
	job := &ComputationJob{
		ID:        computationID,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Request:   req,
	}

	// Store job in memory
	ps.jobsMutex.Lock()
	ps.jobs[computationID] = job
	ps.jobsMutex.Unlock()

	// Start asynchronous execution
	ps.wg.Add(1)
	go ps.executeJobAsync(computationID, req)

	return &ComputationResponse{
		ComputationID: computationID,
	}, nil
}

// GetComputationResult retrieves the result of a computation job
func (ps *privacyService) GetComputationResult(ctx context.Context, computationID string) (*ComputationResult, error) {
	ps.jobsMutex.RLock()
	job, exists := ps.jobs[computationID]
	ps.jobsMutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("computation job not found: %s", computationID)
	}

	result := &ComputationResult{
		Status: job.Status,
	}

	if job.Status == "completed" {
		result.Results = job.Results
	} else if job.Status == "failed" {
		result.Error = job.Error
	}

	return result, nil
}

// executeJobAsync executes a computation job asynchronously
func (ps *privacyService) executeJobAsync(computationID string, req *ComputationRequest) {
	defer ps.wg.Done()

	ps.logger.Info("starting async job execution", "computation_id", computationID)

	// Acquire container from pool
	container := ps.acquireContainer()
	if container == nil {
		ps.updateJobStatus(computationID, "failed", nil, "failed to acquire container from pool")
		return
	}
	defer ps.releaseContainer(container)

	// Create temporary directory for this computation
	tempDir, err := os.MkdirTemp("", "pandacea-computation-*")
	if err != nil {
		ps.updateJobStatus(computationID, "failed", nil, fmt.Sprintf("failed to create temp directory: %v", err))
		return
	}
	defer os.RemoveAll(tempDir)

	// Fetch computation script from IPFS
	computationCode, err := ps.fetchContentFromIPFS(context.Background(), req.ComputationCid)
	if err != nil {
		ps.updateJobStatus(computationID, "failed", nil, fmt.Sprintf("failed to fetch computation script from IPFS: %v", err))
		return
	}

	// Create Python script file
	scriptPath := filepath.Join(tempDir, "computation.py")
	if err := os.WriteFile(scriptPath, []byte(computationCode), 0644); err != nil {
		ps.updateJobStatus(computationID, "failed", nil, fmt.Sprintf("failed to write script file: %v", err))
		return
	}

	// Create data loading script
	dataLoaderPath := filepath.Join(tempDir, "data_loader.py")
	if err := ps.createDataLoader(dataLoaderPath, req.Inputs); err != nil {
		ps.updateJobStatus(computationID, "failed", nil, fmt.Sprintf("failed to create data loader: %v", err))
		return
	}

	// Create PySyft Datasite script
	datasiteScript := ps.createDatasiteScript(req.Inputs)
	datasitePath := filepath.Join(tempDir, "datasite.py")
	if err := os.WriteFile(datasitePath, []byte(datasiteScript), 0644); err != nil {
		ps.updateJobStatus(computationID, "failed", nil, fmt.Sprintf("failed to write datasite script: %v", err))
		return
	}

	// Execute the computation in the container
	output, artifacts, err := ps.executeInContainer(container, tempDir, scriptPath)
	if err != nil {
		ps.updateJobStatus(computationID, "failed", nil, fmt.Sprintf("execution error: %v", err))
		return
	}

	// Encode artifacts as base64
	encodedArtifacts := make(map[string]string)
	for filename, data := range artifacts {
		encodedArtifacts[filename] = base64.StdEncoding.EncodeToString(data)
	}

	// Update job status to completed
	results := &ComputationResults{
		Output:    output,
		Artifacts: encodedArtifacts,
	}
	ps.updateJobStatus(computationID, "completed", results, "")

	ps.logger.Info("async job execution completed", "computation_id", computationID)
}

// updateJobStatus updates the status of a computation job
func (ps *privacyService) updateJobStatus(computationID, status string, results *ComputationResults, errorMsg string) {
	ps.jobsMutex.Lock()
	defer ps.jobsMutex.Unlock()

	if job, exists := ps.jobs[computationID]; exists {
		job.Status = status
		job.UpdatedAt = time.Now()
		if results != nil {
			job.Results = results
		}
		if errorMsg != "" {
			job.Error = errorMsg
		}
	}

	ps.logger.Info("job status updated", "computation_id", computationID, "status", status)
}

// acquireContainer acquires a container from the pool
func (ps *privacyService) acquireContainer() *DockerContainer {
	select {
	case container := <-ps.containerPool:
		return container
	case <-time.After(30 * time.Second):
		ps.logger.Error("timeout waiting for container from pool")
		return nil
	}
}

// releaseContainer returns a container to the pool
func (ps *privacyService) releaseContainer(container *DockerContainer) {
	// Clean the container before returning to pool
	if err := ps.cleanContainer(container); err != nil {
		ps.logger.Error("failed to clean container", "container_id", container.ID, "error", err)
		// Destroy and recreate the container
		ps.destroyContainer(container)
		newContainer, err := ps.createContainer()
		if err != nil {
			ps.logger.Error("failed to create replacement container", "error", err)
			return
		}
		container = newContainer
	}

	select {
	case ps.containerPool <- container:
		// Container returned to pool successfully
	default:
		// Pool is full, destroy the container
		ps.destroyContainer(container)
	}
}

// createContainer creates a new Docker container
func (ps *privacyService) createContainer() (*DockerContainer, error) {
	// Create a new PySyft container
	cmd := exec.Command("docker", "run", "-d",
		"--network", "none",
		"--memory", "512m",
		"--cpus", "1",
		"pandacea/pysyft-datasite:latest",
		"tail", "-f", "/dev/null") // Keep container running

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w, output: %s", err, string(output))
	}

	containerID := strings.TrimSpace(string(output))
	ps.logger.Info("created container", "container_id", containerID)

	return &DockerContainer{
		ID:       containerID,
		IsActive: true,
	}, nil
}

// destroyContainer destroys a Docker container
func (ps *privacyService) destroyContainer(container *DockerContainer) {
	if container == nil || !container.IsActive {
		return
	}

	cmd := exec.Command("docker", "rm", "-f", container.ID)
	if err := cmd.Run(); err != nil {
		ps.logger.Error("failed to destroy container", "container_id", container.ID, "error", err)
	} else {
		ps.logger.Info("destroyed container", "container_id", container.ID)
	}

	container.IsActive = false
}

// cleanContainer cleans a container for reuse
func (ps *privacyService) cleanContainer(container *DockerContainer) error {
	if container == nil || !container.IsActive {
		return fmt.Errorf("container is not active")
	}

	// Clean the workspace directory
	cmd := exec.Command("docker", "exec", container.ID, "rm", "-rf", "/workspace/*")
	return cmd.Run()
}

// executeInContainer executes computation in a specific container
func (ps *privacyService) executeInContainer(container *DockerContainer, tempDir, scriptPath string) (string, map[string][]byte, error) {
	// Copy files to container
	if err := ps.copyToContainer(container.ID, tempDir, "/workspace"); err != nil {
		return "", nil, fmt.Errorf("failed to copy files to container: %w", err)
	}

	// Copy data directory to container
	if err := ps.copyToContainer(container.ID, ps.dataDir, "/data"); err != nil {
		return "", nil, fmt.Errorf("failed to copy data to container: %w", err)
	}

	// Execute the computation
	cmd := exec.Command("docker", "exec", container.ID, "python", "/workspace/datasite.py")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), nil, fmt.Errorf("container execution failed: %w", err)
	}

	// Collect artifacts
	artifacts := make(map[string][]byte)
	artifactDir := filepath.Join(tempDir, "artifacts")
	if _, err := os.Stat(artifactDir); err == nil {
		files, err := os.ReadDir(artifactDir)
		if err == nil {
			for _, file := range files {
				if !file.IsDir() {
					filePath := filepath.Join(artifactDir, file.Name())
					data, err := os.ReadFile(filePath)
					if err == nil {
						artifacts[file.Name()] = data
					}
				}
			}
		}
	}

	return string(output), artifacts, nil
}

// copyToContainer copies files from host to container
func (ps *privacyService) copyToContainer(containerID, srcPath, destPath string) error {
	cmd := exec.Command("docker", "cp", srcPath, containerID+":"+destPath)
	return cmd.Run()
}

// generateComputationID generates a unique computation ID
func (ps *privacyService) generateComputationID() string {
	return fmt.Sprintf("comp-%d", time.Now().UnixNano())
}

// VerifyLease verifies that a lease is valid and active
func (ps *privacyService) VerifyLease(ctx context.Context, leaseID string, spenderAddr string) error {
	// Convert lease ID to bytes32
	if !strings.HasPrefix(leaseID, "0x") {
		leaseID = "0x" + leaseID
	}

	leaseIDBytes := common.FromHex(leaseID)
	if len(leaseIDBytes) != 32 {
		return fmt.Errorf("invalid lease ID format")
	}

	var leaseIDArray [32]byte
	copy(leaseIDArray[:], leaseIDBytes)

	// Check if lease exists
	exists, err := ps.contract.LeaseExists(nil, leaseIDArray)
	if err != nil {
		return fmt.Errorf("failed to check lease existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("lease does not exist")
	}

	// Get lease details
	lease, err := ps.contract.GetLease(nil, leaseIDArray)
	if err != nil {
		return fmt.Errorf("failed to get lease details: %w", err)
	}

	// Verify lease is approved
	if !lease.IsApproved {
		return fmt.Errorf("lease is not approved")
	}

	// Verify lease is not executed
	if lease.IsExecuted {
		return fmt.Errorf("lease has already been executed")
	}

	// Verify lease is not disputed
	if lease.IsDisputed {
		return fmt.Errorf("lease is disputed")
	}

	// Verify spender address matches
	if !strings.EqualFold(lease.Spender.Hex(), spenderAddr) {
		return fmt.Errorf("spender address mismatch")
	}

	return nil
}

// validateComputationRequest validates the computation request
func (ps *privacyService) validateComputationRequest(req *ComputationRequest) error {
	if req.LeaseID == "" {
		return fmt.Errorf("lease_id is required")
	}

	if req.ComputationCid == "" {
		return fmt.Errorf("computationCid is required")
	}

	// Basic CID validation
	if len(req.ComputationCid) != 46 || req.ComputationCid[0] != 'Q' { // IPFS CID is 46 characters long and starts with 'Q'
		return fmt.Errorf("invalid IPFS CID format")
	}

	if len(req.Inputs) == 0 {
		return fmt.Errorf("at least one input is required")
	}

	for _, input := range req.Inputs {
		if input.AssetID == "" {
			return fmt.Errorf("asset_id is required for all inputs")
		}
		if input.VariableName == "" {
			return fmt.Errorf("variable_name is required for all inputs")
		}
	}

	return nil
}

// createDataLoader creates a Python script to load data assets
func (ps *privacyService) createDataLoader(scriptPath string, inputs []DataInput) error {
	var dataLoaderCode strings.Builder
	dataLoaderCode.WriteString("import pandas as pd\n")
	dataLoaderCode.WriteString("import os\n\n")

	for _, input := range inputs {
		dataLoaderCode.WriteString(fmt.Sprintf("# Load %s\n", input.AssetID))
		dataLoaderCode.WriteString(fmt.Sprintf("data_path = os.path.join('/data', '%s.csv')\n", input.AssetID))
		dataLoaderCode.WriteString(fmt.Sprintf("if os.path.exists(data_path):\n"))
		dataLoaderCode.WriteString(fmt.Sprintf("    %s = pd.read_csv(data_path)\n", input.VariableName))
		dataLoaderCode.WriteString(fmt.Sprintf("else:\n"))
		dataLoaderCode.WriteString(fmt.Sprintf("    raise FileNotFoundError(f'Data asset {input.AssetID} not found')\n\n"))
	}

	return os.WriteFile(scriptPath, []byte(dataLoaderCode.String()), 0644)
}

// createDatasiteScript creates a PySyft Datasite script
func (ps *privacyService) createDatasiteScript(inputs []DataInput) string {
	var script strings.Builder

	script.WriteString(`import syft as sy
import torch
import pandas as pd
import os
import sys

# Initialize PySyft
sy.load("pandas")
sy.load("torch")

# Create a virtual machine (Datasite)
vm = sy.VirtualMachine(name="pandacea-datasite")

# Load data assets
`)

	for _, input := range inputs {
		script.WriteString(fmt.Sprintf("data_path = os.path.join('/data', '%s.csv')\n", input.AssetID))
		script.WriteString(fmt.Sprintf("if os.path.exists(data_path):\n"))
		script.WriteString(fmt.Sprintf("    %s = pd.read_csv(data_path)\n", input.VariableName))
		script.WriteString(fmt.Sprintf("    # Convert to PySyft tensor if needed\n"))
		script.WriteString(fmt.Sprintf("    if isinstance(%s, pd.DataFrame):\n", input.VariableName))
		script.WriteString(fmt.Sprintf("        %s = torch.tensor(%s.values, dtype=torch.float32)\n", input.VariableName, input.VariableName))
		script.WriteString(fmt.Sprintf("    %s = %s.send(vm)\n", input.VariableName, input.VariableName))
		script.WriteString(fmt.Sprintf("else:\n"))
		script.WriteString(fmt.Sprintf("    raise FileNotFoundError(f'Data asset {input.AssetID} not found')\n\n"))
	}

	script.WriteString(`# Execute the computation
exec(open('/workspace/computation.py').read())

# Save any artifacts
if 'model' in locals():
    torch.save(model.state_dict(), '/workspace/model_weights.pth')

print("Computation completed successfully")
`)

	return script.String()
}

// fetchContentFromIPFS fetches content from IPFS using the provided CID
func (ps *privacyService) fetchContentFromIPFS(ctx context.Context, cid string) (string, error) {
	// Construct the IPFS API URL for cat operation
	url := fmt.Sprintf("%s/api/v0/cat?arg=%s", ps.ipfsAPIURL, cid)

	ps.logger.Info("fetching content from IPFS", "cid", cid, "url", url)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create IPFS request: %w", err)
	}

	// Make the request
	resp, err := ps.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch content from IPFS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("IPFS API returned status %d", resp.StatusCode)
	}

	// Read the content
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read IPFS content: %w", err)
	}

	// Validate content size (max 1MB)
	if len(content) > 1024*1024 {
		return "", fmt.Errorf("IPFS content too large: %d bytes (max 1MB)", len(content))
	}

	ps.logger.Info("successfully fetched content from IPFS", "cid", cid, "size", len(content))
	return string(content), nil
}
