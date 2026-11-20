import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {afterEach, describe, expect, it, vi} from "vitest";
import {
  main,
  type VersionDetectDependencies,
} from "../../../src/version-detect.js";

afterEach(() => {
  vi.restoreAllMocks();
});

const ROOT = "/repo";

const dependencies = (resolvedRefs: string[]): VersionDetectDependencies => ({
  openRepository: () => ({
    resolveCommit: (ref) => {
      resolvedRefs.push(ref);
      return "commit";
    },
    readFile: () => JSON.stringify({name: "example", version: "1.0.0"}),
  }),
  listProjects: () => [{id: "example", source: "packages/example"}],
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
    const log = vi.spyOn(console, "log").mockImplementation(() => {});
    const refs: string[] = [];

    expect(main([], `${ROOT}/packages`, dependencies(refs))).toBe(0);
    expect(refs).toEqual(["HEAD~1"]);
    expect(log).toHaveBeenCalledWith(JSON.stringify(["example"]));
  });

  it("compares against the ref given as an option or a positional", () => {
    vi.spyOn(console, "log").mockImplementation(() => {});
    const refs: string[] = [];

    main(["--ref", "main"], ROOT, dependencies(refs));
    main(["v1.0.0"], ROOT, dependencies(refs));

    expect(refs).toEqual(["main", "v1.0.0"]);
  });

  it("prints the help text and reads nothing on --help", () => {
    const log = vi.spyOn(console, "log").mockImplementation(() => {});
    const refs: string[] = [];

    expect(main(["--help"], ROOT, dependencies(refs))).toBe(0);
    expect(log).toHaveBeenCalledWith(
      expect.stringContaining("moon-version-detect"),
    );
    expect(refs).toEqual([]);
  });
});
