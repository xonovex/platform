---
name: zod-guide
description: "Use when defining or editing Zod 4.0+ schemas for runtime validation in TypeScript. Triggers on `.ts` files with `zod` imports and prompts about API input validation, schema composition, type inference (`z.infer`), `z.uuid()`, `z.email()`, `z.iso.datetime()`, defaults, or refinements — even when the user doesn't say 'Zod'."
---

# Zod Coding Guidelines

## Requirements

- Zod ≥ 4.0

## Essentials

- **Use v4 validators** - `z.uuid()`, `z.email()`, `z.iso.datetime()` (not `z.string().uuid()`), see [references/migration-v4.md](references/migration-v4.md)
- **Name schemas PascalCase** - e.g. `UserSchema`, see [references/schema-organization.md](references/schema-organization.md)
- **Module-level schemas** - Define for reuse, keep close to usage, see [references/schema-organization.md](references/schema-organization.md)
- **Multi-step validation** - Chain `.transform()` into `.pipe()` before `.default()`, see [references/validation-patterns.md](references/validation-patterns.md)

## Gotchas

- `optional()` makes a field accept `undefined`; `nullable()` accepts `null` — they're orthogonal, combine with `.nullish()` for both
- Refinements (`.refine`) don't narrow the output type — TypeScript still sees the broader input type unless you pair with `.transform`

## Progressive Disclosure

### Guidelines

- Read [references/schema-organization.md](references/schema-organization.md) - Load when naming or placing schemas across modules or files
- Read [references/validation-patterns.md](references/validation-patterns.md) - Load when composing multi-step validation with `.transform()` and `.pipe()`

### Migration from Zod v3

- Read [references/migration-v4.md](references/migration-v4.md) - Load when migrating from Zod v3, replacing z.string().uuid() patterns, or using deprecated validators
