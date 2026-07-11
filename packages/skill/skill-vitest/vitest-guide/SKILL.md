---
name: vitest-guide
description: "Use when writing or editing Vitest 3+ tests in TypeScript. Triggers on `.test.ts`, `.spec.ts`, `vitest.config.*` files, and prompts about test setup, mocking, HTTP assertions, CORS preflight tests, type safety in tests, or snapshots, even when the user doesn't say 'Vitest'."
---

# Vitest Testing Guidelines

## Requirements

- Vitest ≥ 3, TypeScript ≥ 5.8

## Essentials

- **Type safety** - Cast `res.json()` to declared interfaces (references/type-safety.md)
- **HTTP testing** - Assert the status the middleware actually sends (references/http-testing.md)
- **Timestamp testing** - Avoid flaky comparisons; verify existence or add delays (references/timestamp-testing.md)
- **Mock patterns** - `vi.fn()` already returns a `Mock`; skip casts and generics (references/mock-patterns.md)
- **TypeScript config** - Include test paths; verify project reference levels (references/typescript-config.md)
- **Test organization** - Mirror API structure under `test/` with nested describe blocks (references/test-organization.md)

## Gotchas

- Mocks declared with `vi.mock(path)` are hoisted to the top of the file — referencing imported variables in the factory throws at hoist time
- Vitest transforms (`vite-node`) differ from Jest — `__dirname`/`__filename` work in CommonJS but not ESM tests without polyfills
- `vi.spyOn` returns the spy; `vi.fn` creates a new mock — confusing them passes type checks but breaks call-tracking assertions
- `expect.assertions(n)` in async tests catches missed awaits — without it, a forgotten `await` lets the test pass spuriously
- Watch mode caches module graphs; changing `vitest.config.ts` requires a full restart to pick up new transforms

## Progressive disclosure

- Read [references/type-safety.md](references/type-safety.md) - Load when calling `res.json()` or test variables lose type information
- Read [references/http-testing.md](references/http-testing.md) - Load when asserting HTTP status codes or testing CORS OPTIONS requests
- Read [references/timestamp-testing.md](references/timestamp-testing.md) - Load when tests fail intermittently due to timing
- Read [references/mock-patterns.md](references/mock-patterns.md) - Load when creating mocks or stubs for tests
- Read [references/typescript-config.md](references/typescript-config.md) - Load when test files aren't recognized by TypeScript or reference imports fail
- Read [references/test-organization.md](references/test-organization.md) - Load when structuring test suites for large APIs
