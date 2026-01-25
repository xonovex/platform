/**
 * Logging functions for Xonovex platform TypeScript scripts
 */

import { Colors } from "./colors.js";

export type LogLevel = "info" | "success" | "warning" | "error" | "debug";

/**
 * Log info message with blue color
 */
export function logInfo(...args: unknown[]): void {
  console.log(`${Colors.BLUE}[INFO]${Colors.NC}`, ...args);
}

/**
 * Log success message with green color
 */
export function logSuccess(...args: unknown[]): void {
  console.log(`${Colors.GREEN}[SUCCESS]${Colors.NC}`, ...args);
}

/**
 * Log warning message with yellow color
 */
export function logWarning(...args: unknown[]): void {
  console.log(`${Colors.YELLOW}[WARNING]${Colors.NC}`, ...args);
}

/**
 * Log error message with red color
 */
export function logError(...args: unknown[]): void {
  console.error(`${Colors.RED}[ERROR]${Colors.NC}`, ...args);
}

/**
 * Log debug message with purple color (only if DEBUG env var is set)
 */
export function logDebug(...args: unknown[]): void {
  if (process.env.DEBUG) {
    console.error(`${Colors.PURPLE}[DEBUG]${Colors.NC}`, ...args);
  }
}

/**
 * Print a formatted section header
 */
export function printSection(title: string, content?: string): void {
  console.log(`\n${Colors.CYAN}${title}${Colors.NC}`);
  console.log(`${Colors.CYAN}${"=".repeat(title.length)}${Colors.NC}`);
  if (content) {
    console.log(content);
  }
}

/**
 * Print a subsection header
 */
export function printSubsection(title: string): void {
  console.log(`\n${Colors.BLUE}${title}${Colors.NC}`);
  if (title.length > 0) {
    console.log(`${Colors.BLUE}${"-".repeat(title.length)}${Colors.NC}`);
  }
}

export type CheckStatus =
  | "PASS"
  | "OK"
  | "SUCCESS"
  | "FAIL"
  | "ERROR"
  | "FAILED"
  | "WARN"
  | "WARNING"
  | "INFO";

/**
 * Print a check result with appropriate color
 */
export function checkResult(
  checkName: string,
  status: CheckStatus,
  details?: string,
): void {
  let symbol = "•";
  let color: string = Colors.NC;

  switch (status) {
    case "PASS":
    case "OK":
    case "SUCCESS": {
      symbol = "✓";
      color = Colors.GREEN;
      break;
    }
    case "FAIL":
    case "ERROR":
    case "FAILED": {
      symbol = "✗";
      color = Colors.RED;
      break;
    }
    case "WARN":
    case "WARNING": {
      symbol = "⚠";
      color = Colors.YELLOW;
      break;
    }
    case "INFO": {
      symbol = "ℹ";
      color = Colors.BLUE;
      break;
    }
  }

  console.log(`${color}${symbol} ${checkName}${Colors.NC}`);

  if (details) {
    console.log(`${Colors.PURPLE}    ${details}${Colors.NC}`);
  }
}
