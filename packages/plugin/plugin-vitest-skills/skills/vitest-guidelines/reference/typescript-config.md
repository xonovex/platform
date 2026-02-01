# typescript-config: TypeScript Project Configuration for Tests

**Guideline:** Verify directory structure before configuring TypeScript project references and test inclusion in tsconfig.json.

**Rationale:** Incorrect configuration causes compilation failures and broken builds in CI/CD.

**Example:**

```json
// For packages/templates/X referencing packages/shared/Y
{
  "extends": "../../tsconfig.base.json",
  "compilerOptions": {
    "rootDir": "src",
    "outDir": "dist"
  },
  "references": [
    {
      "path": "../../shared/shared-core" // ✅ Correct - verified with ls
    },
    {
      "path": "../../shared/shared-types"
    }
  ],
  "include": [
    "src", // ✅ Source files
    "test", // ✅ Test files
    "vitest.config.ts" // ✅ Vitest config
  ],
  "exclude": [
    "dist", // ✅ Build output
    "node_modules" // ✅ Dependencies
  ]
}
```

**Techniques:**

- Check actual directory structure with `ls` or file explorer before configuring
- Calculate relative paths from current package to referenced packages
- Add project references with correct relative paths
- Include test directories in `include` array (e.g., "test")
- Exclude build artifacts and node_modules from compilation
- Test compilation with `tsc --build` to verify paths resolve
