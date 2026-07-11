# typescript-config: tsconfig for Tests

Add `test` and `vitest.config.ts` to tsconfig `include` or TypeScript won't type-check them. `references` paths are relative from the current package — count `..` levels against the real directory structure (verify with `ls`, test with `tsc --build`).

```json
{
  "extends": "../../tsconfig.base.json",
  "references": [{"path": "../../shared/shared-core"}],
  "include": ["src", "test", "vitest.config.ts"],
  "exclude": ["dist", "node_modules"]
}
```

From `packages/templates/X`, `../../shared/shared-core` is correct; `../../../shared/shared-core` (up 3) and `../../shared-core` (missing `shared/`) are the common off-by-one mistakes.
