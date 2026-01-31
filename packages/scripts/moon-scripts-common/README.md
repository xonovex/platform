# moon-scripts-common

Shared utility library used by other moon script packages. Provides CLI argument parsing, logging, package.json utilities, workspace root detection, and Moon project querying.

## Exports

- `parseCliArgs` — CLI argument parsing with type definitions
- `logError`, `logInfo`, `logSuccess`, `logWarning` — logging utilities
- `readPkg`, `writePkg` — read/write `package.json` files
- `findWorkspaceRoot` — locate the workspace root directory
- `findAllPackageJsonPaths`, `queryMoonProjects` — query Moon projects and package paths
