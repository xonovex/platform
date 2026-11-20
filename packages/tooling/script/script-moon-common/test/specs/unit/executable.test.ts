import {delimiter, join} from "node:path";
import {describe, expect, it} from "vitest";
import {memoryFileSystem} from "../../../src/adapters/file-system-memory.js";
import {
  findExecutable,
  type ExecutableEnvironment,
} from "../../../src/executable.js";

const FIRST = "/path/first";
const SECOND = "/path/second";
const SEARCH_PATH = [FIRST, SECOND].join(delimiter);

const posix = (searchPath: string): ExecutableEnvironment => ({
  searchPath,
  extensions: undefined,
  platform: "linux",
});

describe("findExecutable", () => {
  it("returns the absolute executable from the first matching PATH entry", () => {
    const executable = join(SECOND, "example-command");
    const fs = memoryFileSystem({
      files: {[executable]: "#!/bin/sh\n"},
      executables: [executable],
    });

    expect(findExecutable("example-command", posix(SEARCH_PATH), fs)).toBe(
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

    expect(findExecutable("example-command", posix(SEARCH_PATH), fs)).toBe(
      first,
    );
  });

  it("skips a match that is present but not executable", () => {
    const first = join(FIRST, "example-command");
    const second = join(SECOND, "example-command");
    const fs = memoryFileSystem({
      files: {[first]: "#!/bin/sh\n", [second]: "#!/bin/sh\n"},
      executables: [second],
    });

    expect(findExecutable("example-command", posix(SEARCH_PATH), fs)).toBe(
      second,
    );
  });

  it("skips an executable directory that shares the command name", () => {
    const fs = memoryFileSystem({
      directories: [join(FIRST, "example-command")],
      executables: [join(FIRST, "example-command")],
    });

    expect(() => findExecutable("example-command", posix(FIRST), fs)).toThrow(
      "was not found on PATH",
    );
  });

  it("rejects command paths", () => {
    expect(() =>
      findExecutable("./example-command", posix("/bin"), memoryFileSystem()),
    ).toThrow("must be a bare command");
  });

  it("rejects an empty command name", () => {
    expect(() => findExecutable("", posix("/bin"), memoryFileSystem())).toThrow(
      "must be a bare command",
    );
  });

  it("rejects an empty PATH", () => {
    expect(() =>
      findExecutable("example-command", posix(""), memoryFileSystem()),
    ).toThrow("PATH is empty");
  });

  it("rejects a command missing from PATH", () => {
    expect(() =>
      findExecutable("missing-command", posix(FIRST), memoryFileSystem()),
    ).toThrow("was not found on PATH");
  });

  it("does not search relative PATH entries", () => {
    const fs = memoryFileSystem({
      files: {"/example-command": "#!/bin/sh\n"},
      executables: ["/example-command"],
    });

    expect(() => findExecutable("example-command", posix("."), fs)).toThrow(
      "was not found on PATH",
    );
  });
});

const windows = (extensions: string | undefined): ExecutableEnvironment => ({
  searchPath: FIRST,
  extensions,
  platform: "win32",
});

describe("findExecutable on Windows", () => {
  it("tries each PATHEXT extension in order", () => {
    const executable = join(FIRST, "example-command.CMD");
    const fs = memoryFileSystem({
      files: {[executable]: "@echo off\r\n"},
      executables: [executable],
    });

    expect(findExecutable("example-command", windows(".EXE;.CMD"), fs)).toBe(
      executable,
    );
  });

  it("falls back to the default extensions without PATHEXT", () => {
    const executable = join(FIRST, "example-command.BAT");
    const fs = memoryFileSystem({
      files: {[executable]: "@echo off\r\n"},
      executables: [executable],
    });

    expect(findExecutable("example-command", windows(undefined), fs)).toBe(
      executable,
    );
  });

  it("rejects a command name holding a forward slash", () => {
    expect(() =>
      findExecutable("bin/example", windows(undefined), memoryFileSystem()),
    ).toThrow("must be a bare command");
  });
});
