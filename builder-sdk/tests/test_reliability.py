#!/usr/bin/env python3
"""
Unit tests for the reliability module.
"""

import pytest
import time
import os
from unittest.mock import Mock, patch
from pandacea_sdk.reliability import (
    CircuitBreaker, 
    CircuitState, 
    CircuitBreakerOpenError,
    exponential_backoff_with_jitter,
    ReliabilityManager,
    with_reliability
)


class TestCircuitBreaker:
    """Test cases for CircuitBreaker class."""
    
    def test_circuit_breaker_initial_state(self):
        """Test circuit breaker starts in CLOSED state."""
        cb = CircuitBreaker(failure_threshold=3, reset_timeout=10, name="test")
        assert cb.state == CircuitState.CLOSED
        assert cb.failure_count == 0
    
    def test_circuit_breaker_successful_call(self):
        """Test successful call doesn't change circuit state."""
        cb = CircuitBreaker(failure_threshold=3, reset_timeout=10, name="test")
        
        def successful_func():
            return "success"
        
        result = cb.call(successful_func)
        assert result == "success"
        assert cb.state == CircuitState.CLOSED
        assert cb.failure_count == 0
    
    def test_circuit_breaker_opens_after_threshold(self):
        """Test circuit breaker opens after failure threshold."""
        cb = CircuitBreaker(failure_threshold=2, reset_timeout=10, name="test")
        
        def failing_func():
            raise ValueError("test error")
        
        # First failure
        with pytest.raises(ValueError):
            cb.call(failing_func)
        assert cb.state == CircuitState.CLOSED
        assert cb.failure_count == 1
        
        # Second failure - should open circuit
        with pytest.raises(ValueError):
            cb.call(failing_func)
        assert cb.state == CircuitState.OPEN
        assert cb.failure_count == 2
    
    def test_circuit_breaker_rejects_when_open(self):
        """Test circuit breaker rejects calls when open."""
        cb = CircuitBreaker(failure_threshold=1, reset_timeout=10, name="test")
        cb.state = CircuitState.OPEN
        cb.last_failure_time = time.time()
        
        def any_func():
            return "should not be called"
        
        with pytest.raises(CircuitBreakerOpenError):
            cb.call(any_func)
    
    def test_circuit_breaker_resets_after_timeout(self):
        """Test circuit breaker resets to half-open after timeout."""
        cb = CircuitBreaker(failure_threshold=1, reset_timeout=0.1, name="test")
        cb.state = CircuitState.OPEN
        cb.last_failure_time = time.time() - 0.2  # Past timeout
        
        def successful_func():
            return "success"
        
        # Should attempt reset to half-open
        result = cb.call(successful_func)
        assert result == "success"
        assert cb.state == CircuitState.HALF_OPEN
    
    def test_circuit_breaker_closes_after_successful_calls(self):
        """Test circuit breaker closes after successful calls in half-open state."""
        cb = CircuitBreaker(failure_threshold=1, reset_timeout=0.1, name="test")
        cb.state = CircuitState.HALF_OPEN
        cb.success_count = 1
        
        def successful_func():
            return "success"
        
        result = cb.call(successful_func)
        assert result == "success"
        assert cb.state == CircuitState.CLOSED
        assert cb.success_count == 0


class TestExponentialBackoff:
    """Test cases for exponential backoff functionality."""
    
    def test_exponential_backoff_successful_call(self):
        """Test successful call doesn't retry."""
        @exponential_backoff_with_jitter(base_delay=0.1, max_retries=3)
        def successful_func():
            return "success"
        
        result = successful_func()
        assert result == "success"
    
    def test_exponential_backoff_retries_on_failure(self):
        """Test exponential backoff retries on failure."""
        call_count = 0
        
        @exponential_backoff_with_jitter(base_delay=0.01, max_retries=2)
        def failing_func():
            nonlocal call_count
            call_count += 1
            if call_count < 3:
                raise ValueError("test error")
            return "success"
        
        result = failing_func()
        assert result == "success"
        assert call_count == 3
    
    def test_exponential_backoff_max_retries_exceeded(self):
        """Test exponential backoff raises after max retries."""
        @exponential_backoff_with_jitter(base_delay=0.01, max_retries=2)
        def always_failing_func():
            raise ValueError("always fails")
        
        with pytest.raises(ValueError, match="always fails"):
            always_failing_func()
    
    def test_exponential_backoff_with_jitter(self):
        """Test that jitter is applied to delays."""
        delays = []
        
        @exponential_backoff_with_jitter(base_delay=0.1, max_retries=1, jitter=True)
        def failing_func():
            delays.append(time.time())
            raise ValueError("test error")
        
        start_time = time.time()
        with pytest.raises(ValueError):
            failing_func()
        
        # Should have at least 2 calls (initial + 1 retry)
        assert len(delays) >= 2
        
        # Check that delays are not exactly the same (jitter applied)
        if len(delays) >= 3:
            delay1 = delays[1] - delays[0]
            delay2 = delays[2] - delays[1]
            assert abs(delay1 - delay2) > 0.001  # Should be different due to jitter


class TestReliabilityManager:
    """Test cases for ReliabilityManager class."""
    
    def test_reliability_manager_initialization(self):
        """Test ReliabilityManager initializes with environment variables."""
        with patch.dict(os.environ, {
            'SDK_MAX_RETRIES': '3',
            'SDK_BASE_DELAY_MS': '200',
            'SDK_CIRCUIT_FAIL_THRESHOLD': '5',
            'SDK_CIRCUIT_RESET_SEC': '15'
        }):
            rm = ReliabilityManager()
            assert rm.max_retries == 3
            assert rm.base_delay_ms == 200
            assert rm.circuit_fail_threshold == 5
            assert rm.circuit_reset_sec == 15
    
    def test_reliability_manager_default_values(self):
        """Test ReliabilityManager uses default values when env vars not set."""
        with patch.dict(os.environ, {}, clear=True):
            rm = ReliabilityManager()
            assert rm.max_retries == 5
            assert rm.base_delay_ms == 100
            assert rm.circuit_fail_threshold == 10
            assert rm.circuit_reset_sec == 30
    
    def test_get_circuit_breaker(self):
        """Test getting circuit breakers by name."""
        rm = ReliabilityManager()
        
        cb1 = rm.get_circuit_breaker("test1")
        cb2 = rm.get_circuit_breaker("test2")
        cb1_again = rm.get_circuit_breaker("test1")
        
        assert cb1 is not cb2
        assert cb1 is cb1_again
        assert cb1.name == "test1"
        assert cb2.name == "test2"


class TestWithReliabilityDecorator:
    """Test cases for the with_reliability decorator."""
    
    def test_with_reliability_successful_call(self):
        """Test with_reliability decorator on successful call."""
        @with_reliability(circuit_name="test")
        def successful_func():
            return "success"
        
        result = successful_func()
        assert result == "success"
    
    def test_with_reliability_retries_on_failure(self):
        """Test with_reliability decorator retries on failure."""
        call_count = 0
        
        @with_reliability(circuit_name="test", max_retries=2)
        def failing_func():
            nonlocal call_count
            call_count += 1
            if call_count < 3:
                raise ValueError("test error")
            return "success"
        
        result = failing_func()
        assert result == "success"
        assert call_count == 3
    
    def test_with_reliability_circuit_breaker_integration(self):
        """Test that with_reliability integrates with circuit breaker."""
        call_count = 0
        
        @with_reliability(circuit_name="test", max_retries=1)
        def failing_func():
            nonlocal call_count
            call_count += 1
            raise ValueError("always fails")
        
        # First call should retry once
        with pytest.raises(ValueError):
            failing_func()
        assert call_count == 2
        
        # Second call should be rejected by circuit breaker
        with pytest.raises(CircuitBreakerOpenError):
            failing_func()
        assert call_count == 2  # Should not increment


if __name__ == "__main__":
    pytest.main([__file__])
