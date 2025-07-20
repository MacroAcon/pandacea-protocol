"""
Data models for the Pandacea SDK.
"""

from typing import List
from pydantic import BaseModel, Field, ConfigDict


class DataProduct(BaseModel):
    """
    Represents a data product returned by the Pandacea Agent API.
    
    Attributes:
        product_id: The unique identifier for the data product (DID format)
        name: Human-readable name of the data product
        data_type: The type of data (e.g., "RoboticSensorData")
        keywords: List of keywords/tags associated with the data product
    """
    
    model_config = ConfigDict(
        # Allow extra fields in case the API adds new fields
        extra="ignore",
        populate_by_name=True
    )
    
    product_id: str = Field(..., alias="productId", description="Unique identifier in DID format")
    name: str = Field(..., description="Human-readable name of the data product")
    data_type: str = Field(..., alias="dataType", description="Type of data (e.g., RoboticSensorData)")
    keywords: List[str] = Field(default_factory=list, description="Keywords/tags for the data product") 