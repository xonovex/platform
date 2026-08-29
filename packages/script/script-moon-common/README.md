# moon-scripts-common

Use this internal library for utilities shared by Moon task script packages, including command-line argument parsing, logging, `package.json` access, workspace root detection, and Moon project queries.

## Exports

Import the function that owns the required script concern.

- `parseCliArgs` - CLI argument parsing with type definitions
- `logError`, `logInfo`, `logSuccess`, `logWarning` - logging utilities
- `readPkg`, `writePkg` - read/write `package.json` files
- `findWorkspaceRoot` - locate the workspace root directory
- `findAllPackageJsonPaths`, `queryMoonProjects` - query Moon projects and package paths
