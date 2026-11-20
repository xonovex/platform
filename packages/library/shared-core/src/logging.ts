/**
 * Logging for Xonovex platform TypeScript scripts. The logger writes through a
 * sink it receives, so it owns no console or environment access.
 */

import {Colors} from "./colors.js";

export type LogLevel = "info" | "success" | "warning" | "error" | "debug";

/**
 * Where log lines go: `out` for results and sections, `err` for leveled
 * messages that are diagnostics.
 */
export interface LogSink {
  readonly out: (...args: unknown[]) => void;
  readonly err: (...args: unknown[]) => void;
}

export interface LoggerOptions {
  readonly sink: LogSink;
  /** Debug messages are dropped unless this is true. */
  readonly debug: boolean;
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

export interface Logger {
  readonly info: (...args: unknown[]) => void;
  readonly success: (...args: unknown[]) => void;
  readonly warning: (...args: unknown[]) => void;
  readonly error: (...args: unknown[]) => void;
  readonly debug: (...args: unknown[]) => void;
  readonly printSection: (title: string, content?: string) => void;
  readonly printSubsection: (title: string) => void;
  readonly checkResult: (
    checkName: string,
    status: CheckStatus,
    details?: string,
  ) => void;
}

interface CheckStyle {
  readonly symbol: string;
  readonly color: string;
}

function tag(color: string, label: string): string {
  return `${color}[${label}]${Colors.NC}`;
}

function checkStyle(status: CheckStatus): CheckStyle {
  switch (status) {
    case "PASS":
    case "OK":
    case "SUCCESS": {
      return {symbol: "✓", color: Colors.GREEN};
    }
    case "FAIL":
    case "ERROR":
    case "FAILED": {
      return {symbol: "✗", color: Colors.RED};
    }
    case "WARN":
    case "WARNING": {
      return {symbol: "⚠", color: Colors.YELLOW};
    }
    case "INFO": {
      return {symbol: "ℹ", color: Colors.BLUE};
    }
  }
}

/**
 * Build a logger over a sink.
 */
export function createLogger({sink, debug}: LoggerOptions): Logger {
  return {
    info: (...args) => {
      sink.out(tag(Colors.BLUE, "INFO"), ...args);
    },
    success: (...args) => {
      sink.out(tag(Colors.GREEN, "SUCCESS"), ...args);
    },
    warning: (...args) => {
      sink.out(tag(Colors.YELLOW, "WARNING"), ...args);
    },
    error: (...args) => {
      sink.err(tag(Colors.RED, "ERROR"), ...args);
    },
    debug: (...args) => {
      if (debug) {
        sink.err(tag(Colors.PURPLE, "DEBUG"), ...args);
      }
    },
    printSection: (title, content) => {
      sink.out(`\n${Colors.CYAN}${title}${Colors.NC}`);
      sink.out(`${Colors.CYAN}${"=".repeat(title.length)}${Colors.NC}`);
      if (content) {
        sink.out(content);
      }
    },
    printSubsection: (title) => {
      sink.out(`\n${Colors.BLUE}${title}${Colors.NC}`);
      if (title.length > 0) {
        sink.out(`${Colors.BLUE}${"-".repeat(title.length)}${Colors.NC}`);
      }
    },
    checkResult: (checkName, status, details) => {
      const {symbol, color} = checkStyle(status);
      sink.out(`${color}${symbol} ${checkName}${Colors.NC}`);
      if (details) {
        sink.out(`${Colors.PURPLE}    ${details}${Colors.NC}`);
      }
    },
  };
}
