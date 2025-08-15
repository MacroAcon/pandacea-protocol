class PandaceaError(Exception):
    """Base error for the SDK."""

class LeaseRejected(PandaceaError):
    """Raised when a lease cannot be created."""

class NetworkError(PandaceaError):
    """Raised when network interaction fails."""
