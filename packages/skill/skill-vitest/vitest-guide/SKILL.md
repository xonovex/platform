---
name: vitest-guide
description: "Use when writing or editing Vitest 3+ tests in TypeScript. Triggers on `.test.ts`, `.spec.ts`, `vitest.config.*` files, and prompts about test setup, mocking, HTTP assertions, CORS preflight tests, type safety in tests, or snapshots, even when the user doesn't say 'Vitest'."
---

# Vitest Testing Guidelines

## Requirements

- Vitest ≥ 3, TypeScript ≥ 5.8

## Essentials

- **Type safety** - Define response interfaces, cast JSON results, see [references/type-safety.md](references/type-safety.md), [references/json-response-type-safety.md](references/json-response-type-safety.md)
- **HTTP testing** - Use HTTP 204 for OPTIONS, assert correct status codes, see [references/http-testing.md](references/http-testing.md), [references/cors-preflight-status-code.md](references/cors-preflight-status-code.md)
- **Timestamp testing** - Avoid flaky comparisons, verify existence or add delays, see [references/timestamp-testing.md](references/timestamp-testing.md)
- **Mock patterns** - Use simple type casting instead of complex generics, see [references/mock-patterns.md](references/mock-patterns.md)
- **TypeScript config** - Verify project reference paths match structure, see [references/typescript-config.md](references/typescript-config.md), [references/project-references-path-resolution.md](references/project-references-path-resolution.md)
- **Test organization** - Organize by endpoint/feature with nested describe blocks, see [references/test-organization.md](references/test-organization.md)

## Gotchas

- Mocks declared with `vi.mock(path)` are hoisted to the top of the file — referencing imported variables in the factory throws at hoist time
- Vitest transforms (`vite-node`) differ from Jest — `__dirname`/`__filename` work in CommonJS but not ESM tests without polyfills
- `vi.spyOn` returns the spy; `vi.fn` creates a new mock — confusing them passes type checks but breaks call-tracking assertions
- `expect.assertions(n)` in async tests catches missed awaits — without it, a forgotten `await` lets the test pass spuriously
- Watch mode caches module graphs; changing `vitest.config.ts` requires a full restart to pick up new transforms

## Progressive disclosure

- Read [references/type-safety.md](references/type-safety.md) - Load when test variables lose type information
- Read [references/json-response-type-safety.md](references/json-response-type-safety.md) - Load when calling res.json() without type assertions
- Read [references/http-testing.md](references/http-testing.md) - Load when asserting HTTP status codes in API tests
- Read [references/cors-preflight-status-code.md](references/cors-preflight-status-code.md) - Load when testing CORS OPTIONS requests
- Read [references/timestamp-testing.md](references/timestamp-testing.md) - Load when tests fail intermittently due to timing
- Read [references/mock-patterns.md](references/mock-patterns.md) - Load when creating mocks or stubs for tests
- Read [references/typescript-config.md](references/typescript-config.md) - Load when test files aren't recognized by TypeScript
- Read [references/project-references-path-resolution.md](references/project-references-path-resolution.md) - Load when imports fail in test files
- Read [references/test-organization.md](references/test-organization.md) - Load when structuring test suites for large APIs
