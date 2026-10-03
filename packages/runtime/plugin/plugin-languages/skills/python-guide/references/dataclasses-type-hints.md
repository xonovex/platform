# dataclasses-type-hints: Dataclasses and Type Hints

Model structured data with `@dataclass` and annotated fields. Use `frozen=True` for immutable values, `field(default_factory=...)` for mutable defaults (never a bare `[]`/`{}` default), and PEP 695 `class Name[T]:` for generics.

```python
from dataclasses import dataclass, field
from typing import Any

@dataclass(frozen=True)
class User:
    id: str
    email: str
    metadata: dict[str, str] = field(default_factory=dict)

@dataclass
class Page[T]:
    items: list[T]
    total: int
```
