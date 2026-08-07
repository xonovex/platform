import {chmodSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {delimiter, join} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {resolveExecutable} from "../../../src/executable.js";

describe("resolveExecutable", () => {
  const directories: string[] = [];

  afterEach(() => {
    for (const directory of directories) {
      rmSync(directory, {recursive: true, force: true});
    }
    directories.length = 0;
  });

  it("returns the absolute executable from the first matching PATH entry", () => {
    const first = mkdtempSync(join(tmpdir(), "executable-first-"));
    const second = mkdtempSync(join(tmpdir(), "executable-second-"));
    directories.push(first, second);
    const executable = join(second, "example-command");
    writeFileSync(executable, "#!/bin/sh\n");
    chmodSync(executable, 0o755);

    const result = resolveExecutable(
      "example-command",
      [first, second].join(delimiter),
    );

    expect(result).toBe(executable);
  });

  it("rejects command paths", () => {
    expect(() => resolveExecutable("./example-command", "/bin")).toThrow(
      "must be a bare command",
    );
  });

  it("rejects a command missing from PATH", () => {
    const directory = mkdtempSync(join(tmpdir(), "executable-missing-"));
    directories.push(directory);

    expect(() => resolveExecutable("missing-command", directory)).toThrow(
      "was not found on PATH",
    );
  });

  it("does not search relative PATH entries", () => {
    expect(() => resolveExecutable("example-command", ".")).toThrow(
      "was not found on PATH",
    );
  });
});
