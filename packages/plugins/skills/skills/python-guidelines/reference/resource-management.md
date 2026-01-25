# resource-management: Resource Management with Context Managers

**Guideline:** Use context managers (`with` and `async with`) for automatic resource cleanup.

**Rationale:** Context managers ensure resources are properly released even when exceptions occur. This prevents resource leaks and simplifies error handling. Custom context managers enable reusable resource management patterns like database transactions.

**Example:**

```python
from pathlib import Path
from contextlib import contextmanager, asynccontextmanager
import asyncio

# File operations with pathlib
def read_config(path: Path) -> dict[str, str]:
    with path.open('r') as f:
        return parse_config(f.read())

def write_config(path: Path, config: dict[str, str]) -> None:
    with path.open('w') as f:
        f.write(serialize_config(config))

# Custom context manager
@contextmanager
def database_transaction(connection):
    try:
        yield connection
        connection.commit()
    except Exception:
        connection.rollback()
        raise

# Usage
with database_transaction(conn) as db:
    db.execute("INSERT INTO users ...")

# Async context manager
@asynccontextmanager
async def async_database_transaction(connection):
    try:
        yield connection
        await connection.commit()
    except Exception:
        await connection.rollback()
        raise

# Usage
async with async_database_transaction(conn) as db:
    await db.execute("INSERT INTO users ...")
```

**Techniques:**
- Use `with` for synchronous resource operations
- Use `async with` for asynchronous resource operations
- Create custom context managers with `@contextmanager` decorator
- Create async context managers with `@asynccontextmanager`
- Handle cleanup in finally blocks or with explicit rollback
- Use pathlib.Path with context managers for file operations
