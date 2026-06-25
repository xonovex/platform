---
name: typescript-guide
description: "Use when editing or reviewing TypeScript in Node.js ESM projects. Triggers on `.ts` files, ESM `package.json`, and prompts about async functions, type inference, strict mode, ESM imports, env handling, template literals, even when the user doesn't say 'TypeScript'."
---

# TypeScript Coding Guidelines

## Requirements

- Node.js ESM, TypeScript ≥ 5.8, Vitest ≥ 3, Zod ≥ 4.

## Essentials

- **Strict mode** - Enable strict flags, avoid `@ts-ignore` without comment+issue
- **Type safety** - Explicit public types, `unknown` over `any`, derive from generated types
- **Imports** - Named ESM imports, `import type` for types, direct from source, see [references/avoid-barrel-exports.md](references/avoid-barrel-exports.md), [references/avoid-reexports.md](references/avoid-reexports.md)
- **Async/await** - Handle errors explicitly, only use `async` with `await`, see [references/async-without-await.md](references/async-without-await.md), [references/unnecessary-async-keywords.md](references/unnecessary-async-keywords.md)
- **Immutability** - Use `const`/`readonly` where possible
- **Validation** - Zod for I/O boundaries, infer types from schemas
- **Linting** - Never suppress with eslint-disable, fix root causes, see [references/avoid-eslint-disable.md](references/avoid-eslint-disable.md)
- **Template literals** - Convert numbers with `String(value)`, see [references/template-literals-require-string-conversion.md](references/template-literals-require-string-conversion.md)
- **Numeric literals** - Use underscores in large numbers (`30_000`), see [references/numeric-separator-enforcement.md](references/numeric-separator-enforcement.md)
- **Method references** - Keep object references to maintain `this` binding, see [references/unbound-method-references.md](references/unbound-method-references.md)
- **Environment** - Use dot notation for `process.env` access, see [references/env-access-bracket-notation.md](references/env-access-bracket-notation.md)
- **Paradigm** - Functional style → **general-fp-guide**; class/OO design → **general-oop-guide**

## Gotchas

- Structural typing means `{ a: string }` accepts `{ a: string, b: number }` silently — explicit `satisfies` is the way to catch unintended extras
- `as` casts bypass the type system without check — prefer narrowing functions (type predicates) over casts
- `type` vs `interface`: interfaces merge across declarations, types don't — `declare global { interface Window { … } }` works, `type Window = …` doesn't
- `unknown` is the safer `any` — but it doesn't propagate; narrowing once doesn't carry across assignments
- Module resolution depends on `tsconfig` `moduleResolution` (`bundler` vs `node16` vs `nodenext`) — wrong choice silently breaks deep imports

## Progressive disclosure

- Read [references/async-without-await.md](references/async-without-await.md) - Load when seeing async functions that don't use await
- Read [references/unnecessary-async-keywords.md](references/unnecessary-async-keywords.md) - Load when simplifying synchronous controller functions
- Read [references/avoid-eslint-disable.md](references/avoid-eslint-disable.md) - Load when tempted to suppress linting warnings
- Read [references/template-literals-require-string-conversion.md](references/template-literals-require-string-conversion.md) - Load when inserting numbers in template literals
- Read [references/numeric-separator-enforcement.md](references/numeric-separator-enforcement.md) - Load when writing large numeric literals
- Read [references/unbound-method-references.md](references/unbound-method-references.md) - Load when passing methods as callbacks or references
- Read [references/avoid-barrel-exports.md](references/avoid-barrel-exports.md) - Load when creating or importing from index.ts files
- Read [references/avoid-reexports.md](references/avoid-reexports.md) - Load when organizing module exports
- Read [references/env-access-bracket-notation.md](references/env-access-bracket-notation.md) - Load when accessing process.env variables
