# pathlib-file-ops: Pathlib for File Operations

Use `pathlib.Path` over `os.path`. Join with `/`; read/write via `.read_text()`/`.write_text()`/`.read_bytes()`/`.write_bytes()`; test with `.exists()`/`.is_file()`/`.is_dir()`; match with `.glob()`/`.rglob()`; create dirs with `.mkdir(parents=True, exist_ok=True)`.

```python
config = Path("config.json")
if config.is_file():
    content = config.read_text()

user_dir = Path("users") / "123" / "profile"
user_dir.mkdir(parents=True, exist_ok=True)

py_files = Path("src").rglob("*.py")
```
