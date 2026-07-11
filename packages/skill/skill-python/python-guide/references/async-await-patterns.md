# async-await-patterns: Async/Await Patterns

Use `async`/`await` for I/O-bound work. Run independent awaitables concurrently with `asyncio.gather(*tasks)` — not a sequential `await` loop. Stream with `async for` over an `async def` + `yield` generator; enter resources with `async with`; start from `asyncio.run(main())`.

```python
async def fetch_multiple(urls: Sequence[str]) -> list[str]:
    async with aiohttp.ClientSession() as session:
        return await asyncio.gather(*(fetch_url(session, u) for u in urls))

async def fetch_paginated(session, base_url: str) -> AsyncIterator[str]:
    page = 1
    while data := await fetch_url(session, f"{base_url}?page={page}"):
        yield data
        page += 1
```
