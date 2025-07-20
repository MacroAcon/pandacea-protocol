"""
Unit tests for the PandaceaClient.
"""

import json
import pytest
import requests
import requests_mock
from unittest.mock import Mock

from pandacea_sdk import PandaceaClient, DataProduct
from pandacea_sdk.exceptions import AgentConnectionError, APIResponseError


class TestPandaceaClient:
    """Test cases for the PandaceaClient class."""
    
    def test_client_initialization(self):
        """Test client initialization with default parameters."""
        client = PandaceaClient("http://localhost:8080")
        assert client.base_url == "http://localhost:8080"
        assert client.timeout == 30.0
        assert client.session.headers['User-Agent'] == 'Pandacea-SDK/0.1.0'
        assert client.session.headers['Accept'] == 'application/json'
        assert client.session.headers['Content-Type'] == 'application/json'
    
    def test_client_initialization_with_timeout(self):
        """Test client initialization with custom timeout."""
        client = PandaceaClient("http://localhost:8080", timeout=60.0)
        assert client.timeout == 60.0
    
    def test_client_initialization_with_trailing_slash(self):
        """Test client initialization handles trailing slashes correctly."""
        client = PandaceaClient("http://localhost:8080/")
        assert client.base_url == "http://localhost:8080"
    
    def test_discover_products_success(self):
        """Test successful product discovery."""
        mock_response = {
            "data": [
                {
                    "productId": "did:pandacea:earner:123/abc-456",
                    "name": "Novel Package 3D Scans - Warehouse A",
                    "dataType": "RoboticSensorData",
                    "keywords": ["robotics", "3d-scan", "lidar"]
                },
                {
                    "productId": "did:pandacea:earner:456/def-789",
                    "name": "Temperature Sensor Data - Warehouse B",
                    "dataType": "SensorData",
                    "keywords": ["temperature", "sensor", "monitoring"]
                }
            ],
            "nextCursor": "cursor_def456"
        }
        
        with requests_mock.Mocker() as m:
            m.get("http://localhost:8080/api/v1/products", json=mock_response)
            
            client = PandaceaClient("http://localhost:8080")
            products = client.discover_products()
            
            assert len(products) == 2
            
            # Check first product
            assert products[0].product_id == "did:pandacea:earner:123/abc-456"
            assert products[0].name == "Novel Package 3D Scans - Warehouse A"
            assert products[0].data_type == "RoboticSensorData"
            assert products[0].keywords == ["robotics", "3d-scan", "lidar"]
            
            # Check second product
            assert products[1].product_id == "did:pandacea:earner:456/def-789"
            assert products[1].name == "Temperature Sensor Data - Warehouse B"
            assert products[1].data_type == "SensorData"
            assert products[1].keywords == ["temperature", "sensor", "monitoring"]
    
    def test_discover_products_empty_response(self):
        """Test product discovery with empty data list."""
        mock_response = {
            "data": [],
            "nextCursor": None
        }
        
        with requests_mock.Mocker() as m:
            m.get("http://localhost:8080/api/v1/products", json=mock_response)
            
            client = PandaceaClient("http://localhost:8080")
            products = client.discover_products()
            
            assert len(products) == 0
    
    def test_discover_products_connection_error(self):
        """Test connection error handling."""
        with requests_mock.Mocker() as m:
            m.get("http://localhost:8080/api/v1/products", exc=requests.exceptions.ConnectionError("Connection refused"))
            
            client = PandaceaClient("http://localhost:8080")
            
            with pytest.raises(AgentConnectionError) as exc_info:
                client.discover_products()
            
            assert "Unable to connect to agent" in str(exc_info.value)
            assert exc_info.value.original_error is not None
    
    def test_discover_products_timeout_error(self):
        """Test timeout error handling."""
        with requests_mock.Mocker() as m:
            m.get("http://localhost:8080/api/v1/products", exc=requests.exceptions.Timeout("Request timed out"))
            
            client = PandaceaClient("http://localhost:8080", timeout=5.0)
            
            with pytest.raises(AgentConnectionError) as exc_info:
                client.discover_products()
            
            assert "Request to agent timed out after 5.0 seconds" in str(exc_info.value)
            assert exc_info.value.original_error is not None
    
    def test_discover_products_http_error_500(self):
        """Test HTTP 500 error handling."""
        with requests_mock.Mocker() as m:
            m.get("http://localhost:8080/api/v1/products", status_code=500, text="Internal Server Error")
            
            client = PandaceaClient("http://localhost:8080")
            
            with pytest.raises(APIResponseError) as exc_info:
                client.discover_products()
            
            assert exc_info.value.status_code == 500
            assert exc_info.value.response_text == "Internal Server Error"
            assert "API returned error status 500" in str(exc_info.value)
    
    def test_discover_products_http_error_404(self):
        """Test HTTP 404 error handling."""
        with requests_mock.Mocker() as m:
            m.get("http://localhost:8080/api/v1/products", status_code=404, text="Not Found")
            
            client = PandaceaClient("http://localhost:8080")
            
            with pytest.raises(APIResponseError) as exc_info:
                client.discover_products()
            
            assert exc_info.value.status_code == 404
            assert exc_info.value.response_text == "Not Found"
    
    def test_discover_products_invalid_json(self):
        """Test handling of invalid JSON response."""
        with requests_mock.Mocker() as m:
            m.get("http://localhost:8080/api/v1/products", text="Invalid JSON")
            
            client = PandaceaClient("http://localhost:8080")
            
            with pytest.raises(APIResponseError) as exc_info:
                client.discover_products()
            
            assert "Invalid JSON response from API" in str(exc_info.value)
            assert exc_info.value.status_code == 200
            assert exc_info.value.response_text == "Invalid JSON"
    
    def test_discover_products_missing_data_field(self):
        """Test handling of response missing 'data' field."""
        mock_response = {
            "nextCursor": "cursor_def456"
            # Missing 'data' field
        }
        
        with requests_mock.Mocker() as m:
            m.get("http://localhost:8080/api/v1/products", json=mock_response)
            
            client = PandaceaClient("http://localhost:8080")
            
            with pytest.raises(APIResponseError) as exc_info:
                client.discover_products()
            
            assert "API response missing 'data' field" in str(exc_info.value)
    
    def test_discover_products_data_not_list(self):
        """Test handling of response where 'data' is not a list."""
        mock_response = {
            "data": "not a list",
            "nextCursor": "cursor_def456"
        }
        
        with requests_mock.Mocker() as m:
            m.get("http://localhost:8080/api/v1/products", json=mock_response)
            
            client = PandaceaClient("http://localhost:8080")
            
            with pytest.raises(APIResponseError) as exc_info:
                client.discover_products()
            
            assert "API response 'data' field is not a list" in str(exc_info.value)
    
    def test_discover_products_response_not_dict(self):
        """Test handling of response that is not a dictionary."""
        with requests_mock.Mocker() as m:
            m.get("http://localhost:8080/api/v1/products", json="not a dict")
            
            client = PandaceaClient("http://localhost:8080")
            
            with pytest.raises(APIResponseError) as exc_info:
                client.discover_products()
            
            assert "API response is not a valid JSON object" in str(exc_info.value)
    
    def test_discover_products_invalid_product_data(self):
        """Test handling of invalid product data in response."""
        mock_response = {
            "data": [
                {
                    "productId": "did:pandacea:earner:123/abc-456",
                    "name": "Valid Product",
                    "dataType": "RoboticSensorData",
                    "keywords": ["robotics"]
                },
                {
                    "productId": "invalid-product",  # Missing required fields
                    # Missing name, dataType
                }
            ],
            "nextCursor": "cursor_def456"
        }
        
        with requests_mock.Mocker() as m:
            m.get("http://localhost:8080/api/v1/products", json=mock_response)
            
            client = PandaceaClient("http://localhost:8080")
            products = client.discover_products()
            
            # Should return only the valid product
            assert len(products) == 1
            assert products[0].product_id == "did:pandacea:earner:123/abc-456"
            assert products[0].name == "Valid Product"
    
    def test_client_context_manager(self):
        """Test client as context manager."""
        mock_response = {
            "data": [],
            "nextCursor": None
        }
        
        with requests_mock.Mocker() as m:
            m.get("http://localhost:8080/api/v1/products", json=mock_response)
            
            with PandaceaClient("http://localhost:8080") as client:
                products = client.discover_products()
                assert len(products) == 0
    
    def test_client_close(self):
        """Test client close method."""
        client = PandaceaClient("http://localhost:8080")
        client.close()
        # The session should be closed (we can't easily test this without mocking)
        # But at least it shouldn't raise an exception 