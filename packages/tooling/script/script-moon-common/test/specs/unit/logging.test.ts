import {describe, expect, it} from "vitest";
import {
  errorLine,
  infoLine,
  successLine,
  warningLine,
} from "../../../src/logging.js";

describe("logging", () => {
  it.each([
    [infoLine, "\u001B[0;34m[INFO]\u001B[0m"],
    [successLine, "\u001B[0;32m[SUCCESS]\u001B[0m"],
    [warningLine, "\u001B[1;33m[WARNING]\u001B[0m"],
    [errorLine, "\u001B[0;31m[ERROR]\u001B[0m"],
  ] as const)("prefixes the message with a colored level tag", (line, tag) => {
    expect(line("100% done")).toBe(`${tag} 100% done`);
  });
});
