# type-checking: Type Checking Best Practices

Use `X | None` over `Optional[X]` and `A | B` over `Union[A, B]`. Type parameters as `Sequence`/`Mapping` (read-only, broader) rather than concrete `list`/`dict`. Use `Protocol` for structural typing and PEP 695 `type` aliases for complex types.

```python
from typing import Protocol, Sequence, Mapping

type Config = dict[str, str | int | bool]

class Drawable(Protocol):
    def draw(self) -> None: ...

def sum_numbers(numbers: Sequence[int]) -> int:
    return sum(numbers)

def find_user(user_id: str) -> User | None: ...
```
