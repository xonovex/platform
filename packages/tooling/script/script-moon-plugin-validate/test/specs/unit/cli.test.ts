import {join} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {describe, expect, it} from "vitest";
import {main} from "../../../src/cli.js";

const ROOT = "/repo";
const PACKAGE_DIR = join(ROOT, "packages", "runtime", "plugin", "plugin-test");

const pluginPackage = (readme: string): FileSystem =>
  memoryFileSystem({
    files: {
      [join(PACKAGE_DIR, "package.json")]: "{}",
      [join(PACKAGE_DIR, ".claude-plugin", "plugin.json")]: JSON.stringify({
        name: "xonovex-test",
      }),
      [join(PACKAGE_DIR, "README.md")]: readme,
      [join(PACKAGE_DIR, "skills", "using-test", "SKILL.md")]: "# Test\n",
    },
  });

const LISTED =
  "# Test\n\n## Skills\n\n| Skill | Use |\n| --- | --- |\n| [using-test](skills/using-test/SKILL.md) | Tests |\n";

// Runs the validator from the plugin root and returns the exit code with its output.
const run = (argv: readonly string[], fs: FileSystem = memoryFileSystem()) => {
  const output = memoryOutput();
  const code = main(argv, join(ROOT, "packages", "runtime", "plugin"), {
    fs,
    output,
  });
  return {code, stdout: output.stdoutLines(), stderr: output.stderrLines()};
};

describe("plugin validator CLI", () => {
  it("renders help", () => {
    const result = run(["--help"]);

    expect(result.code).toBe(0);
    expect(result.stdout[0]).toContain("usage:");
  });

  it.each([
    [
      ["--unknown"],
      "plugin validation input error: Unknown option '--unknown'",
    ],
    [
      ["one", "two"],
      "plugin validation input error: expected at most one package directory",
    ],
  ])("rejects the input %j", (argv, message) => {
    const result = run(argv);

    expect(result.code).toBe(2);
    expect(result.stderr[0]).toContain(message);
  });

  it("passes a package resolved against the working directory", () => {
    expect(run(["plugin-test"], pluginPackage(LISTED))).toEqual({
      code: 0,
      stdout: ["Result: PASS (1 skill(s), 0 warning(s))"],
      stderr: [],
    });
  });

  it("fails a package and names each issue on stderr", () => {
    expect(
      run(
        ["plugin-test", "--repo-root", "../../.."],
        pluginPackage("# Test\n"),
      ),
    ).toEqual({
      code: 1,
      stdout: [],
      stderr: [
        "[FAIL] packages/runtime/plugin/plugin-test/README.md: readme.skill-unlisted: skill 'using-test' is not listed in the README skill tables",
        "Result: FAIL (1 error(s), 0 warning(s), 1 skill(s))",
      ],
    });
  });

  it("reports a failure the validator throws", () => {
    const fs = pluginPackage(LISTED);
    const failing: FileSystem = {
      ...fs,
      readDirectory: () => {
        throw new Error("disk gone");
      },
    };

    expect(run(["plugin-test"], failing)).toMatchObject({
      code: 2,
      stderr: ["plugin validation failed unexpectedly: disk gone"],
    });
  });
});
