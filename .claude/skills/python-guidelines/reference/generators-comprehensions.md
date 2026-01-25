# generators-comprehensions: Generators and Comprehensions

**Guideline:** Use generators and comprehensions for memory-efficient, lazy data processing.

**Rationale:** Generators process items on-demand without loading entire datasets into memory, enabling efficient handling of large files and streams. Comprehensions provide concise, readable syntax for transformations. Generator expressions and functions enable pipeline-style processing with minimal memory overhead.

**Example:**

```python
from typing import Iterator, Iterable
from pathlib import Path

# List comprehension
numbers = [x * 2 for x in range(10) if x % 2 == 0]

# Dict comprehension
squares = {x: x**2 for x in range(10)}

# Set comprehension
unique_lengths = {len(word) for word in words}

# Generator expression (lazy, memory efficient)
total = sum(x**2 for x in range(1000000))

# Generator function
def read_large_file(path: Path) -> Iterator[str]:
    """Read file line by line without loading all into memory"""
    with path.open('r') as f:
        for line in f:
            yield line.strip()

# Generator with processing
def process_items(items: Iterable[str]) -> Iterator[dict[str, any]]:
    """Process items lazily"""
    for item in items:
        if item:
            yield {"value": item, "length": len(item)}

# Usage - only processes items as needed
for processed in process_items(read_large_file(Path("data.txt"))):
    print(processed)

# ❌ Bad - loads entire file into memory
def read_large_file_bad(path: Path) -> list[str]:
    with path.open('r') as f:
        return [line.strip() for line in f]  # All in memory!
```

**Techniques:**
- Use list/dict/set comprehensions for small, in-memory collections
- Use generator expressions (parentheses) for large or streamed data
- Create generator functions with `yield` for lazy iteration
- Chain generators for pipeline processing
- Avoid materializing generators until needed (don't convert to list unnecessarily)
