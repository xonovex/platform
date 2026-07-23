# @xonovex/core

Core library functions for Xonovex TypeScript scripts running with Node.js.

## Overview

This package provides common utilities used across all TypeScript scripts in the Xonovex platform:

- **Colors**: ANSI color codes for terminal output
- **Logging**: Structured logging with color-coded levels
- **Platform Detection**: OS detection and platform-specific commands
- **Error Handling**: Graceful error handling and validation
- **Path Utilities**: File system navigation and platform root detection
- **Skill Composition**: Deterministic exact and semantic skill selection with
  scoped preference overlays

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
} from "@xonovex/core";

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

### Skill Composition

Import the composition contract directly from
`@xonovex/core/skill-composition-contract`, primitive selectors from
`@xonovex/core/skill-selection`, preference resolution from
`@xonovex/core/skill-preference-overlays`, and graph composition from
`@xonovex/core/skill-composition`. The modules parse a versioned catalog from exact
source bytes, an installed skill inventory, and a request containing exact skills,
semantic requirements, provider bindings, overlay context, and preference overlays.
Resolution returns a unique dependency-first load order, aggregated selection
provenance, visible failures, effective/shadowed/skipped overlays, catalog identity,
and an overall `ready`, `degraded`, or `blocked` status.

The installed inventory may omit repository-local `packagePath` and
`sourcesPath` fields. Its exact plugin dependency edges carry installed versions.
Exact implementation versions and provision versions use SemVer. Required
resolution failures block; preferred failures remain in the result and degrade it.

The package also exposes a JSON CLI:

```bash
xonovex-skill-compose \
  composition-catalog.json \
  installed-skills.json \
  composition-request.json
```

Normalized workflow requests use the canonical adapter and installed-plugin
discovery:

```bash
xonovex-workflow-compose \
  --catalog composition-catalog.json \
  --request workflow-request.json \
  --installed-root ./installed-plugins
```

The workflow skill remains declarative and packages its catalog snapshot, but no
executable runtime. The host invokes this shared CLI with that snapshot and its
actual installed inventory.

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
import {Colors} from "@xonovex/core";

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
import {getPlatformRoot, logInfo, logSuccess} from "@xonovex/core";

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
} from "@xonovex/core";

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
