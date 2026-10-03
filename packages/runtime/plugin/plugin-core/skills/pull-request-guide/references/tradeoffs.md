# tradeoffs: Surface Limitations and Decisions Early

Add a short "Tradeoffs / risks" section stating each known limitation and why it is acceptable, alternatives considered and why rejected, and anything affecting backwards compatibility, data, config, or other consumers. If the change diverges from a related convention or branch, explain why. Do not invent risks - a change with none needs no section.

```text
## Tradeoffs / risks
- We create the queue names the service actually reads; the legacy script created a differently-named, unused queue.
- Local-dev tooling only, production images are unaffected.
```

## Related

[description.md](./description.md)
