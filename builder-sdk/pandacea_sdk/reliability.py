#!/usr/bin/env python3
"""
Reliability module for Pandacea SDK
Provides exponential backoff with jitter and circuit breaker functionality.
"""

import os
import time
import random
import logging
from typing import Optional, Callable, Any, Dict
from functools import wraps
from enum import Enum

logger = logging.getLogger(__name__)

class CircuitState(Enum):
    """Circuit breaker states."""
    CLOSED = "closed"      # Normal operation
    OPEN = "open"          # Failing, reject requests
    HALF_OPEN = "half_open"  # Testing if service recovered

class CircuitBreaker:
    """Simple circuit breaker implementation."""
    
    def __init__(self, 
                 failure_threshold: int = 10,
                 reset_timeout: int = 30,
                 name: str = "default"):
        self.failure_threshold = failure_threshold
        self.reset_timeout = reset_timeout
        self.name = name
        self.state = CircuitState.CLOSED
        self.failure_count = 0
        self.last_failure_time = None
        self.success_count = 0
        
    def call(self, func: Callable, *args, **kwargs) -> Any:
        """Execute function with circuit breaker protection."""
        if self.state == CircuitState.OPEN:
            if self._should_attempt_reset():
                logger.info(f"Circuit {self.name}: attempting reset to half-open")
                self.state = CircuitState.HALF_OPEN
            else:
                raise CircuitBreakerOpenError(f"Circuit {self.name} is open")
        
        try:
            result = func(*args, **kwargs)
            self._on_success()
            return result
        except Exception as e:
            self._on_failure()
            raise
    
    def _should_attempt_reset(self) -> bool:
        """Check if enough time has passed to attempt reset."""
        if self.last_failure_time is None:
            return True
        return time.time() - self.last_failure_time >= self.reset_timeout
    
    def _on_success(self):
        """Handle successful call."""
        self.failure_count = 0
        if self.state == CircuitState.HALF_OPEN:
            self.success_count += 1
            if self.success_count >= 2:  # Require 2 successes to close
                logger.info(f"Circuit {self.name}: closing circuit after successful calls")
                self.state = CircuitState.CLOSED
                self.success_count = 0
    
    def _on_failure(self):
        """Handle failed call."""
        self.failure_count += 1
        self.last_failure_time = time.time()
        
        if self.state == CircuitState.HALF_OPEN:
            logger.warning(f"Circuit {self.name}: reopening circuit after failure")
            self.state = CircuitState.OPEN
            self.success_count = 0
        elif self.failure_count >= self.failure_threshold:
            logger.warning(f"Circuit {self.name}: opening circuit after {self.failure_count} failures")
            self.state = CircuitState.OPEN

class CircuitBreakerOpenError(Exception):
    """Raised when circuit breaker is open."""
    pass

def exponential_backoff_with_jitter(base_delay: float = 0.1, 
                                  max_delay: float = 60.0,
                                  max_retries: int = 5,
                                  jitter: bool = True) -> Callable:
    """
    Decorator for exponential backoff with jitter.
    
    Args:
        base_delay: Base delay in seconds
        max_delay: Maximum delay in seconds
        max_retries: Maximum number of retries
        jitter: Whether to add jitter to delays
    """
    def decorator(func: Callable) -> Callable:
        @wraps(func)
        def wrapper(*args, **kwargs):
            last_exception = None
            
            for attempt in range(max_retries + 1):
                try:
                    return func(*args, **kwargs)
                except Exception as e:
                    last_exception = e
                    
                    if attempt == max_retries:
                        logger.error(f"Max retries ({max_retries}) exceeded for {func.__name__}")
                        raise last_exception
                    
                    # Calculate delay with exponential backoff
                    delay = min(base_delay * (2 ** attempt), max_delay)
                    
                    # Add jitter if enabled
                    if jitter:
                        delay = delay * (0.5 + random.random() * 0.5)
                    
                    logger.warning(f"Attempt {attempt + 1}/{max_retries + 1} failed for {func.__name__}, "
                                 f"retrying in {delay:.2f}s: {e}")
                    
                    time.sleep(delay)
            
            raise last_exception
        return wrapper
    return decorator

class ReliabilityManager:
    """Manages reliability features for the SDK."""
    
    def __init__(self):
        # Load configuration from environment
        self.max_retries = int(os.getenv("SDK_MAX_RETRIES", "5"))
        self.base_delay_ms = int(os.getenv("SDK_BASE_DELAY_MS", "100"))
        self.circuit_fail_threshold = int(os.getenv("SDK_CIRCUIT_FAIL_THRESHOLD", "10"))
        self.circuit_reset_sec = int(os.getenv("SDK_CIRCUIT_RESET_SEC", "30"))
        
        # Initialize circuit breakers
        self.circuit_breakers: Dict[str, CircuitBreaker] = {}
        
        logger.info(f"ReliabilityManager initialized: max_retries={self.max_retries}, "
                   f"base_delay_ms={self.base_delay_ms}, circuit_fail_threshold={self.circuit_fail_threshold}")
    
    def get_circuit_breaker(self, name: str) -> CircuitBreaker:
        """Get or create a circuit breaker for the given name."""
        if name not in self.circuit_breakers:
            self.circuit_breakers[name] = CircuitBreaker(
                failure_threshold=self.circuit_fail_threshold,
                reset_timeout=self.circuit_reset_sec,
                name=name
            )
        return self.circuit_breakers[name]
    
    def with_reliability(self, 
                        circuit_name: str = "default",
                        max_retries: Optional[int] = None,
                        base_delay: Optional[float] = None) -> Callable:
        """
        Decorator that adds both circuit breaker and exponential backoff.
        
        Args:
            circuit_name: Name for the circuit breaker
            max_retries: Override max retries for this call
            base_delay: Override base delay for this call
        """
        def decorator(func: Callable) -> Callable:
            @wraps(func)
            def wrapper(*args, **kwargs):
                # Get circuit breaker
                circuit = self.get_circuit_breaker(circuit_name)
                
                # Apply exponential backoff
                retries = max_retries if max_retries is not None else self.max_retries
                delay = (base_delay if base_delay is not None else self.base_delay_ms) / 1000.0
                
                @exponential_backoff_with_jitter(
                    base_delay=delay,
                    max_retries=retries
                )
                def circuit_protected_func():
                    return circuit.call(func, *args, **kwargs)
                
                return circuit_protected_func()
            
            return wrapper
        return decorator

# Global reliability manager instance
_reliability_manager = ReliabilityManager()

def with_reliability(circuit_name: str = "default", 
                    max_retries: Optional[int] = None,
                    base_delay: Optional[float] = None) -> Callable:
    """Convenience function to apply reliability features."""
    return _reliability_manager.with_reliability(circuit_name, max_retries, base_delay)

def get_circuit_breaker(name: str) -> CircuitBreaker:
    """Get a circuit breaker by name."""
    return _reliability_manager.get_circuit_breaker(name)
