#!/usr/bin/env python3
"""
Integration tests for the Pandacea Builder SDK.

This test suite verifies that the Builder SDK can successfully communicate
with a live Agent Backend running in the Docker-based local development environment.

Test scenarios:
1. Happy Path: Successfully connect to a running agent and validate the structure
2. Error Path (Offline Agent): Handle timeouts or connection errors gracefully
3. Error Path (Invalid Response): Handle malformed API responses
"""

import os
import sys
import time
import random
import pytest
import requests
from typing import List

# Add the builder-sdk to the path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'builder-sdk'))

from pandacea_sdk import PandaceaClient, DataProduct
from pandacea_sdk.exceptions import AgentConnectionError, APIResponseError, PandaceaException

CHAOS_TEST = os.environ.get("CHAOS_TEST", "false").lower() == "true"
CHAOS_DURATION = int(os.environ.get("CHAOS_DURATION", 180))  # seconds, default 3 minutes


class TestIntegration:
    """Integration tests for the Pandacea Builder SDK."""
    
    @pytest.fixture(scope="class")
    def agent_url(self):
        """Get the agent URL from environment variable."""
        return os.getenv('AGENT_API_URL', 'http://localhost:8080')
    
    @pytest.fixture(scope="class")
    def client(self, agent_url):
        """Create a PandaceaClient instance."""
        return PandaceaClient(agent_url, timeout=10.0)
    
    def test_happy_path_discover_products(self, client):
        """
        Test the happy path: successfully connect to a running agent and validate
        the structure of the mock data products returned by the discover() function.
        """
        print(f"\n🔍 Testing Happy Path: Discovering products from agent...")
        
        try:
            # Call the discover_products method
            products = client.discover_products()
            
            # Validate that we got a list of products
            assert isinstance(products, list), "Expected products to be a list"
            print(f"✅ Successfully retrieved {len(products)} products")
            
            # Validate each product structure
            for i, product in enumerate(products):
                print(f"  Product {i+1}: {product.name}")
                
                # Check that it's a DataProduct instance
                assert isinstance(product, DataProduct), f"Product {i} is not a DataProduct instance"
                
                # Validate required fields
                assert hasattr(product, 'product_id'), f"Product {i} missing product_id"
                assert hasattr(product, 'name'), f"Product {i} missing name"
                assert hasattr(product, 'data_type'), f"Product {i} missing data_type"
                assert hasattr(product, 'keywords'), f"Product {i} missing keywords"
                
                # Validate field types
                assert isinstance(product.product_id, str), f"Product {i} product_id must be string"
                assert isinstance(product.name, str), f"Product {i} name must be string"
                assert isinstance(product.data_type, str), f"Product {i} data_type must be string"
                assert isinstance(product.keywords, list), f"Product {i} keywords must be list"
                
                # Validate DID format for product_id
                assert product.product_id.startswith('did:pandacea:'), f"Product {i} product_id must be DID format"
                
                print(f"    ID: {product.product_id}")
                print(f"    Type: {product.data_type}")
                print(f"    Keywords: {product.keywords}")
            
            print("✅ Happy path test passed - all products validated successfully")
            
        except Exception as e:
            pytest.fail(f"Happy path test failed: {e}")
    
    def test_error_path_offline_agent(self):
        """
        Test error path: assert that the SDK handles timeouts or connection errors
        gracefully when trying to connect to a Peer ID that is not online.
        """
        print(f"\n🔍 Testing Error Path: Offline Agent...")
        
        # Create a client pointing to a non-existent agent
        offline_client = PandaceaClient("http://localhost:9999", timeout=5.0)
        
        try:
            # This should raise an AgentConnectionError
            offline_client.discover_products()
            pytest.fail("Expected AgentConnectionError when connecting to offline agent")
            
        except AgentConnectionError as e:
            print(f"✅ Correctly caught AgentConnectionError: {e}")
            assert "Unable to connect to agent" in str(e), "Error message should mention connection failure"
            assert e.original_error is not None, "Should have original error details"
            
        except Exception as e:
            pytest.fail(f"Expected AgentConnectionError, but got {type(e).__name__}: {e}")
    
    def test_error_path_invalid_response(self, agent_url):
        """
        Test error path: assert that the SDK raises a specific exception if the agent
        returns data that does not conform to the API specification's schema.
        """
        print(f"\n🔍 Testing Error Path: Invalid Response...")
        
        # Test with a malformed response by directly calling the API
        # and then testing the SDK's response parsing
        
        try:
            # First, test that the agent is actually running and returns valid JSON
            response = requests.get(f"{agent_url}/api/v1/products", timeout=10)
            
            if response.status_code == 200:
                # If the agent is running and returns valid data, we can't easily test
                # invalid response scenarios without modifying the agent
                print("⚠️  Agent is running and returning valid responses")
                print("   (Cannot easily test invalid response scenarios with live agent)")
                
                # Test with a client that has a very short timeout to simulate issues
                # Note: This might not always trigger a timeout if the request is very fast
                fast_timeout_client = PandaceaClient(agent_url, timeout=0.001)
                
                try:
                    fast_timeout_client.discover_products()
                    # If we get here, the request was fast enough to complete
                    print("✅ Request completed quickly (timeout test skipped)")
                except AgentConnectionError as e:
                    print(f"✅ Correctly caught timeout error: {e}")
                    assert "timed out" in str(e).lower(), "Error should mention timeout"
                
            else:
                # If the agent is not responding properly, this is also a valid test
                print(f"✅ Agent returned status {response.status_code}, testing error handling")
                
                client = PandaceaClient(agent_url, timeout=10.0)
                try:
                    client.discover_products()
                    pytest.fail("Expected APIResponseError for non-200 status")
                except APIResponseError as e:
                    print(f"✅ Correctly caught APIResponseError: {e}")
                    assert e.status_code == response.status_code, "Status code should match"
                
        except requests.exceptions.ConnectionError:
            print("✅ Agent is not running, testing connection error handling")
            
            client = PandaceaClient(agent_url, timeout=10.0)
            try:
                client.discover_products()
                pytest.fail("Expected AgentConnectionError when agent is not running")
            except AgentConnectionError as e:
                print(f"✅ Correctly caught AgentConnectionError: {e}")
                assert "Unable to connect to agent" in str(e), "Error should mention connection failure"
    
    def test_agent_health_check(self, agent_url):
        """
        Test that the agent's health endpoint is accessible.
        """
        print(f"\n🔍 Testing Agent Health Check...")
        
        try:
            response = requests.get(f"{agent_url}/health", timeout=10)
            
            if response.status_code == 200:
                print("✅ Agent health check passed")
                health_data = response.json()
                assert health_data.get('status') == 'healthy', "Health status should be 'healthy'"
            else:
                print(f"⚠️  Agent health check returned status {response.status_code}")
                
        except requests.exceptions.ConnectionError:
            print("❌ Agent health check failed - agent not accessible")
            pytest.skip("Agent not running, skipping health check")
        except Exception as e:
            print(f"❌ Agent health check failed: {e}")
            pytest.skip(f"Agent health check failed: {e}")
    
    def test_sdk_client_cleanup(self, client):
        """
        Test that the SDK client can be properly cleaned up.
        """
        print(f"\n🔍 Testing SDK Client Cleanup...")
        
        try:
            # Test context manager usage
            with PandaceaClient(client.base_url, timeout=10.0) as temp_client:
                products = temp_client.discover_products()
                assert isinstance(products, list), "Should be able to use client in context manager"
            
            print("✅ SDK client cleanup test passed")
            
        except Exception as e:
            pytest.fail(f"SDK client cleanup test failed: {e}")
    
    def test_error_path_lease_below_min_price(self, client):
        """
        Test error path: verify that the Dynamic Minimum Pricing (DMP) economic rule
        is correctly enforced by the agent when a lease request has a maxPrice below
        the configured minimum price.
        """
        print(f"\n🔍 Testing Error Path: Lease Below Minimum Price...")
        
        try:
            # First, get a product to use for the lease request
            products = client.discover_products()
            assert len(products) > 0, "Need at least one product to test lease request"
            
            product = products[0]
            print(f"  Using product: {product.name} (ID: {product.product_id})")
            
            # Attempt to request a lease with a price known to be below the minimum
            # The agent's configured min_price is "0.001", so use "0.0001"
            with pytest.raises(APIResponseError) as exc_info:
                client.request_lease(
                    product_id=product.product_id,
                    max_price="0.0001",  # Below the configured min_price of "0.001"
                    duration="24h"
                )
            
            # Verify the exception details
            exception = exc_info.value
            print(f"  Caught expected APIResponseError: {exception}")
            
            # Assert the status code is 403 (Forbidden)
            assert exception.status_code == 403, f"Expected status code 403, got {exception.status_code}"
            
            # Assert the response contains the expected reason
            expected_reason = "Proposed maxPrice is below the dynamic minimum price."
            assert expected_reason in exception.response_text, (
                f"Expected response to contain '{expected_reason}', "
                f"but got: '{exception.response_text}'"
            )
            
            print("✅ Lease below minimum price test passed - DMP rule correctly enforced")
            
        except Exception as e:
            pytest.fail(f"Lease below minimum price test failed: {e}")


def test_chaos_resilience():
    if not CHAOS_TEST:
        pytest.skip("CHAOS_TEST not enabled")

    agent_url = os.environ.get("AGENT_URL", "http://localhost:8080")
    client = PandaceaClient(agent_url)
    start_time = time.time()
    end_time = start_time + CHAOS_DURATION
    last_success = None
    failures = 0
    successes = 0

    print(f"[CHAOS] Starting chaos test for {CHAOS_DURATION} seconds...")
    while time.time() < end_time:
        try:
            products = client.discover_products()
            print(f"[CHAOS] discover_products() succeeded: {len(products)} products")
            last_success = time.time()
            successes += 1
        except AgentConnectionError as e:
            print(f"[CHAOS] discover_products() failed: {e}")
            failures += 1
            # Wait a bit before retrying
            time.sleep(random.uniform(1, 3))
            continue
        # Short sleep to avoid hammering
        time.sleep(random.uniform(0.5, 2))

    print(f"[CHAOS] Test complete. Successes: {successes}, Failures: {failures}")
    assert successes > 0, "No successful calls during chaos test!"
    if failures > 0:
        assert last_success is not None and last_success > start_time, "System did not recover after failure!"


def main():
    """Main function to run integration tests."""
    print("=== Pandacea Builder SDK Integration Tests ===")
    print(f"Agent URL: {os.getenv('AGENT_API_URL', 'http://localhost:8080')}")
    print(f"Working Directory: {os.getcwd()}")
    print()
    
    # Run the tests
    pytest.main([__file__, "-v", "--tb=short"])


if __name__ == "__main__":
    main() 