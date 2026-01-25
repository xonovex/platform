/**
 * ANSI color codes for terminal output
 */

export const Colors = {
  RED: "\u001B[0;31m",
  GREEN: "\u001B[0;32m",
  YELLOW: "\u001B[1;33m",
  BLUE: "\u001B[0;34m",
  CYAN: "\u001B[0;36m",
  PURPLE: "\u001B[0;35m",
  NC: "\u001B[0m", // No Color
} as const;

export type ColorName = keyof typeof Colors;
