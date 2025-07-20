"""
Unit tests for the custom exception classes.
"""

import pytest
from pandacea_sdk.exceptions import PandaceaException, AgentConnectionError, APIResponseError


class TestPandaceaException:
    """Test cases for the base PandaceaException class."""
    
    def test_pandacea_exception_creation(self):
        """Test creating a PandaceaException."""
        message = "Test error message"
        exception = PandaceaException(message)
        
        assert str(exception) == message
        assert exception.message == message
    
    def test_pandacea_exception_inheritance(self):
        """Test that PandaceaException inherits from Exception."""
        exception = PandaceaException("Test message")
        assert isinstance(exception, Exception)


class TestAgentConnectionError:
    """Test cases for the AgentConnectionError class."""
    
    def test_agent_connection_error_creation(self):
        """Test creating an AgentConnectionError."""
        message = "Connection failed"
        exception = AgentConnectionError(message)
        
        assert str(exception) == message
        assert exception.message == message
        assert exception.original_error is None
    
    def test_agent_connection_error_with_original_error(self):
        """Test creating an AgentConnectionError with original error."""
        message = "Connection failed"
        original_error = ConnectionError("Network unreachable")
        exception = AgentConnectionError(message, original_error)
        
        assert str(exception) == message
        assert exception.message == message
        assert exception.original_error == original_error
    
    def test_agent_connection_error_inheritance(self):
        """Test that AgentConnectionError inherits from PandaceaException."""
        exception = AgentConnectionError("Test message")
        assert isinstance(exception, PandaceaException)
        assert isinstance(exception, Exception)


class TestAPIResponseError:
    """Test cases for the APIResponseError class."""
    
    def test_api_response_error_creation(self):
        """Test creating an APIResponseError."""
        message = "API error occurred"
        exception = APIResponseError(message)
        
        assert str(exception) == message
        assert exception.message == message
        assert exception.status_code is None
        assert exception.response_text is None
    
    def test_api_response_error_with_status_code(self):
        """Test creating an APIResponseError with status code."""
        message = "API error occurred"
        status_code = 500
        exception = APIResponseError(message, status_code)
        
        assert str(exception) == message
        assert exception.message == message
        assert exception.status_code == status_code
        assert exception.response_text is None
    
    def test_api_response_error_with_all_parameters(self):
        """Test creating an APIResponseError with all parameters."""
        message = "API error occurred"
        status_code = 404
        response_text = "Not Found"
        exception = APIResponseError(message, status_code, response_text)
        
        assert str(exception) == message
        assert exception.message == message
        assert exception.status_code == status_code
        assert exception.response_text == response_text
    
    def test_api_response_error_inheritance(self):
        """Test that APIResponseError inherits from PandaceaException."""
        exception = APIResponseError("Test message")
        assert isinstance(exception, PandaceaException)
        assert isinstance(exception, Exception)


class TestExceptionHierarchy:
    """Test the exception hierarchy and catching behavior."""
    
    def test_catch_all_pandacea_exceptions(self):
        """Test that all Pandacea exceptions can be caught with the base class."""
        exceptions = [
            PandaceaException("Base exception"),
            AgentConnectionError("Connection error"),
            APIResponseError("API error")
        ]
        
        for exception in exceptions:
            try:
                raise exception
            except PandaceaException as e:
                assert isinstance(e, PandaceaException)
            except Exception:
                pytest.fail(f"Should have been caught by PandaceaException: {exception}")
    
    def test_exception_message_preservation(self):
        """Test that exception messages are preserved correctly."""
        test_message = "This is a test error message"
        
        # Test base exception
        base_exception = PandaceaException(test_message)
        assert str(base_exception) == test_message
        
        # Test connection error
        conn_exception = AgentConnectionError(test_message)
        assert str(conn_exception) == test_message
        
        # Test API error
        api_exception = APIResponseError(test_message)
        assert str(api_exception) == test_message 