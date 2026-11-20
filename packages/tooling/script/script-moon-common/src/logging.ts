// These log lines restate the ones in shared-core rather than importing them.
// shared-core is tagged npm, so its npm-check, publish, and version tasks depend
// on the script packages; a script package that depended on shared-core would
// close that loop and moon would reject the project graph with would_cycle.
const C = {
  RED: "\u001B[0;31m",
  GREEN: "\u001B[0;32m",
  YELLOW: "\u001B[1;33m",
  BLUE: "\u001B[0;34m",
  NC: "\u001B[0m",
} as const;

/** A log line for the Output port: a colored level tag, then the message. */
export const infoLine = (message: string): string =>
  `${C.BLUE}[INFO]${C.NC} ${message}`;

export const successLine = (message: string): string =>
  `${C.GREEN}[SUCCESS]${C.NC} ${message}`;

export const warningLine = (message: string): string =>
  `${C.YELLOW}[WARNING]${C.NC} ${message}`;

export const errorLine = (message: string): string =>
  `${C.RED}[ERROR]${C.NC} ${message}`;
