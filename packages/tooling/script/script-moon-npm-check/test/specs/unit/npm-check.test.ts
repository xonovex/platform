import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {afterEach, beforeEach, describe, expect, it, vi} from "vitest";
import {main, type NpmCheckDependencies} from "../../../src/npm-check.js";

const ROOT = "/repo/packages/example";

const PUBLISHABLE = {
  name: "@xonovex/example",
  version: "1.0.0",
  description: "An example package",
  license: "MIT",
  repository: {type: "git", url: "https://example.com/repo.git"},
  type: "module",
  files: ["dist"],
  exports: {".": "./dist/index.js"},
  publishConfig: {access: "public"},
};

const dependencies = (
  manifest: unknown,
  packed: readonly string[] = ["package.json", "dist/index.js"],
): NpmCheckDependencies => ({
  fs: memoryFileSystem({
    files: {
      [`${ROOT}/package.json`]: JSON.stringify(manifest),
      [`${ROOT}/dist/index.js`]: 'import "./missing.js";\n',
    },
  }),
  listPackedFiles: () => packed,
});

// Everything the run printed, on either stream, one call per line.
let printed: string[];

const record = (...parts: readonly unknown[]): void => {
  printed.push(parts.map(String).join(" "));
};

beforeEach(() => {
  printed = [];
  vi.spyOn(console, "error").mockImplementation(record);
  vi.spyOn(console, "log").mockImplementation(record);
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe("main", () => {
  it("prints the help text on --help", () => {
    expect(main(["--help"], ROOT, dependencies(PUBLISHABLE))).toBe(0);
    expect(printed.join("\n")).toContain("moon-npm-check");
  });

  it("skips a private package without packing it", () => {
    const listPackedFiles = vi.fn(() => []);

    expect(
      main([], ROOT, {
        ...dependencies({...PUBLISHABLE, private: true}),
        listPackedFiles,
      }),
    ).toBe(0);
    expect(listPackedFiles).not.toHaveBeenCalled();
  });

  it("fails on a manifest it cannot read", () => {
    expect(
      main([], ROOT, {
        fs: memoryFileSystem(),
        listPackedFiles: () => [],
      }),
    ).toBe(1);
    expect(printed.join("\n")).toContain("Failed to parse package.json");
  });

  it("fails when the packed files cannot be listed", () => {
    expect(
      main([], ROOT, {
        ...dependencies(PUBLISHABLE),
        listPackedFiles: () => {
          throw new Error("npm pack failed: offline");
        },
      }),
    ).toBe(1);
    expect(printed.join("\n")).toContain(
      "Failed to inspect packed package: Error: npm pack failed: offline",
    );
  });

  it("lists every problem of a package that is not ready", () => {
    expect(
      main(
        [],
        ROOT,
        dependencies(PUBLISHABLE, ["package.json", "dist/index.js"]),
      ),
    ).toBe(1);
    const output = printed.join("\n");
    expect(output).toContain("@xonovex/example is not ready for publishing");
    expect(output).toContain(
      'Packed file "dist/index.js" imports missing relative target: "./missing.js"',
    );
  });

  it("names the manifest path of a package without a name", () => {
    const {name: _name, ...unnamed} = PUBLISHABLE;

    expect(main([], ROOT, dependencies(unnamed))).toBe(1);
    expect(printed.join("\n")).toContain(
      `${ROOT}/package.json is not ready for publishing`,
    );
  });

  it("reports a package that is ready", () => {
    const ready = dependencies(PUBLISHABLE);
    ready.fs.writeFile(`${ROOT}/dist/index.js`, "export {};\n");

    expect(main([], ROOT, ready)).toBe(0);
    expect(printed.join("\n")).toContain(
      "@xonovex/example@1.0.0 is ready for publishing",
    );
  });
});
