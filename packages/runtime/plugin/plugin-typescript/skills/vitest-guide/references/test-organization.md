# test-organization: Test Suite Layout

Put tests under `test/` at the package root, mirroring the API structure; group endpoints with nested `describe` blocks and one `it` per behaviour (cover success and error cases).

```
test/
├── app.test.ts
├── health.test.ts
└── api/
    ├── users.test.ts
    └── auth.test.ts
```
