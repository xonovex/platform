# pathlib-file-ops: Pathlib for File Operations

**Guideline:** Use `pathlib.Path` for all file and directory operations instead of `os.path`.

**Rationale:** Pathlib provides object-oriented path manipulation with intuitive methods and operators. The `/` operator for path joining is more readable than `os.path.join()`. Built-in methods like `.read_text()`, `.write_text()`, and `.glob()` simplify common operations. Pathlib is cross-platform and type-safe.

**Example:**

```python
from pathlib import Path
from typing import Iterator, Optional

# Use Path objects
config_path = Path("config.json")
data_dir = Path("data")

# Path operations
if config_path.exists():
    content = config_path.read_text()

if data_dir.is_dir():
    for file in data_dir.glob("*.txt"):
        process_file(file)

# Path construction
user_dir = Path("users") / "123" / "profile"
user_dir.mkdir(parents=True, exist_ok=True)

# Read/write operations
def save_data(path: Path, data: str) -> None:
    path.write_text(data)

def load_data(path: Path) -> str:
    return path.read_text()

# Iterate directory
def find_python_files(directory: Path) -> Iterator[Path]:
    return directory.rglob("*.py")

# Safe file operations
def read_if_exists(path: Path) -> Optional[str]:
    if path.exists() and path.is_file():
        return path.read_text()
    return None

# ❌ Avoid os.path
import os
old_path = os.path.join("users", "123", "profile")  # Don't do this

# ✅ Use Path instead
new_path = Path("users") / "123" / "profile"
```

**Techniques:**
- Create Path objects: `path = Path("file.txt")`
- Use `/` operator for path joining: `Path("dir") / "file.txt"`
- Check existence with `.exists()`, `.is_file()`, `.is_dir()`
- Read/write with `.read_text()`, `.write_text()`, `.read_bytes()`, `.write_bytes()`
- Glob patterns with `.glob()` and `.rglob()`
- Create directories with `.mkdir(parents=True, exist_ok=True)`
