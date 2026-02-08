---
name: vitest-guidelines
description: Trigger on `.test.ts`, `.spec.ts` files, test blocks. Use when writing Vitest 3+ tests with TypeScript. Apply for type safety in tests, HTTP testing, mocking patterns, test organization. Keywords: Vitest, test, spec, HTTP assertions, mock patterns, type safety, CORS preflight, timestamps, response casting.
---

# Vitest Testing Guidelines

## Requirements

- Vitest ≥ 3, TypeScript ≥ 5.8

## Essentials

- **Type safety** - Define response interfaces, cast JSON results, see [reference/type-safety.md](reference/type-safety.md), [reference/json-response-type-safety.md](reference/json-response-type-safety.md)
- **HTTP testing** - Use HTTP 204 for OPTIONS, assert correct status codes, see [reference/http-testing.md](reference/http-testing.md), [reference/cors-preflight-status-code.md](reference/cors-preflight-status-code.md)
- **Timestamp testing** - Avoid flaky comparisons, verify existence or add delays, see [reference/timestamp-testing.md](reference/timestamp-testing.md)
- **Mock patterns** - Use simple type casting instead of complex generics, see [reference/mock-patterns.md](reference/mock-patterns.md)
- **TypeScript config** - Verify project reference paths match structure, see [reference/typescript-config.md](reference/typescript-config.md), [reference/project-references-path-resolution.md](reference/project-references-path-resolution.md)
- **Test organization** - Organize by endpoint/feature with nested describe blocks, see [reference/test-organization.md](reference/test-organization.md)

## Progressive disclosure

- Read [reference/type-safety.md](reference/type-safety.md) - When test variables lose type information
- Read [reference/json-response-type-safety.md](reference/json-response-type-safety.md) - When calling res.json() without type assertions
- Read [reference/http-testing.md](reference/http-testing.md) - When asserting HTTP status codes in API tests
- Read [reference/cors-preflight-status-code.md](reference/cors-preflight-status-code.md) - When testing CORS OPTIONS requests
- Read [reference/timestamp-testing.md](reference/timestamp-testing.md) - When tests fail intermittently due to timing
- Read [reference/mock-patterns.md](reference/mock-patterns.md) - When creating mocks or stubs for tests
- Read [reference/typescript-config.md](reference/typescript-config.md) - When test files aren't recognized by TypeScript
- Read [reference/project-references-path-resolution.md](reference/project-references-path-resolution.md) - When imports fail in test files
- Read [reference/test-organization.md](reference/test-organization.md) - When structuring test suites for large APIs
