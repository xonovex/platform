const C = {
  RED: "\u001B[0;31m",
  GREEN: "\u001B[0;32m",
  YELLOW: "\u001B[1;33m",
  BLUE: "\u001B[0;34m",
  NC: "\u001B[0m",
} as const;

export const logInfo = (...args: unknown[]): void => {
  console.log(`${C.BLUE}[INFO]${C.NC}`, ...args);
};

export const logSuccess = (...args: unknown[]): void => {
  console.log(`${C.GREEN}[SUCCESS]${C.NC}`, ...args);
};

export const logWarning = (...args: unknown[]): void => {
  console.log(`${C.YELLOW}[WARNING]${C.NC}`, ...args);
};

export const logError = (...args: unknown[]): void => {
  console.error(`${C.RED}[ERROR]${C.NC}`, ...args);
};
