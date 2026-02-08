# tsconfig: TypeScript Configuration for TSTL

**Guideline:** Configure TypeScript with strict mode and TSTL settings for optimal Lua generation.

**Rationale:** Strict options catch errors early and generate predictable Lua. TSTL settings control targeting, libraries, and debugging.

**Example:**

```json
{
  "compilerOptions": {
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "strictFunctionTypes": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "target": "ES2020",
    "module": "esnext",
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "outDir": "./dist",
    "rootDir": "./src",
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true
  },
  "tstl": {
    "luaTarget": "JIT",
    "luaLibImport": "require",
    "sourceMapTraceback": true,
    "noHeader": false
  }
}
```

**Techniques:**

- Enable `strict: true` for comprehensive type checking
- Enable individual strict options: `noImplicitAny`, `strictNullChecks`, `strictFunctionTypes`
- Enable `noUnusedLocals` and `noUnusedParameters` to catch dead code
- Enable `noImplicitReturns` to ensure all code paths return
- Enable `noFallthroughCasesInSwitch` for safer switch statements
- Set `luaTarget` to appropriate runtime: "JIT" for LuaJIT, "5.3" for Lua 5.3, etc.
- Set `luaLibImport` to "require" or "inline" for library loading strategy
- Enable `sourceMapTraceback` for debugging information and stack traces
