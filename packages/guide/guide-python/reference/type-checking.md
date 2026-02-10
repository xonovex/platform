# type-checking: Type Checking Best Practices

**Guideline:** Use type aliases, protocols, and modern type syntax for clear, maintainable type hints.

**Rationale:** Type aliases make complex types readable and reusable. Protocols enable structural typing (duck typing with type safety). Using `Sequence` and `Mapping` instead of `list` and `dict` in function parameters enables broader compatibility. Modern union syntax (`|`) is more concise than `Union`. Proper typing enables static analysis and better IDE support.

**Example:**

```python
from typing import TypeAlias, Protocol, Sequence, Mapping

# Type aliases
UserId: TypeAlias = str
Config: TypeAlias = dict[str, str | int | bool]

# Protocol for structural typing (duck typing with types)
class Drawable(Protocol):
    def draw(self) -> None: ...

def render(item: Drawable) -> None:
    item.draw()

# Sequence for read-only list-like
def sum_numbers(numbers: Sequence[int]) -> int:
    return sum(numbers)

# Mapping for read-only dict-like
def get_value(config: Mapping[str, str], key: str, default: str = "") -> str:
    return config.get(key, default)

# Use | for union types (Python 3.10+)
def parse_value(value: str | int | float) -> float:
    return float(value)

# Use None instead of Optional when clear
def find_user(user_id: str) -> User | None:
    """Returns User or None if not found"""
    pass
```

**Techniques:**

- Create type aliases with `TypeAlias` for complex types
- Use `Protocol` for structural typing requirements
- Use `Sequence` for read-only list-like parameters
- Use `Mapping` for read-only dict-like parameters
- Use `|` for union types instead of `Optional` or `Union`
- Return `Type | None` instead of `Optional[Type]` when clear
