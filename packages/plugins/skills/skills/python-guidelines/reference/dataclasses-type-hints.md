# dataclasses-type-hints: Dataclasses and Type Hints

**Guideline:** Use dataclasses with type hints for structured data definitions.

**Rationale:** Dataclasses reduce boilerplate code while providing automatic `__init__`, `__repr__`, and comparison methods. Type hints enable static analysis, IDE support, and self-documenting code. Frozen dataclasses create immutable objects that are safer in concurrent contexts.

**Example:**

```python
from dataclasses import dataclass, field
from datetime import datetime
from typing import Optional

# Immutable dataclass with type hints
@dataclass(frozen=True)
class User:
    id: str
    email: str
    name: str
    created_at: datetime
    metadata: dict[str, str] = field(default_factory=dict)

# Mutable dataclass
@dataclass
class Config:
    host: str = "localhost"
    port: int = 5432
    debug: bool = False
    timeout: float = 30.0

# Optional and union types
@dataclass(frozen=True)
class Response:
    status: int
    data: dict[str, any]
    error: Optional[str] = None

# Generic types
@dataclass
class Page[T]:
    items: list[T]
    total: int
    page: int
    page_size: int

# Usage with type safety
user = User(
    id="123",
    email="user@example.com",
    name="John Doe",
    created_at=datetime.now()
)

# Type checker knows this is str
print(user.email)
```

**Techniques:**
- Import `dataclass` decorator and type hints
- Annotate all fields with types using modern syntax (`list[T]`, `dict[K,V]`)
- Use `frozen=True` for immutable data structures
- Use `field(default_factory=...)` for mutable defaults
- Leverage generics with `[T]` syntax for reusable data structures
