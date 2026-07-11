# tsconfig: TSTL Compiler Options

Run TS strict mode, then add the TSTL-specific `tstl` block that controls Lua emit:

- `luaTarget` — `"JIT"`, `"5.3"`, `"5.4"`, `"5.1"`, or `"universal"`.
- `luaLibImport` — `"require"` (default) or `"inline"` for the runtime library.
- `sourceMapTraceback` — remap Lua stack traces back to TS lines.
- `noHeader` — omit the generated-by header comment.

```json
{
  "compilerOptions": {
    "strict": true,
    "target": "ES2020",
    "module": "esnext",
    "outDir": "./dist",
    "rootDir": "./src"
  },
  "tstl": {
    "luaTarget": "JIT",
    "luaLibImport": "require",
    "sourceMapTraceback": true,
    "noHeader": false
  }
}
```
