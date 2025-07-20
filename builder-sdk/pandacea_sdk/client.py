"""
Main client for interacting with the Pandacea Agent API.
"""

import json
import requests
from typing import List, Optional
from urllib.parse import urljoin

from .exceptions import AgentConnectionError, APIResponseError
from .models import DataProduct


class PandaceaClient:
    """
    Client for interacting with the Pandacea Agent API.
    
    This class provides a high-level interface for discovering data products
    and interacting with the Pandacea network.
    """
    
    def __init__(self, base_url: str, timeout: Optional[float] = 30.0):
        """
        Initialize the Pandacea client.
        
        Args:
            base_url: The base URL of the agent's API (e.g., "http://localhost:8080")
            timeout: Request timeout in seconds (default: 30.0)
        """
        self.base_url = base_url.rstrip('/')
        self.timeout = timeout
        self.session = requests.Session()
        
        # Set default headers
        self.session.headers.update({
            'User-Agent': 'Pandacea-SDK/0.1.0',
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        })
    
    def discover_products(self) -> List[DataProduct]:
        """
        Discover available data products from the agent.
        
        Makes a GET request to the /api/v1/products endpoint and returns
        a list of DataProduct objects.
        
        Returns:
            List of DataProduct objects representing available data products.
            
        Raises:
            AgentConnectionError: If unable to connect to the agent.
            APIResponseError: If the API returns an error response.
        """
        url = urljoin(self.base_url, '/api/v1/products')
        
        try:
            response = self.session.get(url, timeout=self.timeout)
            
            # Check if the request was successful
            response.raise_for_status()
            
            # Parse the JSON response
            try:
                data = response.json()
            except json.JSONDecodeError as e:
                raise APIResponseError(
                    f"Invalid JSON response from API: {e}",
                    status_code=response.status_code,
                    response_text=response.text
                )
            
            # Validate the response structure
            if not isinstance(data, dict):
                raise APIResponseError(
                    "API response is not a valid JSON object",
                    status_code=response.status_code,
                    response_text=response.text
                )
            
            if 'data' not in data:
                raise APIResponseError(
                    "API response missing 'data' field",
                    status_code=response.status_code,
                    response_text=response.text
                )
            
            if not isinstance(data['data'], list):
                raise APIResponseError(
                    "API response 'data' field is not a list",
                    status_code=response.status_code,
                    response_text=response.text
                )
            
            # Convert each product to a DataProduct object
            products = []
            for product_data in data['data']:
                try:
                    product = DataProduct(**product_data)
                    products.append(product)
                except Exception as e:
                    # Log the error but continue processing other products
                    # In a production environment, you might want to raise this
                    print(f"Warning: Failed to parse product data: {e}")
                    continue
            
            return products
            
        except requests.exceptions.ConnectionError as e:
            raise AgentConnectionError(
                f"Unable to connect to agent at {self.base_url}: {e}",
                original_error=e
            )
        except requests.exceptions.Timeout as e:
            raise AgentConnectionError(
                f"Request to agent timed out after {self.timeout} seconds: {e}",
                original_error=e
            )
        except requests.exceptions.HTTPError as e:
            # This handles 4xx and 5xx status codes
            raise APIResponseError(
                f"API returned error status {e.response.status_code}: {e.response.text}",
                status_code=e.response.status_code,
                response_text=e.response.text
            )
        except requests.exceptions.RequestException as e:
            raise AgentConnectionError(
                f"Request failed: {e}",
                original_error=e
            )
    
    def request_lease(self, product_id: str, max_price: str, duration: str) -> str:
        """
        Request a lease for a data product.
        
        Makes a POST request to the /api/v1/leases endpoint and returns
        the lease proposal ID.
        
        Args:
            product_id: The ID of the product to lease
            max_price: The maximum price willing to pay
            duration: The duration of the lease (e.g., "24h", "30m")
            
        Returns:
            The lease proposal ID as a string.
            
        Raises:
            AgentConnectionError: If unable to connect to the agent.
            APIResponseError: If the API returns an error response.
        """
        url = urljoin(self.base_url, '/api/v1/leases')
        
        # Prepare the request payload
        payload = {
            "productId": product_id,
            "maxPrice": max_price,
            "duration": duration
        }
        
        try:
            response = self.session.post(url, json=payload, timeout=self.timeout)
            
            # Check if the request was successful
            response.raise_for_status()
            
            # Parse the JSON response
            try:
                data = response.json()
            except json.JSONDecodeError as e:
                raise APIResponseError(
                    f"Invalid JSON response from API: {e}",
                    status_code=response.status_code,
                    response_text=response.text
                )
            
            # Validate the response structure
            if not isinstance(data, dict):
                raise APIResponseError(
                    "API response is not a valid JSON object",
                    status_code=response.status_code,
                    response_text=response.text
                )
            
            if 'leaseProposalId' not in data:
                raise APIResponseError(
                    "API response missing 'leaseProposalId' field",
                    status_code=response.status_code,
                    response_text=response.text
                )
            
            if not isinstance(data['leaseProposalId'], str):
                raise APIResponseError(
                    "API response 'leaseProposalId' field is not a string",
                    status_code=response.status_code,
                    response_text=response.text
                )
            
            return data['leaseProposalId']
            
        except requests.exceptions.ConnectionError as e:
            raise AgentConnectionError(
                f"Unable to connect to agent at {self.base_url}: {e}",
                original_error=e
            )
        except requests.exceptions.Timeout as e:
            raise AgentConnectionError(
                f"Request to agent timed out after {self.timeout} seconds: {e}",
                original_error=e
            )
        except requests.exceptions.HTTPError as e:
            # This handles 4xx and 5xx status codes
            raise APIResponseError(
                f"API returned error status {e.response.status_code}: {e.response.text}",
                status_code=e.response.status_code,
                response_text=e.response.text
            )
        except requests.exceptions.RequestException as e:
            raise AgentConnectionError(
                f"Request failed: {e}",
                original_error=e
            )
    
    def close(self):
        """
        Close the client session and free resources.
        
        This method should be called when you're done using the client
        to properly clean up the underlying HTTP session.
        """
        self.session.close()
    
    def __enter__(self):
        """Support for context manager protocol."""
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Support for context manager protocol."""
        self.close() 