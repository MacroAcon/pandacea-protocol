#!/usr/bin/env python3
"""
Basic usage example for the Pandacea SDK.

This example demonstrates how to use the PandaceaClient to discover
available data products from a Pandacea Agent.
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
    client = PandaceaClient(agent_url, timeout=10.0)
    
    print(f"Connecting to agent at: {agent_url}")
    print()
    
    try:
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
    
    finally:
        # Clean up
        client.close()
        print("Client connection closed.")


if __name__ == "__main__":
    main() 