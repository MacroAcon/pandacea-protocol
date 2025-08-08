"""
Pandacea SDK - Python SDK for interacting with the Pandacea Agent API.
"""

from .client import PandaceaClient
from .telemetry import init as telemetry_init
from .exceptions import PandaceaException, AgentConnectionError, APIResponseError
from .models import DataProduct

__version__ = "0.3.0"
__all__ = [
    "PandaceaClient",
    "PandaceaException", 
    "AgentConnectionError",
    "APIResponseError",
    "DataProduct",
    "telemetry_init",
] 