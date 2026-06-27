---
name: python-guide
description: "Use when writing or editing Python 3.12+ for APIs, data processing, scripting, or tooling. Triggers on `.py` files and prompts about async functions, type hints, dataclasses, pathlib, pytest, generators, context managers, f-strings, even when the user doesn't say 'Python'."
---

# Python Coding Guidelines

## Requirements

- Python ≥ 3.12; pytest ≥ 8.

## Essentials

- **Data models** - Use dataclasses and type hints, see [references/dataclasses-type-hints.md](references/dataclasses-type-hints.md), [references/type-checking.md](references/type-checking.md)
- **Iteration** - Prefer generators/comprehensions for data processing, see [references/generators-comprehensions.md](references/generators-comprehensions.md)
- **Async I/O** - Use async/await for I/O operations, see [references/async-await-patterns.md](references/async-await-patterns.md)
- **Performance** - Cache pure functions with `@cache`, see [references/caching-functions.md](references/caching-functions.md)
- **Resource management** - Use context managers for cleanup, see [references/resource-management.md](references/resource-management.md)
- **Modern syntax** - Use pathlib, f-strings, specific exceptions, see [references/pathlib-file-ops.md](references/pathlib-file-ops.md), [references/string-formatting.md](references/string-formatting.md), [references/exception-handling.md](references/exception-handling.md)
- **Paradigm** - Functional style → **fp-guide**; class/OO design → **oop-guide**

## Gotchas

- Mutable default arguments (`def f(x=[]):`) share state across calls — use `None` and assign inside
- The GIL serializes pure-Python execution — threads only help on I/O; CPU-bound work needs `multiprocessing` or compiled extensions
- `is` checks identity, not equality — small-int caching means `a is b` works for `1` but fails for `300`
- `__init__.py` is no longer required for packages (PEP 420), but mixing namespace and regular packages causes silent import-shadowing bugs
- `async`/sync mixing without `asyncio.to_thread` blocks the event loop — a single `requests.get()` in an async handler kills concurrency

## Progressive disclosure

- Read [references/dataclasses-type-hints.md](references/dataclasses-type-hints.md) - Load when defining structured data models or adding type annotations
- Read [references/type-checking.md](references/type-checking.md) - Load when using Protocols, type aliases, or complex Union types
- Read [references/async-await-patterns.md](references/async-await-patterns.md) - Load when building async APIs or handling concurrent I/O operations
- Read [references/resource-management.md](references/resource-management.md) - Load when working with files, connections, or resources needing cleanup
- Read [references/caching-functions.md](references/caching-functions.md) - Load when optimizing expensive computations or repeated function calls
- Read [references/generators-comprehensions.md](references/generators-comprehensions.md) - Load when processing large datasets or streaming data
- Read [references/string-formatting.md](references/string-formatting.md) - Load when formatting output, building messages, or templating
- Read [references/pathlib-file-ops.md](references/pathlib-file-ops.md) - Load when reading/writing files or traversing directories
- Read [references/exception-handling.md](references/exception-handling.md) - Load when defining error handling or creating custom exceptions
