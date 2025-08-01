"""
Main client for interacting with the Pandacea Agent API.
"""

import base64
import json
import logging
import os
import requests
from typing import List, Optional
from urllib.parse import urljoin

from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import rsa, padding
from cryptography.hazmat.backends import default_backend
import multibase
import multihash
from web3 import Web3

from .exceptions import AgentConnectionError, APIResponseError, PandaceaException
from .models import DataProduct


class PandaceaClient:
    """
    Client for interacting with the Pandacea Agent API.
    
    This class provides a high-level interface for discovering data products
    and interacting with the Pandacea network with cryptographic authentication.
    """
    
    def __init__(self, base_url: str, private_key_path: Optional[str] = None, timeout: Optional[float] = 30.0):
        """
        Initialize the Pandacea client.
        
        Args:
            base_url: The base URL of the agent's API (e.g., "http://localhost:8080")
            private_key_path: Path to the private key file for signing requests
            timeout: Request timeout in seconds (default: 30.0)
        """
        self.base_url = base_url.rstrip('/')
        self.timeout = timeout
        self.session = requests.Session()
        
        # Load private key if provided
        self.private_key = None
        self.peer_id = None
        if private_key_path:
            self._load_private_key(private_key_path)
        
        # Set default headers
        self.session.headers.update({
            'User-Agent': 'Pandacea-SDK/0.1.0',
            'Accept': 'application/json',
            'Content-Type': 'application/json',
        })

        # START ADDITIONS: Blockchain Integration
        self.rpc_url = os.getenv("RPC_URL", "http://127.0.0.1:8545")
        self.contract_address = os.getenv("CONTRACT_ADDRESS")
        self.spender_private_key = os.getenv("SPENDER_PRIVATE_KEY")
        
        # Initialize Web3 connection (but don't fail if not connected)
        try:
            self.w3 = Web3(Web3.HTTPProvider(self.rpc_url))
            try:
                if not self.w3.is_connected():
                    logging.warning(f"Failed to connect to blockchain node at {self.rpc_url}. On-chain functions will fail.")
                    self.w3 = None
            except Exception as e:
                logging.warning(f"Failed to check Web3 connection: {e}. On-chain functions will fail.")
                self.w3 = None
        except Exception as e:
            logging.warning(f"Failed to initialize Web3 connection: {e}. On-chain functions will fail.")
            self.w3 = None

        if not self.contract_address:
            logging.warning("CONTRACT_ADDRESS environment variable not set. On-chain functions will fail.")
            self.contract = None
        elif not self.w3:
            logging.warning("Web3 not connected. On-chain functions will fail.")
            self.contract = None
        else:
            try:
                # Assuming the ABI file is in a known relative path
                # NOTE: In a real SDK, this might be packaged differently
                abi_path = os.path.join(os.path.dirname(__file__), '../../../contracts/LeaseAgreement.abi.json')
                with open(abi_path, 'r') as f:
                    abi = json.load(f)
                
                self.contract = self.w3.eth.contract(address=self.contract_address, abi=abi)
            except FileNotFoundError:
                logging.warning("LeaseAgreement.abi.json not found. Please run the binding generation script.")
                self.contract = None
            except Exception as e:
                logging.warning(f"Failed to load contract: {e}")
                self.contract = None
        # END ADDITIONS
    
    def _load_private_key(self, key_path: str):
        """
        Load private key from file and derive peer ID.
        
        Args:
            key_path: Path to the private key file
        """
        try:
            # Expand tilde in path if present
            if key_path.startswith('~'):
                key_path = os.path.expanduser(key_path)
            
            # Read private key from file
            with open(key_path, 'rb') as f:
                key_data = f.read()
            
            # Load the private key
            self.private_key = serialization.load_pem_private_key(
                key_data,
                password=None,  # No password protection for now
                backend=default_backend()
            )
            
            # Extract public key and derive peer ID
            public_key = self.private_key.public_key()
            public_key_bytes = public_key.public_bytes(
                encoding=serialization.Encoding.DER,
                format=serialization.PublicFormat.SubjectPublicKeyInfo
            )
            
            # Create multihash of the public key
            mh = multihash.encode(public_key_bytes, 'sha2-256')
            # Create peer ID (base58-encoded multihash without multibase prefix)
            # This matches the libp2p peer ID format
            import base58
            self.peer_id = base58.b58encode(mh).decode('ascii')
            
        except Exception as e:
            raise ValueError(f"Failed to load private key from {key_path}: {e}")
    
    def _sign_request(self, data: bytes) -> str:
        """
        Sign request data with the client's private key.
        
        Args:
            data: Data to sign
            
        Returns:
            Base64-encoded signature
        """
        if not self.private_key:
            raise ValueError("No private key loaded. Initialize client with private_key_path.")
        
        # Sign the data
        signature = self.private_key.sign(
            data,
            padding.PKCS1v15(),
            hashes.SHA256()
        )
        
        # Return base64-encoded signature
        return base64.b64encode(signature).decode('ascii')
    
    def _prepare_headers(self, data: Optional[bytes] = None) -> dict:
        """
        Prepare headers for authenticated requests.
        
        Args:
            data: Request data to sign (if any)
            
        Returns:
            Dictionary of headers
        """
        headers = {}
        
        if self.peer_id:
            headers['X-Pandacea-Peer-ID'] = self.peer_id
            
            if data:
                # Sign the data and add signature header
                signature = self._sign_request(data)
                headers['X-Pandacea-Signature'] = signature
        
        return headers
    
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
        
        # Prepare headers for GET request
        # For GET requests, we sign a canonical representation
        canonical_data = f"GET /api/v1/products".encode('utf-8')
        headers = self._prepare_headers(canonical_data)
        
        try:
            response = self.session.get(url, headers=headers, timeout=self.timeout)
            
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
                    logging.warning("Failed to parse product data: %s. Product data: %s", e, product_data)
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
        
        # Serialize payload to JSON
        payload_json = json.dumps(payload, separators=(',', ':'))
        payload_bytes = payload_json.encode('utf-8')
        
        # Prepare headers with signature
        headers = self._prepare_headers(payload_bytes)
        
        try:
            response = self.session.post(url, data=payload_bytes, headers=headers, timeout=self.timeout)
            
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

    # START ADDITION: New method for on-chain lease creation
    def execute_lease_on_chain(self, earner: str, data_product_id: bytes, max_price: int, payment_in_wei: int) -> str:
        """
        Builds, signs, and sends a transaction to the createLease function.

        Args:
            earner: The blockchain address of the data earner.
            data_product_id: The 32-byte ID of the data product.
            max_price: The maximum price the spender is willing to pay, in wei.
            payment_in_wei: The amount to send with the transaction, in wei.

        Returns:
            The transaction hash as a hex string.
        """
        if not self.w3:
            raise PandaceaException("Web3 is not connected. Check RPC_URL and ensure blockchain node is running.")
        if not self.contract:
            raise PandaceaException("Contract is not initialized. Check CONTRACT_ADDRESS.")
        if not self.spender_private_key:
            raise PandaceaException("SPENDER_PRIVATE_KEY environment variable not set.")

        spender_account = self.w3.eth.account.from_key(self.spender_private_key)
        
        # Build the transaction
        tx_data = self.contract.functions.createLease(
            Web3.to_checksum_address(earner),
            data_product_id,
            max_price
        ).build_transaction({
            'from': spender_account.address,
            'value': payment_in_wei,
            'nonce': self.w3.eth.get_transaction_count(spender_account.address),
            'gas': 2000000, # This can be estimated more accurately
            'gasPrice': self.w3.eth.gas_price
        })

        # Sign the transaction
        signed_tx = self.w3.eth.account.sign_transaction(tx_data, self.spender_private_key)

        # Send the transaction
        tx_hash = self.w3.eth.send_raw_transaction(signed_tx.rawTransaction)

        # Wait for the transaction receipt (optional, but good for testing)
        receipt = self.w3.eth.wait_for_transaction_receipt(tx_hash)
        
        if receipt['status'] == 0:
            raise APIResponseError(f"Transaction failed. Receipt: {receipt}")

        return tx_hash.hex()
    # END ADDITION
    
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