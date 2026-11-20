import {afterEach, describe, expect, it, vi} from "vitest";
import {runCli} from "../../../src/cli.js";
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

afterEach(() => {
  vi.restoreAllMocks();
});

describe("runCli", () => {
  it("prints the help text and writes nothing on --help", () => {
    const log = vi.spyOn(console, "log").mockImplementation(() => {});
    const fs = repositoryFileSystem([core]);
    const before = fs.readText(readmePath);

    expect(runCli(["--help"], ROOT, fs)).toBe(0);
    expect(log).toHaveBeenCalledWith(
      expect.stringContaining("moon-plugin-sync-installation"),
    );
    expect(fs.readText(readmePath)).toBe(before);
  });

  it("exits 1 and names the stale README under --check without writing", () => {
    vi.spyOn(console, "log").mockImplementation(() => {});
    const errors = vi.spyOn(console, "error").mockImplementation(() => {});
    const fs = repositoryFileSystem([core]);
    const before = fs.readText(readmePath);

    expect(runCli(["--check"], ROOT, fs)).toBe(1);
    expect(fs.readText(readmePath)).toBe(before);
    const printed = errors.mock.calls.map((call) => call.join(" ")).join("\n");
    expect(printed).toContain(directory + "/README.md");
    expect(printed).toContain(
      "npx moon run script-moon-plugin-sync-installation:plugin-installation-sync",
    );
  });

  it("writes the stale block without --check and then passes --check", () => {
    vi.spyOn(console, "log").mockImplementation(() => {});
    vi.spyOn(console, "error").mockImplementation(() => {});
    const fs = repositoryFileSystem([core]);

    expect(runCli([], ROOT, fs)).toBe(0);
    expect(fs.readText(readmePath)).toContain("## Install");
    expect(runCli(["--check"], ROOT, fs)).toBe(0);
  });
});
