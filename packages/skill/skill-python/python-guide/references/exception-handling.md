# exception-handling: Specific Exception Handling

Raise and catch specific exception types. Never use bare `except:` — it swallows `KeyboardInterrupt`/`SystemExit`; use `except Exception` as the catch-all. Subclass built-ins for custom errors (`class ValidationError(ValueError)`). Re-raise after logging unexpected errors.

```python
def load_user(user_id: str) -> dict | None:
    try:
        return parse_json(read_user_file(user_id))
    except FileNotFoundError:
        return None
    except Exception:
        raise  # re-raise unexpected

class ValidationError(ValueError):
    """Raised when input validation fails."""
```
