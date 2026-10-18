import {describe, expect, it} from "vitest";
import {
  installationEnd,
  installationStart,
} from "../../../src/installation-section.js";
import {
  checkInstallations,
  syncInstallations,
} from "../../../src/readme-sync.js";
import {
  repositoryFileSystem,
  ROOT,
  type PluginFixture,
} from "../../support/repository.js";

const directory = "packages/runtime/plugin/plugin-workflow";
const readmePath = ROOT + "/" + directory + "/README.md";

const workflow: PluginFixture = {
  name: "xonovex-workflow",
  directory,
  version: "5.3.0",
  dependencies: ["xonovex-core"],
  claude: true,
  codex: true,
};

const staleReadme = [
  "# Workflow",
  "",
  "Authored introduction.",
  "",
  installationStart,
  "",
  "## Install",
  "",
  "Stale instructions for version 5.2.0.",
  "",
  installationEnd,
  "",
  "## Usage",
  "",
  "Keep this procedure.",
  "",
].join("\n");

describe("checkInstallations", () => {
  it("reports a README whose block is stale and writes nothing", () => {
    const fs = repositoryFileSystem([{...workflow, readme: staleReadme}]);
    expect(checkInstallations(ROOT, fs)).toEqual([directory + "/README.md"]);
    expect(fs.readText(readmePath)).toBe(staleReadme);
  });

  it("reports nothing when every block matches the catalog", () => {
    const fs = repositoryFileSystem([{...workflow, readme: staleReadme}]);
    syncInstallations(ROOT, fs);
    expect(checkInstallations(ROOT, fs)).toEqual([]);
  });
});

describe("syncInstallations", () => {
  it("writes only the marked block and keeps the prose around it", () => {
    const fs = repositoryFileSystem([{...workflow, readme: staleReadme}]);
    expect(syncInstallations(ROOT, fs)).toEqual([directory + "/README.md"]);
    const updated = fs.readText(readmePath);
    expect(updated).not.toContain("Stale instructions");
    expect(updated).toContain("# Workflow\n\nAuthored introduction.\n\n");
    expect(updated).toContain("\n\n## Usage\n\nKeep this procedure.\n");
    expect(updated).toContain(
      "The bundle version is `5.3.0`.\n\n### Claude Code",
    );
    expect(updated).toContain("Plugin dependencies: `xonovex-core`.");
  });

  it("writes nothing for a README already in sync", () => {
    const fs = repositoryFileSystem([{...workflow, readme: staleReadme}]);
    syncInstallations(ROOT, fs);
    const synced = fs.readText(readmePath);
    expect(syncInstallations(ROOT, fs)).toEqual([]);
    expect(fs.readText(readmePath)).toBe(synced);
  });

  it("names only the READMEs that drifted", () => {
    const core: PluginFixture = {
      name: "xonovex-core",
      directory: "packages/runtime/plugin/plugin-core",
      version: "5.3.0",
      claude: true,
      codex: false,
      readme: "# Core\n\nIntro.\n",
    };
    const fs = repositoryFileSystem([core, {...workflow, readme: staleReadme}]);
    syncInstallations(ROOT, fs);
    fs.writeFile(readmePath, staleReadme);
    expect(checkInstallations(ROOT, fs)).toEqual([directory + "/README.md"]);
  });
});
