# caching-functions: Caching Pure Functions

Memoize pure, deterministic functions: `@cache` for unbounded, `@lru_cache(maxsize=N)` for bounded LRU eviction. Never cache functions with side effects or non-hashable args. Cache key is the full call signature.

```python
from functools import cache, lru_cache

@cache
def load_config() -> dict[str, str]:
    return parse_config(Path("config.txt").read_text())

@lru_cache(maxsize=128)
def fibonacci(n: int) -> int:
    return n if n < 2 else fibonacci(n - 1) + fibonacci(n - 2)
```
