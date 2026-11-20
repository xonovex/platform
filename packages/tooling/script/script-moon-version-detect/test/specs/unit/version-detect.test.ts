import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import {describe, expect, it} from "vitest";
import {
  main,
  type VersionDetectDependencies,
} from "../../../src/version-detect.js";

const ROOT = "/repo";

const dependencies = (
  resolvedRefs: string[],
  output = memoryOutput(),
): VersionDetectDependencies => ({
  openRepository: () => ({
    resolveCommit: (ref) => {
      resolvedRefs.push(ref);
      return "commit";
    },
    readFile: () => JSON.stringify({name: "example", version: "1.0.0"}),
  }),
  listProjects: () => [{id: "example", source: "packages/example"}],
  output,
  fs: memoryFileSystem({
    directories: [`${ROOT}/.moon`],
    files: {
      [`${ROOT}/packages/example/package.json`]: JSON.stringify({
        name: "example",
        version: "1.1.0",
      }),
    },
  }),
});

describe("main", () => {
  it("prints the changed projects since HEAD~1 by default", () => {
    const output = memoryOutput();
    const refs: string[] = [];

    expect(main([], `${ROOT}/packages`, dependencies(refs, output))).toBe(0);
    expect(refs).toEqual(["HEAD~1"]);
    expect(output.stdoutLines()).toEqual([JSON.stringify(["example"])]);
  });

  it("compares against the ref given as an option or a positional", () => {
    const refs: string[] = [];

    main(["--ref", "main"], ROOT, dependencies(refs));
    main(["v1.0.0"], ROOT, dependencies(refs));

    expect(refs).toEqual(["main", "v1.0.0"]);
  });

  it("prints the help text and reads nothing on --help", () => {
    const output = memoryOutput();
    const refs: string[] = [];

    expect(main(["--help"], ROOT, dependencies(refs, output))).toBe(0);
    expect(output.stdoutLines()[0]).toContain("moon-version-detect");
    expect(refs).toEqual([]);
  });
});
