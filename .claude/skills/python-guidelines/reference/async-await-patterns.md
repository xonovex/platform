# async-await-patterns: Async/Await Patterns

**Guideline:** Use async/await for I/O-bound operations with proper concurrent execution patterns.

**Rationale:** Async programming enables efficient I/O operations without blocking threads. Using `asyncio.gather()` allows multiple operations to run concurrently, improving performance. Async generators enable memory-efficient streaming of data. Proper async context management ensures resources are cleaned up correctly.

**Example:**

```python
import asyncio
import aiohttp
from typing import Sequence

# Async function with type hints
async def fetch_url(session: aiohttp.ClientSession, url: str) -> str:
    async with session.get(url) as response:
        response.raise_for_status()
        return await response.text()

# Async context manager
async def fetch_multiple(urls: Sequence[str]) -> list[str]:
    async with aiohttp.ClientSession() as session:
        tasks = [fetch_url(session, url) for url in urls]
        return await asyncio.gather(*tasks)

# Async generator
async def fetch_paginated(
    session: aiohttp.ClientSession,
    base_url: str,
    page_size: int = 100
):
    page = 1
    while True:
        url = f"{base_url}?page={page}&size={page_size}"
        data = await fetch_url(session, url)

        if not data:
            break

        yield data
        page += 1

# Usage
async def main():
    urls = ["https://api.example.com/1", "https://api.example.com/2"]
    results = await fetch_multiple(urls)

    async with aiohttp.ClientSession() as session:
        async for page_data in fetch_paginated(session, "https://api.example.com/items"):
            process_page(page_data)

# Run
asyncio.run(main())
```

**Techniques:**
- Define async functions with `async def`
- Use `await` for async operations
- Use `async with` for async context managers
- Gather concurrent operations with `asyncio.gather(*tasks)`
- Create async generators with `async def` and `yield`
- Run async code with `asyncio.run(main())`
