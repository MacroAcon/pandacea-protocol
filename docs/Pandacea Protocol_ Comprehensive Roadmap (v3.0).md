# **Pandacea MVP: Comprehensive Roadmap & Technical Plan (v3.0)**

Document Version: 3.0  
Date: July 20, 2025

### **Overview**

The current MVP has successfully proven the core off-chain discovery loop and established a secure authentication layer. This document provides a comprehensive technical roadmap to evolve the MVP from its current state into a feature-complete, robust, and scalable system ready for public testing.

**Version 3.0 Update:** This roadmap has been restructured based on a detailed project critique to de-risk development. The previously monolithic "Path 2" has been broken down into three distinct, sequential sprints (Sprints 2, 3, and 4\) to ensure a manageable scope for each development cycle. This version also elevates the priority of testing the core economic loop and explicitly adds end-to-end testing tasks for all new major features.

### **Path 0: MVP Bootstrap & Hardening (Completed)**

**Objective:** This path documents the foundational work completed to establish the initial MVP, including core infrastructure, component scaffolding, and critical security enhancements.

#### **1\. Initial Infrastructure & Scaffolding**

* **Status:** ✅ **Completed**  
* **Details:** Set up repositories, CI/CD pipelines (linting, testing), and a containerized local development environment using Docker Compose for all core components (Smart Contracts, Agent Backend, Builder SDK).

#### **2\. Core Component Implementation**

* **Status:** ✅ **Completed**  
* **Details:** Implemented the initial versions of the LeaseAgreement.sol smart contract, the Go-based agent backend with P2P discovery, and the Python builder SDK.

#### **3\. Persistent Agent Identity**

* **Status:** ✅ **Completed**  
* **Details:** The agent backend was updated to generate and load a persistent cryptographic key from a local file (\~/.pandacea/agent.key), giving each agent a stable Peer ID on the network.

#### **4\. Dynamic Data Product Listings**

* **Status:** ✅ **Completed**  
* **Details:** Decoupled the agent's data offerings from its code by implementing logic to load and serve data products from an external products.json file.

#### **5\. Critical Security Fix: API Authentication**

* **Status:** ✅ **Completed**  
* **Details:** Patched a critical vulnerability by implementing cryptographic signing on all agent-to-agent API requests. The Go agent now verifies signatures, and the Python SDK signs all outgoing requests, ensuring all communication is authenticated.

### **Path 1: Hardening the MVP (Sprint 1\)**

**Objective:** To enhance the existing system's resilience and testability before adding major new features.

#### **1\. Enhance Integration Testing for Economic Rules**

* **Status:** 🟩 **To Do**  
* **Objective:** Create an automated end-to-end test that verifies the protocol's economic rules (specifically Dynamic Minimum Pricing) are correctly enforced by the agent.  
* **Technical Implementation (integration & builder-sdk):**  
  1. **SDK Enhancement:** Ensure the request\_lease() method in the PandaceaClient can handle and parse API error responses gracefully.  
  2. **New Integration Test:** In integration/test\_integration.py, add a new test case, test\_error\_path\_lease\_below\_min\_price.  
  3. **Test Logic:** This test will use the SDK to send a lease request with a maxPrice known to be below the agent's configured min\_price (e.g., "0.0001").  
  4. **Assertions:** The test must use pytest.raises(APIResponseError) to assert that the call fails as expected. It should then inspect the caught exception to verify that the status\_code is 403 and the response text contains the expected reason, "Proposed maxPrice is below the dynamic minimum price.".

### **Path 2: On-Chain Integration (Sprint 2\)**

**Objective:** To enable the agent and SDK to interact directly with the LeaseAgreement.sol smart contract on the local Anvil testnet.

#### **1\. Implement On-Chain Lease Creation**

* **Status:** 🟩 **To Do**  
* **Objective:** Enable a Spender (via the SDK) to create a lease on the blockchain.  
* **Technical Implementation:**  
  1. **Contract ABI Generation:** Add a script to the contracts directory that uses abigen to generate Go bindings for the LeaseAgreement contract. Commit the generated LeaseAgreement.go file to the agent-backend repository.  
  2. **Configuration:** Add the deployed LeaseAgreement contract address to a shared configuration file or environment variable accessible by both the agent and the SDK.  
  3. **SDK Wallet (builder-sdk):** Use the web3.py library. In a new lease.execute() function, load a private key from an environment variable (SPENDER\_PRIVATE\_KEY), then build, sign, and send a transaction to the createLease function on the smart contract, including the correct payment amount.

#### **2\. Create E2E Test for On-Chain Lease Creation**

* **Status:** 🟩 **To Do**  
* **Objective:** Write an automated test to verify the full on-chain lease creation flow.  
* **Task:** Create a new integration test that uses the SDK's lease.execute() function and verifies (by querying the local Anvil node) that the LeaseCreated event was successfully emitted with the correct parameters.

### **Path 3: Lease State Machine (Sprint 3\)**

**Objective:** To build out the complete, asynchronous negotiation flow, allowing the agent to react to on-chain events.

#### **1\. Implement Agent's On-Chain Event Listener**

* **Status:** 🟩 **To Do**  
* **Objective:** Build the agent's ability to listen for and react to blockchain events.  
* **Technical Implementation (agent-backend):**  
  1. **State Management:** Implement an in-memory map within the Server struct (e.g., pendingLeases map\[string\]\*LeaseProposal) to track the state of lease negotiations.  
  2. **Event Subscription:** Use the go-ethereum ethclient.SubscribeFilterLogs function to listen for LeaseCreated events from the LeaseAgreement contract.  
  3. **State Transition:** When the agent's event listener receives a LeaseCreated event, it should find the corresponding proposal in its pendingLeases map and update its status from "pending" to "approved".  
  4. **API Endpoint:** Implement the GET /api/v1/leases/{leaseProposalId} endpoint. This handler will look up the ID in the pendingLeases map and return the current status, allowing the SDK to poll and know when the lease has been successfully created on-chain.

#### **2\. Create E2E Test for Lease State Machine**

* **Status:** 🟩 **To Do**  
* **Objective:** Write an automated test to verify the full asynchronous lease approval flow.  
* **Task:** Create a new integration test that: 1\) Creates a lease on-chain using the SDK. 2\) Polls the new GET /api/v1/leases/{id} endpoint until the status changes to "approved". 3\) Fails if the status does not update within a reasonable timeout.

### **Path 4: Privacy & Core Economics (Sprint 4\)**

**Objective:** To implement the "trust-by-proof" privacy model using PySyft and to deploy the MVP of the core economic contracts.

#### **1\. Privacy Layer Integration (PySyft)**

* **Status:** 🟩 **To Do**  
* **Objective:** Implement federated analytics as outlined in the SDD.  
* **Technical Implementation:**  
  1. **Environment Setup:** Add an official openmined/syft-datasite container to the docker-compose.yml file.  
  2. **Agent as Data Owner (agent-backend):** On startup, the agent will use API calls to the PySyft Datasite to log in, create a dataset, and upload its pre-processed, non-sensitive feature data as a private "asset."  
  3. **SDK as Data Scientist (builder-sdk):** Implement a new request\_computation method in the SDK to define a simple function and submit it as a code request to the Datasite.  
  4. **Agent as Approver (agent-backend):** Update the lease endpoint to handle computation requests. When a valid request is received, the agent will find and approve the corresponding code request in the PySyft Datasite.

#### **2\. Implement MVP Reputation & Royalty Contracts**

* **Status:** 🟩 **To Do**  
* **Objective:** Deploy and integrate the foundational contracts for the Reputation-Based Royalties (RBR) system to enable early testing of the economic loop.  
* **Tasks:**  
  * Deploy the RoyaltyDistributor.sol and Reputation.sol contracts to the local Anvil testnet.  
  * Integrate simple calls to these contracts within the agent to be triggered after a lease is successfully completed, updating a placeholder reputation score and distributing a portion of the fee.

#### **3\. Create E2E Tests for Privacy & Economics**

* **Status:** 🟩 **To Do**  
* **Objective:** Write automated tests for the new privacy and economic features.  
* **Tasks:**  
  * Create an E2E test for a successful PySyft computation request.  
  * Create an E2E test that verifies the Reputation and RoyaltyDistributor contracts are correctly updated after a successful lease.

### **Path 5: Scaling & Interoperability (Sprint 5 and Beyond)**

**Objective:** To transition the system to a public testnet and ensure it can seamlessly integrate with the broader AI agent ecosystem.

#### **1\. Protocol Interoperability**

* **Status:** 🟩 **To Do**  
* **Objective:** Ensure Pandacea agents can be discovered and utilized by any standard AI agent. **This path is dependent on the completion of Paths 2, 3, and 4\.**  
* **Phase 1: Implement MCP Server Compliance:**  
  * **Task:** Integrate the official Go MCP SDK into the agent-backend.  
  * **Task:** Wrap the existing agent logic in an MCP Server interface. Expose data products as MCP Resources and core actions (request\_lease, request\_computation) as MCP Tools.  
* **Phase 2: Implement A2A Negotiation Protocol:**  
  * **Task:** Research and select a suitable A2A protocol standard.  
  * **Task:** Implement the protocol to allow for advanced, multi-agent workflows like data auctions and complex negotiations.

#### **2\. Public Testnet Deployment**

* **Status:** 🟩 **To Do**  
* **Objective:** Move the entire system from the local Anvil node to a public Ethereum testnet like Polygon Mumbai.  
* **Tasks:**  
  * Update deployment scripts in the contracts repo to target Mumbai.  
  * Fund deployer and test wallets with testnet MATIC.  
  * Update agent and SDK configurations to use a public RPC URL and the new public contract addresses.

#### **3\. Decentralized Discovery with Bootnodes**

* **Status:** 🟩 **To Do**  
* **Objective:** Evolve from local mDNS discovery to a robust DHT-based discovery mechanism suitable for a global network.  
* **Tasks:**  
  * Set up and run one or more "bootnode" agents at stable, public IP addresses.  
  * Hardcode the multiaddresses of these bootnodes into the agent software.