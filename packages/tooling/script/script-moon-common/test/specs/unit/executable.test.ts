import {delimiter, join} from "node:path";
import {describe, expect, it} from "vitest";
import {resolveExecutable} from "../../../src/executable.js";
import {memoryFileSystem} from "../../../src/file-system-memory.js";

const FIRST = "/path/first";
const SECOND = "/path/second";
const SEARCH_PATH = [FIRST, SECOND].join(delimiter);

describe("resolveExecutable", () => {
  it("returns the absolute executable from the first matching PATH entry", () => {
    const executable = join(SECOND, "example-command");
    const fs = memoryFileSystem({
      files: {[executable]: "#!/bin/sh\n"},
      executables: [executable],
    });

    expect(resolveExecutable("example-command", SEARCH_PATH, fs)).toBe(
      executable,
    );
  });

  it("prefers the earlier PATH entry when both hold the command", () => {
    const first = join(FIRST, "example-command");
    const second = join(SECOND, "example-command");
    const fs = memoryFileSystem({
      files: {[first]: "#!/bin/sh\n", [second]: "#!/bin/sh\n"},
      executables: [first, second],
    });

    expect(resolveExecutable("example-command", SEARCH_PATH, fs)).toBe(first);
  });

  it("skips a match that is present but not executable", () => {
    const first = join(FIRST, "example-command");
    const second = join(SECOND, "example-command");
    const fs = memoryFileSystem({
      files: {[first]: "#!/bin/sh\n", [second]: "#!/bin/sh\n"},
      executables: [second],
    });

    expect(resolveExecutable("example-command", SEARCH_PATH, fs)).toBe(second);
  });

  it("skips an executable directory that shares the command name", () => {
    const fs = memoryFileSystem({
      directories: [join(FIRST, "example-command")],
      executables: [join(FIRST, "example-command")],
    });

    expect(() => resolveExecutable("example-command", FIRST, fs)).toThrow(
      "was not found on PATH",
    );
  });

  it("rejects command paths", () => {
    expect(() =>
      resolveExecutable("./example-command", "/bin", memoryFileSystem()),
    ).toThrow("must be a bare command");
  });

  it("rejects an empty command name", () => {
    expect(() => resolveExecutable("", "/bin", memoryFileSystem())).toThrow(
      "must be a bare command",
    );
  });

  it("rejects an empty PATH", () => {
    expect(() =>
      resolveExecutable("example-command", "", memoryFileSystem()),
    ).toThrow("PATH is empty");
  });

  it("rejects a command missing from PATH", () => {
    expect(() =>
      resolveExecutable("missing-command", FIRST, memoryFileSystem()),
    ).toThrow("was not found on PATH");
  });

  it("does not search relative PATH entries", () => {
    const fs = memoryFileSystem({
      files: {"/example-command": "#!/bin/sh\n"},
      executables: ["/example-command"],
    });

    expect(() => resolveExecutable("example-command", ".", fs)).toThrow(
      "was not found on PATH",
    );
  });
});
