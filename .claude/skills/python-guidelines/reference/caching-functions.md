# caching-functions: Caching Pure Functions

**Guideline:** Cache expensive pure function results with `@cache` or `@lru_cache` decorators.

**Rationale:** Caching eliminates redundant computations for pure functions (functions that always return the same output for the same input). `@cache` provides unbounded caching for small datasets, while `@lru_cache` offers size-limited caching with LRU eviction for larger datasets. This dramatically improves performance for recursive algorithms and expensive operations.

**Example:**

```python
from functools import cache, lru_cache
from pathlib import Path

# Cache with @cache (unbounded cache)
@cache
def load_config() -> dict[str, str]:
    """Load and parse config file (cached after first call)"""
    return parse_config(Path("config.txt").read_text())

# LRU cache with size limit
@lru_cache(maxsize=128)
def fibonacci(n: int) -> int:
    """Calculate Fibonacci number (cached with LRU eviction)"""
    if n < 2:
        return n
    return fibonacci(n - 1) + fibonacci(n - 2)

# Cache expensive computations
@lru_cache(maxsize=1000)
def calculate_hash(data: str) -> str:
    """Compute expensive hash (cached)"""
    import hashlib
    return hashlib.sha256(data.encode()).hexdigest()

# Usage - function only runs once per unique input
config1 = load_config()  # Loads from file
config2 = load_config()  # Returns cached value

fib_10 = fibonacci(10)  # Calculates
fib_10_again = fibonacci(10)  # Returns cached
```

**Techniques:**
- Use `@cache` for unbounded caching of pure functions
- Use `@lru_cache(maxsize=N)` for size-limited caching
- Only cache pure functions (no side effects, deterministic output)
- Consider cache size based on memory constraints
- Cache is per-function-call signature (different args = different cache entry)
