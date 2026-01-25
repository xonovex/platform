# @xonovex/tool-lib

Core library functions for Xonovex TypeScript scripts running with Node.js.

## Overview

This package provides common utilities used across all TypeScript scripts in the Xonovex platform:

- **Colors**: ANSI color codes for terminal output
- **Logging**: Structured logging with color-coded levels
- **Platform Detection**: OS detection and platform-specific commands
- **Error Handling**: Graceful error handling and validation
- **Path Utilities**: File system navigation and platform root detection

## Installation

This package is designed to be imported directly by other script packages in the monorepo workspace.

## Usage

```typescript
import {
  die,
  getPlatformRoot,
  logDebug,
  logError,
  logInfo,
  logSuccess,
  logWarning,
  printSection,
  requireCommand,
} from "@xonovex/tool-lib";

// Logging
logInfo("Starting operation...");
logSuccess("Operation completed!");
logWarning("This is a warning");
logError("An error occurred");
logDebug("Debug information"); // Only shown if DEBUG env var is set

// Sections
printSection("Configuration", "Loading platform configuration...");

// Platform detection
const root = await getPlatformRoot();
logInfo("Platform root:", root);

// Error handling
await requireCommand("git", "git");
die("Fatal error occurred", 1);
```

## API Reference

### Logging Functions

- `logInfo(...args)` - Log info message (blue)
- `logSuccess(...args)` - Log success message (green)
- `logWarning(...args)` - Log warning message (yellow)
- `logError(...args)` - Log error message (red)
- `logDebug(...args)` - Log debug message (purple, only if DEBUG is set)
- `printSection(title, content?)` - Print formatted section header
- `printSubsection(title)` - Print subsection header
- `checkResult(name, status, details?)` - Print check result with color

### Platform Detection

- `isMacOS()` - Check if running on macOS
- `isLinux()` - Check if running on Linux
- `isWindows()` - Check if running on Windows
- `getPlatformCommand(macCmd, linuxCmd, winCmd?)` - Get OS-specific command
- `getOS()` - Get current OS name

### Error Handling

- `die(message, exitCode?)` - Exit with error message
- `requireCommand(cmd, package?)` - Verify command exists
- `requireFile(path, description?)` - Verify file exists
- `requireDirectory(path, description?)` - Verify directory exists
- `validateInArray(value, array)` - Check if value is in array
- `validateBoolean(value, varName)` - Validate boolean value
- `validateRepository(repo, platformRoot?)` - Validate git repository

### Path Utilities

- `getScriptDir(importMeta)` - Get script directory path
- `getPlatformRoot(startDir?)` - Find platform root directory
- `getGitRoot()` - Get git repository root
- `fileExists(path)` - Check if file exists
- `dirExists(path)` - Check if directory exists
- `getFileMtime(path)` - Get file modification time
- `formatTimestamp(date)` - Format date for display
- `findClusterDirectory(platformRoot?)` - Find cluster directory
- `findInfrastructureDirectory(platformRoot?)` - Find infrastructure directory
- `detectAvailableEnvironments(infraDir?)` - List available environments

### Colors

```typescript
import {Colors} from "@xonovex/tool-lib";

console.log(`${Colors.GREEN}Success!${Colors.NC}`);
console.log(`${Colors.RED}Error!${Colors.NC}`);
```

Available colors:

- `RED`, `GREEN`, `YELLOW`, `BLUE`, `CYAN`, `PURPLE`, `NC` (no color)

## Environment Variables

- `DEBUG` - Enable debug logging
- `PLATFORM_ROOT` - Override platform root detection

## Examples

### Simple Script

```typescript
#!/usr/bin/env node
import {getPlatformRoot, logInfo, logSuccess} from "@xonovex/tool-lib";

async function main() {
  const root = await getPlatformRoot();
  logInfo("Working in:", root);

  // Do something...

  logSuccess("Done!");
}

main();
```

### Script with Error Handling

```typescript
#!/usr/bin/env node
import {
  die,
  logError,
  logInfo,
  requireCommand,
  requireDirectory,
} from "@xonovex/tool-lib";

async function main() {
  try {
    await requireCommand("git");
    await requireDirectory("./cluster", "cluster directory");

    logInfo("All checks passed!");
  } catch (error) {
    logError("Validation failed:", error);
    die("Cannot continue", 1);
  }
}

main();
```

## Development

```bash
# Build
npm run build

# Type check
npm run check

# Format
npm run fmt

# Lint
npm run lint

# Test
npm test
```
