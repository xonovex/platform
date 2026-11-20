# @xonovex/core

Use `@xonovex/core` for terminal colors and leveled logging in Xonovex TypeScript packages.

The package provides these utilities:

- **Colors**: ANSI color codes for terminal output
- **Logging**: a logger that writes through a sink it receives, so it reads no console or environment itself
- **Console logger**: the one function that wires a logger to a console and the `DEBUG` environment variable

## Installation

Import this internal package directly from another package in the monorepo. It is not a standalone application dependency.

## Usage

Build the logger once at the entry point and pass it to the code that logs.

```typescript
import {createConsoleLogger} from "@xonovex/core";

const logger = createConsoleLogger(console, process.env);

logger.info("Starting operation...");
logger.success("Operation completed!");
logger.warning("This is a warning");
logger.error("An error occurred");
logger.debug("Debug information"); // Only shown if DEBUG is set

logger.printSection("Configuration", "Loading platform configuration...");
logger.checkResult("lint", "PASS", "no findings");
```

Tests build a logger over an in-memory sink with `createLogger`.

```typescript
import {createLogger} from "@xonovex/core";

const lines: unknown[][] = [];
const logger = createLogger({
  sink: {
    out: (...args) => {
      lines.push(args);
    },
    err: (...args) => {
      lines.push(args);
    },
  },
  debug: false,
});
```

## API Reference

### Logger

- `createLogger({sink, debug})` - Build a logger over a `LogSink` with `out` and `err` functions
- `createConsoleLogger(console, environment)` - Build a logger over a console, with debug output enabled when `environment.DEBUG` is non-empty
- `logger.info(...args)` - Info message (blue, `out`)
- `logger.success(...args)` - Success message (green, `out`)
- `logger.warning(...args)` - Warning message (yellow, `out`)
- `logger.error(...args)` - Error message (red, `err`)
- `logger.debug(...args)` - Debug message (purple, `err`, only when the debug gate is open)
- `logger.printSection(title, content?)` - Print a section header
- `logger.printSubsection(title)` - Print a subsection header
- `logger.checkResult(name, status, details?)` - Print a check result with color

### Colors

Use `Colors` only when direct ANSI color control is required.

```typescript
import {Colors} from "@xonovex/core";

console.log(`${Colors.GREEN}Success!${Colors.NC}`);
```

Available colors:

- `RED`, `GREEN`, `YELLOW`, `BLUE`, `CYAN`, `PURPLE`, `NC` (no color)

## Environment Variables

- `DEBUG` - Enable debug logging when passed to `createConsoleLogger`

## Development

Run the package gates from this directory.

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
