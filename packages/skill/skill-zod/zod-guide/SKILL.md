---
name: zod-guide
description: "Use when defining or editing Zod 4.0+ schemas for runtime validation in TypeScript. Triggers on `.ts` files with `zod` imports and prompts about API input validation, schema composition, type inference (`z.infer`), `z.uuid()`, `z.email()`, `z.iso.datetime()`, defaults, or refinements — even when the user doesn't say 'Zod'."
---

# Zod Coding Guidelines

## Requirements

- Zod ≥ 4.0

## Essentials

- **Use v4 validators** - `z.uuid()`, `z.email()`, `z.iso.datetime()` (not `z.string().uuid()`)
- **Name schemas PascalCase** - Derive with `.omit()`, `.extend()`, `.pick()`, see [references/schema-organization.md](references/schema-organization.md)
- **Validate at boundaries** - Use `.safeParse()` for I/O, `.pipe()` for transforms, see [references/validation-patterns.md](references/validation-patterns.md)
- **Infer types from schemas** - `z.infer<typeof Schema>` for TypeScript types, see [references/validation-patterns.md](references/validation-patterns.md)
- **Module-level schemas** - Define for reuse, keep close to usage, see [references/schema-organization.md](references/schema-organization.md)
- **Match default output types** - Defaults must align with transformations, see [references/default-values-output-type.md](references/default-values-output-type.md)

## Gotchas

- `safeParse` returns `{ success, data | error }` — using `parse` in handler code throws, breaking the response on invalid input
- Transforms change the output type — `z.string().transform(s => s.length)` infers as `number`, not `string`
- `optional()` makes a field accept `undefined`; `nullable()` accepts `null` — they're orthogonal, combine with `.nullish()` for both
- `default(x)` makes the input optional but the **output** type required — `z.infer` reflects the output side, which can surprise callers
- Refinements (`.refine`) don't narrow the output type — TypeScript still sees the broader input type unless you pair with `.transform`

## Progressive Disclosure

### Guidelines

- Read [references/schema-organization.md](references/schema-organization.md) - Load when organizing schemas across modules or files
- Read [references/validation-patterns.md](references/validation-patterns.md) - Load when choosing between safeParse, parse, or pipe
- Read [references/default-values-output-type.md](references/default-values-output-type.md) - Load when default values cause type mismatches

### Migration from Zod v3

- Read [references/migration-v4.md](references/migration-v4.md) - Load when migrating from Zod v3 or using deprecated validators
- Read [references/migration-string-validators.md](references/migration-string-validators.md) - Load when replacing z.string().uuid() patterns
