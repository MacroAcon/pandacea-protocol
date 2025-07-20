"""
Unit tests for the DataProduct model.
"""

import pytest
from pandacea_sdk import DataProduct


class TestDataProduct:
    """Test cases for the DataProduct model."""
    
    def test_data_product_creation(self):
        """Test creating a DataProduct with valid data."""
        product = DataProduct(
            product_id="did:pandacea:earner:123/abc-456",
            name="Test Product",
            data_type="RoboticSensorData",
            keywords=["robotics", "sensor"]
        )
        
        assert product.product_id == "did:pandacea:earner:123/abc-456"
        assert product.name == "Test Product"
        assert product.data_type == "RoboticSensorData"
        assert product.keywords == ["robotics", "sensor"]
    
    def test_data_product_creation_without_keywords(self):
        """Test creating a DataProduct without keywords."""
        product = DataProduct(
            product_id="did:pandacea:earner:123/abc-456",
            name="Test Product",
            data_type="RoboticSensorData"
        )
        
        assert product.product_id == "did:pandacea:earner:123/abc-456"
        assert product.name == "Test Product"
        assert product.data_type == "RoboticSensorData"
        assert product.keywords == []
    
    def test_data_product_from_dict(self):
        """Test creating a DataProduct from a dictionary."""
        data = {
            "productId": "did:pandacea:earner:123/abc-456",
            "name": "Test Product",
            "dataType": "RoboticSensorData",
            "keywords": ["robotics", "sensor"]
        }
        
        product = DataProduct(**data)
        
        assert product.product_id == "did:pandacea:earner:123/abc-456"
        assert product.name == "Test Product"
        assert product.data_type == "RoboticSensorData"
        assert product.keywords == ["robotics", "sensor"]
    
    def test_data_product_missing_required_fields(self):
        """Test that missing required fields raise validation errors."""
        # Missing product_id
        with pytest.raises(ValueError):
            DataProduct(
                name="Test Product",
                data_type="RoboticSensorData"
            )
        
        # Missing name
        with pytest.raises(ValueError):
            DataProduct(
                product_id="did:pandacea:earner:123/abc-456",
                data_type="RoboticSensorData"
            )
        
        # Missing data_type
        with pytest.raises(ValueError):
            DataProduct(
                product_id="did:pandacea:earner:123/abc-456",
                name="Test Product"
            )
    
    def test_data_product_extra_fields_ignored(self):
        """Test that extra fields in the data are ignored."""
        data = {
            "productId": "did:pandacea:earner:123/abc-456",
            "name": "Test Product",
            "dataType": "RoboticSensorData",
            "keywords": ["robotics"],
            "extraField": "should be ignored",
            "anotherExtra": 123
        }
        
        product = DataProduct(**data)
        
        assert product.product_id == "did:pandacea:earner:123/abc-456"
        assert product.name == "Test Product"
        assert product.data_type == "RoboticSensorData"
        assert product.keywords == ["robotics"]
        
        # Extra fields should not be accessible
        assert not hasattr(product, 'extraField')
        assert not hasattr(product, 'anotherExtra')
    
    def test_data_product_serialization(self):
        """Test that DataProduct can be serialized to JSON."""
        product = DataProduct(
            product_id="did:pandacea:earner:123/abc-456",
            name="Test Product",
            data_type="RoboticSensorData",
            keywords=["robotics", "sensor"]
        )
        
        # Convert to dict
        product_dict = product.model_dump()
        
        assert product_dict["product_id"] == "did:pandacea:earner:123/abc-456"
        assert product_dict["name"] == "Test Product"
        assert product_dict["data_type"] == "RoboticSensorData"
        assert product_dict["keywords"] == ["robotics", "sensor"]
    
    def test_data_product_equality(self):
        """Test that DataProduct instances can be compared for equality."""
        product1 = DataProduct(
            product_id="did:pandacea:earner:123/abc-456",
            name="Test Product",
            data_type="RoboticSensorData",
            keywords=["robotics"]
        )
        
        product2 = DataProduct(
            product_id="did:pandacea:earner:123/abc-456",
            name="Test Product",
            data_type="RoboticSensorData",
            keywords=["robotics"]
        )
        
        product3 = DataProduct(
            product_id="did:pandacea:earner:456/def-789",
            name="Different Product",
            data_type="SensorData",
            keywords=["sensor"]
        )
        
        assert product1 == product2
        assert product1 != product3 