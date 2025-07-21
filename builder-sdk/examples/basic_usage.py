#!/usr/bin/env python3
"""
Basic usage example for the Pandacea SDK.

This example demonstrates how to use the PandaceaClient to discover
available data products from a Pandacea Agent with cryptographic authentication.
"""

import sys
import os

# Add the parent directory to the path so we can import the SDK
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..'))

from pandacea_sdk import PandaceaClient, PandaceaException, AgentConnectionError, APIResponseError


def main():
    """Main function demonstrating SDK usage."""
    print("=== Pandacea SDK Basic Usage Example ===\n")
    
    # Initialize the client
    # Replace with your actual agent URL
    agent_url = "http://localhost:8080"
    
    # Example 1: Unauthenticated client (for backward compatibility)
    print("Example 1: Unauthenticated client (limited functionality)")
    print("Note: This will only work if the agent allows unauthenticated requests")
    print()
    
    try:
        client = PandaceaClient(agent_url, timeout=10.0)
        print(f"Connecting to agent at: {agent_url}")
        
        # Discover available products
        print("Discovering available data products...")
        products = client.discover_products()
        
        if not products:
            print("No data products found.")
        else:
            print(f"Found {len(products)} data product(s):\n")
            
            for i, product in enumerate(products, 1):
                print(f"Product {i}:")
                print(f"  ID: {product.product_id}")
                print(f"  Name: {product.name}")
                print(f"  Type: {product.data_type}")
                print(f"  Keywords: {', '.join(product.keywords) if product.keywords else 'None'}")
                print()
        
        client.close()
        
    except Exception as e:
        print(f"❌ Error with unauthenticated client: {e}")
        print("This is expected if the agent requires authentication.\n")
    
    # Example 2: Authenticated client with private key
    print("Example 2: Authenticated client with private key")
    print("This demonstrates the secure, authenticated approach")
    print()
    
    # Path to your private key file (adjust as needed)
    private_key_path = "~/.pandacea/agent.key"  # Example path
    
    try:
        # Initialize authenticated client
        client = PandaceaClient(
            base_url=agent_url,
            private_key_path=private_key_path,
            timeout=10.0
        )
        
        print(f"Connecting to agent at: {agent_url}")
        print(f"Using private key from: {private_key_path}")
        print(f"Peer ID: {client.peer_id}")
        print()
        
        # Discover available products (now with authentication)
        print("Discovering available data products (authenticated)...")
        products = client.discover_products()
        
        if not products:
            print("No data products found.")
        else:
            print(f"Found {len(products)} data product(s):\n")
            
            for i, product in enumerate(products, 1):
                print(f"Product {i}:")
                print(f"  ID: {product.product_id}")
                print(f"  Name: {product.name}")
                print(f"  Type: {product.data_type}")
                print(f"  Keywords: {', '.join(product.keywords) if product.keywords else 'None'}")
                print()
            
            # Example: Request a lease for the first product
            if products:
                print("Example: Requesting a lease...")
                try:
                    lease_id = client.request_lease(
                        product_id=products[0].product_id,
                        max_price="10.50",
                        duration="24h"
                    )
                    print(f"✅ Lease request successful! Lease ID: {lease_id}")
                except Exception as e:
                    print(f"❌ Lease request failed: {e}")
        
        client.close()
        
    except ValueError as e:
        print(f"❌ Configuration Error: {e}")
        print("Make sure the private key file exists and is accessible.")
        print("You can generate a key pair using the Pandacea agent tools.")
    
    except AgentConnectionError as e:
        print(f"❌ Connection Error: {e}")
        print("Make sure the Pandacea Agent is running and accessible.")
        if e.original_error:
            print(f"Original error: {e.original_error}")
    
    except APIResponseError as e:
        print(f"❌ API Error: {e}")
        if e.status_code:
            print(f"Status code: {e.status_code}")
        if e.response_text:
            print(f"Response: {e.response_text}")
    
    except PandaceaException as e:
        print(f"❌ Pandacea Error: {e}")
    
    except Exception as e:
        print(f"❌ Unexpected Error: {e}")
    
    print("\n=== Example completed ===")


if __name__ == "__main__":
    main() 