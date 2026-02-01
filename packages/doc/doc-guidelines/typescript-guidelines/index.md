---
name: typescript-guidelines
description: >-
  Trigger on `.ts` files with Node.js/ESM patterns. Use when writing TypeScript code for Node.js ESM projects. Apply for async functions, module imports, type safety, linting. Keywords: TypeScript, ESM, strict mode, async/await, Zod, type inference, avoid barrel exports, process.env, template literals, numeric separators.
---

# TypeScript Coding Guidelines

## Requirements

- Node.js ESM, TypeScript ≥ 5.8, Vitest ≥ 3, Zod ≥ 4.

## Essentials

- **Strict mode** - Enable strict flags, avoid `@ts-ignore` without comment+issue
- **Type safety** - Explicit public types, `unknown` over `any`, derive from generated types
- **Imports** - Named ESM imports, `import type` for types, direct from source, see [reference/avoid-barrel-exports.md](reference/avoid-barrel-exports.md), [reference/avoid-reexports.md](reference/avoid-reexports.md)
- **Async/await** - Handle errors explicitly, only use `async` with `await`, see [reference/async-without-await.md](reference/async-without-await.md), [reference/unnecessary-async-keywords.md](reference/unnecessary-async-keywords.md)
- **Immutability** - Use `const`/`readonly` where possible
- **Validation** - Zod for I/O boundaries, infer types from schemas
- **Linting** - Never suppress with eslint-disable, fix root causes, see [reference/avoid-eslint-disable.md](reference/avoid-eslint-disable.md)
- **Template literals** - Convert numbers with `String(value)`, see [reference/template-literals-require-string-conversion.md](reference/template-literals-require-string-conversion.md)
- **Numeric literals** - Use underscores in large numbers (`30_000`), see [reference/numeric-separator-enforcement.md](reference/numeric-separator-enforcement.md)
- **Method references** - Keep object references to maintain `this` binding, see [reference/unbound-method-references.md](reference/unbound-method-references.md)
- **Environment** - Use dot notation for `process.env` access, see [reference/env-access-bracket-notation.md](reference/env-access-bracket-notation.md)

## Progressive disclosure

- Read [reference/async-without-await.md](reference/async-without-await.md) - When seeing async functions that don't use await
- Read [reference/unnecessary-async-keywords.md](reference/unnecessary-async-keywords.md) - When simplifying synchronous controller functions
- Read [reference/avoid-eslint-disable.md](reference/avoid-eslint-disable.md) - When tempted to suppress linting warnings
- Read [reference/template-literals-require-string-conversion.md](reference/template-literals-require-string-conversion.md) - When inserting numbers in template literals
- Read [reference/numeric-separator-enforcement.md](reference/numeric-separator-enforcement.md) - When writing large numeric literals
- Read [reference/unbound-method-references.md](reference/unbound-method-references.md) - When passing methods as callbacks or references
- Read [reference/avoid-barrel-exports.md](reference/avoid-barrel-exports.md) - When creating or importing from index.ts files
- Read [reference/avoid-reexports.md](reference/avoid-reexports.md) - When organizing module exports
- Read [reference/env-access-bracket-notation.md](reference/env-access-bracket-notation.md) - When accessing process.env variables
