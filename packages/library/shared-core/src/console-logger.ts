/**
 * The edge that wires a logger to a console and the process environment.
 */

import {createLogger, type Logger} from "./logging.js";

export interface ConsoleLike {
  readonly log: (...args: unknown[]) => void;
  readonly error: (...args: unknown[]) => void;
}

/**
 * Build a logger that writes to a console and enables debug output when the
 * `DEBUG` entry of the given environment is non-empty.
 */
export function createConsoleLogger(
  console: ConsoleLike,
  environment: Readonly<Record<string, string | undefined>>,
): Logger {
  return createLogger({
    sink: {
      out: (...args) => {
        console.log(...args);
      },
      err: (...args) => {
        console.error(...args);
      },
    },
    debug: Boolean(environment.DEBUG),
  });
}
