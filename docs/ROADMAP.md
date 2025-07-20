## **Pandacea MVP: Comprehensive Roadmap & Technical Plan**

The current MVP successfully proves the core off-chain discovery loop of the protocol. This document provides a comprehensive technical roadmap to evolve the MVP from its current state into a feature-complete, robust, and scalable system ready for public testing.

### **Path 1: Hardening the MVP (Sprint 1 Polish)**

These improvements focus on making the existing system more resilient, configurable, and aligned with production-ready engineering practices without adding major new features.

#### **1\. Implement Persistent Agent Identity**

* **Objective:** Give the Earner agent a stable, recognizable address on the P2P network, which is essential for building reputation.  
* **Technical Implementation (agent-backend):**  
  1. **Modify p2p.NewNode:** In internal/p2p/node.go, before generating a new key, check for a key file at a specified path (e.g., from a config variable or a default like \~/.pandacea/agent.key).  
  2. **Key Handling:**  
     * If the file exists, use ioutil.ReadFile to load the marshalled private key. Use crypto.UnmarshalRsaPrivateKey to convert it back into a key object.  
     * If the file does not exist, generate a new key using crypto.GenerateKeyPair. Use crypto.MarshalRsaPrivateKey to serialize it, and ioutil.WriteFile to save it to the path with 0600 permissions.  
  3. **Configuration:** Add a key\_file\_path to config.yaml to make the location configurable.  
  4. **Security:** Ensure the key file path is added to the project's .gitignore file.

#### **2\. Implement Dynamic Data Product Listings**

* **Objective:** Decouple the agent's data listings from its code, allowing an Earner to manage their offerings without recompiling the application.  
* **Technical Implementation (agent-backend):**  
  1. **Create products.json:** In the root of the agent-backend, create a products.json file. Populate it with an array of data product objects that conform to the DataProduct schema.  
  2. **Modify api.NewServer:** In internal/api/server.go, load and parse products.json during the server's initialization. Store the list of products in the Server struct.  
  3. **Update handleGetProducts:** Modify this handler to serve the product list from the Server struct instead of a hardcoded mock.  
  4. **Error Handling:** Implement robust error handling for cases where products.json is missing or contains malformed JSON. The agent should log an error and start with an empty product list.

#### **3\. Enhance Integration Testing for Economic Rules**

* **Objective:** Create an automated end-to-end test that verifies the protocol's economic rules are being correctly enforced by the agent.  
* **Technical Implementation (integration & builder-sdk):**  
  1. **SDK Enhancement:** Add a request\_lease() method to the PandaceaClient in the SDK. This method should make a POST request to /api/v1/leases with a given productId and maxPrice.  
  2. **New Integration Test:** In integration/test\_integration.py, add a new test case, test\_error\_path\_lease\_below\_min\_price.  
  3. **Test Logic:** This test will use the new request\_lease() method to send a lease request with a price known to be below the agent's configured min\_price (e.g., "0.0001").  
  4. **Assertions:** The test must use pytest.raises(APIResponseError) to assert that the call fails as expected. It should then inspect the caught exception to verify that the status\_code is 403 and the response text contains the expected reason, "Proposed maxPrice is below the dynamic minimum price.".

### **Path 2: Activating the Protocol (Sprint 2\)**

These are major new features that connect the off-chain and on-chain components, bringing the full economic loop of the protocol to life.

#### **1\. Full On-Chain Integration**

* **Objective:** Enable the agent and SDK to interact directly with the LeaseAgreement.sol smart contract on the local Anvil testnet.  
* **Technical Implementation:**  
  1. **Contract ABI Generation:** Add a script to the contracts directory that uses abigen to generate Go bindings for the LeaseAgreement contract. Commit the generated LeaseAgreement.go file to the agent-backend repository.  
  2. **Configuration:** Add the deployed LeaseAgreement contract address to a shared configuration file or environment variable accessible by both the agent and the SDK.  
  3. **Agent Wallet (agent-backend):**  
     * Use the go-ethereum library.  
     * Load a private key from an environment variable (AGENT\_PRIVATE\_KEY).  
     * Instantiate a new contract instance using the generated Go bindings, the contract address, and a client connected to the Anvil RPC endpoint (http://anvil:8545).  
  4. **SDK Wallet (builder-sdk):**  
     * Use the web3.py library.  
     * Load a private key from an environment variable (SPENDER\_PRIVATE\_KEY).  
     * In a new lease.execute() function, build, sign, and send a transaction to the createLease function on the smart contract, including the correct payment amount.

#### **2\. Real Data Payloads with IPFS**

* **Objective:** Move from advertising data to making a verifiable sample available via the local IPFS node.  
* **Technical Implementation:**  
  1. **Agent IPFS Integration (agent-backend):**  
     * Add the go-ipfs-api library.  
     * On startup, the agent connects to the local IPFS node's API (http://ipfs:5001).  
     * The agent adds a sample data file (e.g., sample\_data/scan.pcd) to IPFS, receiving a Content ID (CID) in return.  
  2. **Dynamic Product Updates:** The agent should then update its in-memory product list (loaded from products.json) to include this real CID in the sampleCid field for the relevant product.  
  3. **SDK Verification:** Add a small utility to the SDK or a step in the integration test to take a sampleCid from a discovered product, fetch it from the IPFS gateway (http://localhost:9090), and verify its content.

#### **3\. Implement the Full Lease State Machine**

* **Objective:** Build out the complete, asynchronous negotiation flow, allowing the agent to react to on-chain events.  
* **Technical Implementation (agent-backend):**  
  1. **State Management:** Implement an in-memory map within the Server struct (e.g., pendingLeases map\[string\]\*LeaseProposal) to track the state of lease negotiations.  
  2. **Event Subscription:** Use the go-ethereum ethclient.SubscribeFilterLogs function to listen for LeaseCreated events from the LeaseAgreement contract.  
  3. **State Transition:** When the agent's event listener receives a LeaseCreated event, it should find the corresponding proposal in its pendingLeases map and update its status from "pending" to "approved".  
  4. **API Endpoint:** Implement the GET /api/v1/leases/{leaseProposalId} endpoint. This handler will look up the ID in the pendingLeases map and return the current status, allowing the SDK to poll and know when the lease has been successfully created on-chain.

### **Path 3: Scaling to a Public Testnet (Sprint 3 and Beyond)**

This path outlines the transition from a local, self-contained demo to a system capable of operating on a public network.

#### **1\. Public Testnet Deployment**

* **Objective:** Move the entire system from the local Anvil node to a public Ethereum testnet like Polygon Mumbai.  
* **Tasks:**  
  * Update deployment scripts in the contracts repo to target Mumbai.  
  * Fund deployer and test wallets with testnet MATIC.  
  * Update agent and SDK configurations to use a public RPC URL (e.g., from Infura or Alchemy) and the new public contract addresses.

#### **2\. Decentralized Discovery with Bootnodes**

* **Objective:** Evolve from local mDNS discovery to a robust DHT-based discovery mechanism suitable for a global network.  
* **Tasks:**  
  * Set up and run one or more "bootnode" agents at stable, public IP addresses.  
  * Hardcode the multiaddresses of these bootnodes into the agent software.  
  * When a new agent starts, it will first connect to these bootnodes to learn about other peers in the network, enabling true decentralized discovery.

#### **3\. Reputation & Royalty Systems**

* **Objective:** Implement the foundational contracts for the Reputation-Based Royalties (RBR) system.  
* **Tasks:**  
  * Deploy the RoyaltyDistributor.sol and Reputation.sol contracts.  
  * Integrate calls to these contracts within the agent after a lease is successfully completed or fails, updating reputation scores and distributing protocol fees.

#### **4\. Privacy Layer Integration (PySyft)**

* **Objective:** Begin implementing the "trust-by-proof" privacy model as outlined in the SDD.  
* **Tasks:**  
  * **Focus on Federated Analytics:** As decided in the SDD, the initial integration will not be end-to-end model training.  
  * The Earner agent will be responsible for locally pre-processing its raw data to extract key, non-sensitive features.  
  * The Spender will define a simple federated task (e.g., calculating the average of a specific feature) that gets executed across multiple Earner agents without exposing the underlying data.