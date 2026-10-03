# resource-management: Resource Management with Context Managers

Acquire resources with `with` / `async with` so cleanup runs on exception. Author custom managers with `@contextmanager` / `@asynccontextmanager`: `yield` the resource, commit on success, rollback in `except` then re-raise.

```python
from contextlib import contextmanager

@contextmanager
def transaction(conn):
    try:
        yield conn
        conn.commit()
    except Exception:
        conn.rollback()
        raise

with transaction(conn) as db:
    db.execute("INSERT INTO users ...")
```
