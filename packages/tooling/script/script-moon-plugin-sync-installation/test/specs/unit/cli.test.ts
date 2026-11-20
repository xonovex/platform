import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {describe, expect, it} from "vitest";
import {main} from "../../../src/cli.js";
import {
  installationEnd,
  installationStart,
} from "../../../src/installation-section.js";
import {
  repositoryFileSystem,
  ROOT,
  type PluginFixture,
} from "../../support/repository.js";

const directory = "packages/runtime/plugin/plugin-core";
const readmePath = ROOT + "/" + directory + "/README.md";

const core: PluginFixture = {
  name: "xonovex-core",
  directory,
  version: "5.3.0",
  claude: true,
  codex: false,
  readme:
    "# Core\n\nIntro.\n\n" +
    installationStart +
    "\n\nOld.\n\n" +
    installationEnd +
    "\n",
};

// Runs the command on the repository and returns the exit code with its output.
const run = (argv: readonly string[], fs: FileSystem) => {
  const output = memoryOutput();
  const code = main(argv, ROOT, {fs, output});
  return {
    code,
    stdout: output.stdoutLines().join("\n"),
    stderr: output.stderrLines().join("\n"),
  };
};

describe("main", () => {
  it("prints the help text and writes nothing on --help", () => {
    const fs = repositoryFileSystem([core]);
    const before = fs.readText(readmePath);
    const result = run(["--help"], fs);

    expect(result.code).toBe(0);
    expect(result.stdout).toContain("moon-plugin-sync-installation");
    expect(fs.readText(readmePath)).toBe(before);
  });

  it("exits 1 and names the stale README under --check without writing", () => {
    const fs = repositoryFileSystem([core]);
    const before = fs.readText(readmePath);
    const result = run(["--check"], fs);

    expect(result.code).toBe(1);
    expect(fs.readText(readmePath)).toBe(before);
    expect(result.stderr).toContain(directory + "/README.md");
    expect(result.stderr).toContain(
      "npx moon run script-moon-plugin-sync-installation:plugin-installation-sync",
    );
  });

  it("writes the stale block without --check and then passes --check", () => {
    const fs = repositoryFileSystem([core]);

    expect(run([], fs)).toMatchObject({
      code: 0,
      stdout: expect.stringContaining(
        `Updated ${directory}/README.md`,
      ) as unknown,
    });
    expect(fs.readText(readmePath)).toContain("## Install");
    expect(run(["--check"], fs)).toMatchObject({
      code: 0,
      stdout: expect.stringContaining(
        "Plugin installation sections are in sync.",
      ) as unknown,
    });
  });
});
