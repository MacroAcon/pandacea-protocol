"""Pandacea Protocol SDK Client."""

import requests
from typing import Optional
from .errors import PandaceaError, LeaseRejected, NetworkError
from .models import LeaseCreateReq, LeaseCreateRes

class PandaceaClient:
    """Client for interacting with the Pandacea Protocol."""
    
    def __init__(self, base_url: str = "http://localhost:8080"):
        self.base_url = base_url.rstrip('/')
        self.session = requests.Session()
    
    def create_lease(self, request: LeaseCreateReq) -> LeaseCreateRes:
        """Create a new lease with typed request/response."""
        try:
            response = self.session.post(
                f"{self.base_url}/api/leases",
                json=request.model_dump(),
                timeout=30
            )
            
            if response.status_code == 400:
                raise LeaseRejected(f"Lease rejected: {response.text}")
            elif response.status_code >= 500:
                raise NetworkError(f"Server error: {response.status_code}")
            elif response.status_code != 200:
                raise PandaceaError(f"Unexpected status: {response.status_code}")
            
            data = response.json()
            return LeaseCreateRes(**data)
            
        except requests.RequestException as e:
            raise NetworkError(f"Network request failed: {e}")
        except Exception as e:
            raise PandaceaError(f"Unexpected error: {e}")
    
    def get_lease(self, lease_id: str) -> Optional[LeaseCreateRes]:
        """Get lease information by ID."""
        try:
            response = self.session.get(
                f"{self.base_url}/api/leases/{lease_id}",
                timeout=30
            )
            
            if response.status_code == 404:
                return None
            elif response.status_code >= 500:
                raise NetworkError(f"Server error: {response.status_code}")
            elif response.status_code != 200:
                raise PandaceaError(f"Unexpected status: {response.status_code}")
            
            data = response.json()
            return LeaseCreateRes(**data)
            
        except requests.RequestException as e:
            raise NetworkError(f"Network request failed: {e}")
        except Exception as e:
            raise PandaceaError(f"Unexpected error: {e}")
