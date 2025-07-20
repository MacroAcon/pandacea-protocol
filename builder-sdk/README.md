# Pandacea SDK

A professional Python SDK for interacting with the Pandacea Agent API.

## Installation

```bash
# Using Poetry (recommended)
poetry install

# Or using pip
pip install -e .
```

## Quick Start

```python
from pandacea_sdk import PandaceaClient

# Initialize the client
client = PandaceaClient("http://localhost:8080")

# Discover available data products
try:
    products = client.discover_products()
    for product in products:
        print(f"Product: {product.name}")
        print(f"ID: {product.product_id}")
        print(f"Type: {product.data_type}")
        print(f"Keywords: {product.keywords}")
except Exception as e:
    print(f"Error: {e}")
```

## Development

```bash
# Install dependencies
poetry install

# Run tests
poetry run pytest

# Run linting
poetry run flake8 pandacea_sdk tests

# Format code
poetry run black pandacea_sdk tests
```

## API Reference

### PandaceaClient

The main client class for interacting with the Pandacea Agent API.

#### Methods

- `discover_products()`: Retrieve available data products from the agent

#### Exceptions

- `PandaceaException`: Base exception class
- `AgentConnectionError`: Raised when unable to connect to the agent
- `APIResponseError`: Raised when the API returns an error response 