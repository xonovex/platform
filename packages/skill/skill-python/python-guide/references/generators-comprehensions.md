# generators-comprehensions: Generators and Comprehensions

Use comprehensions for small in-memory collections; use generator expressions (parentheses) and `yield` functions for large/streamed data so items are processed lazily without materializing the whole set. Don't wrap a generator in `list()` unless you need random access or reuse.

```python
squares = {x: x**2 for x in range(10)}          # dict comprehension
total = sum(x**2 for x in range(1_000_000))     # lazy generator expr

def read_lines(path: Path) -> Iterator[str]:    # lazy, O(1) memory
    with path.open() as f:
        for line in f:
            yield line.strip()
```
