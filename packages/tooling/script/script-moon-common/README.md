# moon-scripts-common

Use this internal library for code shared by the Moon task script packages. The modules at the top of `src/` are the pure core and the ports; `src/adapters/` holds the implementations of the ports, so the unit tier drives the core with in-memory adapters and only the integration tier reaches the host.

## Exports

Import the module that owns the required script concern.

- `file-system` - the `FileSystem` port; `adapters/file-system-node` reaches the real disk, `adapters/file-system-memory` holds a tree in memory
- `moon-query` - `parseMoonProjects` and `packageJsonPaths`; `adapters/moon-query-node` runs `moon query projects`
- `adapters/executable-node` - `resolveExecutable` finds a command on the PATH of the current process
- `adapters/child-process` - `terminateProcess` ends a child process
- `cli-args` - `parseCliArgs` returns the parsed arguments or the help text
- `logging` - `logError`, `logInfo`, `logSuccess`, `logWarning`
- `package-json` - `readPkg`, `writePkg`
- `workspace` - `findWorkspaceRoot` locates the workspace root directory
- `skill-directory`, `claude-plugin`, `package-layout` - skill and plugin directory layout
