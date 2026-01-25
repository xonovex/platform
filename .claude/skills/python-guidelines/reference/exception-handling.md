# exception-handling: Specific Exception Handling

**Guideline:** Raise and catch specific exception types, never use bare `except`.

**Rationale:** Specific exceptions enable precise error handling and better debugging. Bare `except` catches system exceptions like `KeyboardInterrupt` and `SystemExit`, making programs unresponsive. Custom exceptions document error conditions and enable targeted error handling. Specific exception handling improves code reliability and maintainability.

**Example:**

```python
from pathlib import Path
from typing import Optional

# Raise specific exceptions
def read_user_file(user_id: str) -> str:
    path = Path(f"users/{user_id}.json")

    if not path.exists():
        raise FileNotFoundError(f"User file not found: {user_id}")

    if not path.is_file():
        raise ValueError(f"Path is not a file: {path}")

    return path.read_text()

# Catch specific exceptions
def load_user(user_id: str) -> Optional[dict]:
    try:
        data = read_user_file(user_id)
        return parse_json(data)
    except FileNotFoundError:
        print(f"User {user_id} not found")
        return None
    except ValueError as e:
        print(f"Invalid data: {e}")
        return None
    except Exception as e:
        print(f"Unexpected error: {e}")
        raise

# Custom exceptions
class ValidationError(ValueError):
    """Raised when input validation fails"""
    pass

class DatabaseError(RuntimeError):
    """Raised when database operations fail"""
    pass

def validate_email(email: str) -> None:
    if "@" not in email:
        raise ValidationError(f"Invalid email format: {email}")

# ❌ Avoid bare except
try:
    risky_operation()
except:  # Don't do this - catches everything including KeyboardInterrupt
    pass

# ✅ Better - catch specific exceptions
try:
    risky_operation()
except (ValueError, TypeError) as e:
    handle_error(e)
```

**Techniques:**
- Raise specific built-in exceptions (`ValueError`, `TypeError`, `FileNotFoundError`)
- Create custom exceptions by subclassing built-in exceptions
- Catch specific exceptions or tuples of exceptions
- Use `except Exception` as a catch-all (not bare `except`)
- Re-raise unexpected exceptions after logging
- Provide meaningful error messages with context
