import {describe, expect, it} from "vitest";
import {
  filesystemImportFailures,
  staleAllowlistFailures,
  type SourceFile,
} from "../../../src/filesystem-imports.js";

// One entry that is really on the allowlist, so a case can assert the allowance
// without restating the list here.
const ALLOWED_PATH = "packages/script/script-moon-common/src/file-system.ts";
const OTHER_PATH = "packages/script/script-moon-example/src/reader.ts";

const file = (path: string, text: string): SourceFile => ({path, text});

describe("filesystemImportFailures", () => {
  it("reports a module that imports node:fs without an allowance", () => {
    const failures = filesystemImportFailures([
      file(OTHER_PATH, 'import {readFileSync} from "node:fs";\n'),
    ]);

    expect(failures).toHaveLength(1);
    expect(failures[0]).toContain(OTHER_PATH);
    expect(failures[0]).toContain("FileSystem port");
  });

  it("allows a module named on the allowlist", () => {
    expect(
      filesystemImportFailures([
        file(ALLOWED_PATH, 'import {statSync} from "node:fs";\n'),
      ]),
    ).toEqual([]);
  });

  it("reports node:fs/promises and a re-export as well as a plain import", () => {
    for (const text of [
      'import {readFile} from "node:fs/promises";\n',
      'export {readFileSync} from "node:fs";\n',
      'import type {Stats} from "node:fs";\n',
    ]) {
      expect(filesystemImportFailures([file(OTHER_PATH, text)])).toHaveLength(
        1,
      );
    }
  });

  it("ignores a module that only mentions node:fs in prose", () => {
    expect(
      filesystemImportFailures([
        file(OTHER_PATH, "// Reaching node:fs here would bypass the port.\n"),
      ]),
    ).toEqual([]);
  });

  it("ignores a module that imports something else entirely", () => {
    expect(
      filesystemImportFailures([
        file(OTHER_PATH, 'import {join} from "node:path";\n'),
      ]),
    ).toEqual([]);
  });

  it("names every offending module, sorted", () => {
    const failures = filesystemImportFailures([
      file("packages/b/src/b.ts", 'import {readFileSync} from "node:fs";'),
      file("packages/a/src/a.ts", 'import {readFileSync} from "node:fs";'),
    ]);

    expect(failures.map((failure) => failure.split(" ", 1)[0])).toEqual([
      "packages/a/src/a.ts",
      "packages/b/src/b.ts",
    ]);
  });
});

describe("staleAllowlistFailures", () => {
  it("reports an allowance whose module no longer imports node:fs", () => {
    const failures = staleAllowlistFailures([
      file(ALLOWED_PATH, 'import {join} from "node:path";\n'),
    ]);

    expect(failures.some((failure) => failure.startsWith(ALLOWED_PATH))).toBe(
      true,
    );
    expect(failures[0]).toContain("no longer does");
  });

  it("reports nothing when an allowed module still imports node:fs", () => {
    expect(
      staleAllowlistFailures([
        file(ALLOWED_PATH, 'import {statSync} from "node:fs";'),
      ]),
    ).toEqual([]);
  });

  it("ignores a module that imports node:fs without an allowance", () => {
    expect(
      staleAllowlistFailures([
        file(OTHER_PATH, 'import {readFileSync} from "node:fs";'),
      ]),
    ).toEqual([]);
  });

  it("leaves an allowance alone when the scan does not reach its module", () => {
    // A fixture repository ships none of the allowed modules, and their
    // allowances say nothing about it.
    expect(staleAllowlistFailures([file(OTHER_PATH, "")])).toEqual([]);
  });
});
