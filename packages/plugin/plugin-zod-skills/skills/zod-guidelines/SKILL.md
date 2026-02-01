---
name: zod-guidelines
description: >-
  Trigger on `.ts` files with Zod imports or validation patterns. Use when working with Zod 4.0+ for runtime validation. Apply for API validation, schema composition, type inference, default values. Keywords: Zod, z.uuid(), z.email(), z.iso.datetime(), schema composition, safeParse, z.infer, defaults, validation patterns.
---

# Zod Coding Guidelines

## Requirements

- Zod ≥ 4.0

## Essentials

- **Use v4 validators** - `z.uuid()`, `z.email()`, `z.iso.datetime()` (not `z.string().uuid()`)
- **Name schemas PascalCase** - Derive with `.omit()`, `.extend()`, `.pick()`, `.merge()`, see [reference/schema-organization.md](reference/schema-organization.md)
- **Validate at boundaries** - Use `.safeParse()` for I/O, `.pipe()` for transforms, see [reference/validation-patterns.md](reference/validation-patterns.md)
- **Infer types from schemas** - `z.infer<typeof Schema>` for TypeScript types, see [reference/validation-patterns.md](reference/validation-patterns.md)
- **Module-level schemas** - Define for reuse, keep close to usage, see [reference/schema-organization.md](reference/schema-organization.md)
- **Match default output types** - Defaults must align with transformations, see [reference/default-values-output-type.md](reference/default-values-output-type.md)

## Progressive Disclosure

### Guidelines

- Read [reference/schema-organization.md](reference/schema-organization.md) - When organizing schemas across modules or files
- Read [reference/validation-patterns.md](reference/validation-patterns.md) - When choosing between safeParse, parse, or pipe
- Read [reference/default-values-output-type.md](reference/default-values-output-type.md) - When default values cause type mismatches

### Migration from Zod v3

- Read [reference/migration-v4.md](reference/migration-v4.md) - When migrating from Zod v3 or using deprecated validators
- Read [reference/migration-string-validators.md](reference/migration-string-validators.md) - When replacing z.string().uuid() patterns
