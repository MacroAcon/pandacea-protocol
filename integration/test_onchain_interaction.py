# integration/test_onchain_interaction.py

import os
import sys
import pytest
import time
import requests
from web3 import Web3

# Add the builder-sdk to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'builder-sdk'))

from pandacea_sdk.client import PandaceaClient

@pytest.mark.integration
class TestOnChainInteraction:
    @pytest.fixture(scope="class")
    def configured_client(self):
        """
        Initializes a PandaceaClient with blockchain configuration from environment variables.
        Skips the test if essential configuration is missing.
        """
        rpc_url = os.getenv("RPC_URL", "http://127.0.0.1:8545")
        contract_address = os.getenv("CONTRACT_ADDRESS")
        spender_private_key = os.getenv("SPENDER_PRIVATE_KEY")

        if not all([contract_address, spender_private_key]):
            pytest.skip("Missing required env vars: CONTRACT_ADDRESS, SPENDER_PRIVATE_KEY")

        client = PandaceaClient("http://localhost:8080") # Agent URL is not used for this specific test
        return client

    def test_create_lease_on_chain_and_verify_event(self, configured_client):
        """
        Tests the full on-chain lease creation flow and verifies the LeaseCreated event.
        """
        print("\n E2E Test: Creating lease on-chain and verifying event...")

        # 1. Define Test Parameters
        # These should match what the contract expects.
        # The earner address can be a dummy address for this test.
        earner_address = "0x" + "2" * 40 # A valid, but likely unused address
        data_product_id = b'test-data-product-id'.ljust(32, b'\0') # Must be 32 bytes
        max_price_wei = Web3.to_wei(0.01, 'ether')
        payment_wei = Web3.to_wei(0.001, 'ether')

        # 2. Get the latest block number before the transaction
        latest_block = configured_client.w3.eth.block_number

        # 3. Execute the on-chain lease creation
        print(f"Submitting createLease transaction...")
        tx_hash = configured_client.execute_lease_on_chain(
            earner=earner_address,
            data_product_id=data_product_id,
            max_price=max_price_wei,
            payment_in_wei=payment_wei
        )
        print(f" Transaction successful with hash: {tx_hash}")
        assert tx_hash is not None

        # 4. Verify the LeaseCreated Event
        print("Verifying LeaseCreated event on the blockchain...")
        
        # Give the node a moment, then check for the event in recent blocks
        time.sleep(2) 
        
        event_filter = configured_client.contract.events.LeaseCreated.create_filter(
            fromBlock=latest_block + 1
        )
        logs = event_filter.get_all_entries()

        assert len(logs) == 1, "Expected exactly one LeaseCreated event"
        
        event_data = logs[0]['args']
        spender_account = configured_client.w3.eth.account.from_key(configured_client.spender_private_key)

        # 5. Assert Event Parameters
        print(f" Found event. Validating parameters...")
        assert event_data['spender'] == spender_account.address
        assert event_data['earner'] == Web3.to_checksum_address(earner_address)
        assert event_data['price'] == payment_wei
        
        # The leaseId is dynamically generated, so we just check it exists
        assert event_data['leaseId'] is not None

        print("✅ E2E test passed: On-chain lease created and event verified successfully!")

    @pytest.fixture(scope="class")
    def agent_url(self):
        """Get the agent URL from environment variable."""
        return os.getenv('AGENT_API_URL', 'http://localhost:8080')

    def test_full_asynchronous_lease_flow(self, configured_client, agent_url):
        """
        Tests the full E2E flow:
        1. Create a lease proposal via the agent API.
        2. Execute the lease on-chain using the SDK.
        3. Poll the agent's status endpoint until the state is 'approved'.
        """
        print("\n E2E Test: Verifying full asynchronous lease state machine...")

        # Step 1: Discover a product to get a valid product ID
        products = configured_client.discover_products()
        assert len(products) > 0, "No products found to conduct the test"
        product_to_lease = products[0]
        print(f" Found product to lease: {product_to_lease.name}")

        # Step 2: Initiate the lease request via the Agent API to get a proposal ID
        # Note: This is an off-chain request to the agent.
        print("Requesting off-chain lease proposal from agent...")
        lease_proposal_id = configured_client.request_lease(
            product_id=product_to_lease.product_id,
            max_price="0.01",
            duration="24h"
        )
        assert lease_proposal_id is not None
        print(f" Received leaseProposalId: {lease_proposal_id}")

        # Step 3: Execute the lease on the blockchain
        print("Executing the lease on-chain...")
        earner_address = "0x" + "3" * 40 # Dummy earner for the test
        data_product_id_bytes = product_to_lease.product_id.encode('utf-8').ljust(32, b'\0')
        payment_wei = Web3.to_wei(0.001, 'ether')
        max_price_wei = Web3.to_wei(0.01, 'ether')

        tx_hash = configured_client.execute_lease_on_chain(
            earner=earner_address,
            data_product_id=data_product_id_bytes,
            max_price=max_price_wei,
            payment_in_wei=payment_wei
        )
        print(f" On-chain transaction sent with hash: {tx_hash}")

        # Step 4: Poll the agent's status endpoint until approved
        print(f"Polling agent for status update on lease: {lease_proposal_id}")
        timeout = 30  # seconds
        start_time = time.time()
        status = ""
        while time.time() - start_time < timeout:
            try:
                response = requests.get(f"{agent_url}/api/v1/leases/{lease_proposal_id}")
                if response.status_code == 200:
                    status_data = response.json()
                    status = status_data.get("status")
                    if status == "approved":
                        print(f"✅ Lease status successfully updated to 'approved'!")
                        break
            except requests.ConnectionError:
                pass # Agent might be restarting, continue polling
            time.sleep(2) # Poll every 2 seconds

        # Step 5: Assert the final status
        assert status == "approved", f"Test timed out. Lease status did not become 'approved'. Last status: '{status}'"

        print("✅ E2E test passed: Full asynchronous lease flow verified successfully!") 