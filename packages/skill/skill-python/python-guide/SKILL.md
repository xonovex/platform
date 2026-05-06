---
name: python-guide
description: "Use when writing or editing Python 3.12+ for APIs, data processing, scripting, or tooling. Triggers on `.py` files and prompts about async functions, type hints, dataclasses, pathlib, pytest, generators, context managers, f-strings, even when the user doesn't say 'Python'. Skip Python 2 code, notebook-only data exploration, and ML model code that needs framework-specific guidance."
---

# Python Coding Guidelines

## Requirements

- Python ≥ 3.12; pytest ≥ 8.

## Essentials

- **Data models** - Use dataclasses and type hints, see [reference/dataclasses-type-hints.md](reference/dataclasses-type-hints.md), [reference/type-checking.md](reference/type-checking.md)
- **Iteration** - Prefer generators/comprehensions for data processing, see [reference/generators-comprehensions.md](reference/generators-comprehensions.md)
- **Async I/O** - Use async/await for I/O operations, see [reference/async-await-patterns.md](reference/async-await-patterns.md)
- **Performance** - Cache pure functions with `@cache`, see [reference/caching-functions.md](reference/caching-functions.md)
- **Resource management** - Use context managers for cleanup, see [reference/resource-management.md](reference/resource-management.md)
- **Modern syntax** - Use pathlib, f-strings, specific exceptions, see [reference/pathlib-file-ops.md](reference/pathlib-file-ops.md), [reference/string-formatting.md](reference/string-formatting.md), [reference/exception-handling.md](reference/exception-handling.md)

## Progressive disclosure

- Read [reference/dataclasses-type-hints.md](reference/dataclasses-type-hints.md) - When defining structured data models or adding type annotations
- Read [reference/type-checking.md](reference/type-checking.md) - When using Protocols, type aliases, or complex Union types
- Read [reference/async-await-patterns.md](reference/async-await-patterns.md) - When building async APIs or handling concurrent I/O operations
- Read [reference/resource-management.md](reference/resource-management.md) - When working with files, connections, or resources needing cleanup
- Read [reference/caching-functions.md](reference/caching-functions.md) - When optimizing expensive computations or repeated function calls
- Read [reference/generators-comprehensions.md](reference/generators-comprehensions.md) - When processing large datasets or streaming data
- Read [reference/string-formatting.md](reference/string-formatting.md) - When formatting output, building messages, or templating
- Read [reference/pathlib-file-ops.md](reference/pathlib-file-ops.md) - When reading/writing files or traversing directories
- Read [reference/exception-handling.md](reference/exception-handling.md) - When defining error handling or creating custom exceptions
