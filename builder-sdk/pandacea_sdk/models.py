from typing import Optional
from pydantic import BaseModel, HttpUrl

class LeaseCreateReq(BaseModel):
    dataset_id: str
    lessee: str
    max_price: int

class LeaseCreateRes(BaseModel):
    lease_id: str
    tx_hash: str
    explorer_url: Optional[HttpUrl] = None
