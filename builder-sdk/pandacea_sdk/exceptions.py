"""
Custom exceptions for the Pandacea SDK.
"""


class PandaceaException(Exception):
    """
    Base exception class for all Pandacea SDK exceptions.
    
    All other exceptions in this module inherit from this class,
    allowing developers to catch all Pandacea-related exceptions
    with a single except clause.
    """
    
    def __init__(self, message: str, *args):
        super().__init__(message, *args)
        self.message = message


class AgentConnectionError(PandaceaException):
    """
    Raised when the SDK cannot connect to the Pandacea Agent.
    
    This exception is raised when there are network connectivity issues,
    DNS resolution failures, or the agent is not running.
    """
    
    def __init__(self, message: str, original_error: Exception = None):
        super().__init__(message)
        self.original_error = original_error


class APIResponseError(PandaceaException):
    """
    Raised when the Pandacea Agent API returns an error response.
    
    This exception is raised when the API returns a non-200 status code
    or when the response cannot be parsed as valid JSON.
    """
    
    def __init__(self, message: str, status_code: int = None, response_text: str = None):
        super().__init__(message)
        self.status_code = status_code
        self.response_text = response_text 