import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {describe, expect, it} from "vitest";
import {loadPlugins} from "../../../src/catalog.js";
import {
  repositoryFileSystem,
  ROOT,
  type PluginFixture,
} from "../../support/repository.js";

const claudeOnly: PluginFixture = {
  name: "xonovex-core",
  directory: "packages/runtime/plugin/plugin-core",
  version: "5.3.0",
  claude: true,
  codex: false,
};

const claudeAndCodex: PluginFixture = {
  name: "xonovex-workflow",
  directory: "packages/runtime/plugin/plugin-workflow",
  version: "5.3.0",
  dependencies: ["xonovex-core", "xonovex-agentic"],
  claude: true,
  codex: true,
};

describe("loadPlugins", () => {
  it("reads a Claude Code only plugin", () => {
    const fs = repositoryFileSystem([claudeOnly]);
    expect(loadPlugins(ROOT, fs)).toEqual({
      marketplace: "xonovex-marketplace",
      plugins: [
        {
          directory: "packages/runtime/plugin/plugin-core",
          name: "xonovex-core",
          version: "5.3.0",
          dependencies: [],
          companions: [],
          harnesses: ["Claude Code"],
        },
      ],
    });
  });

  it("reads a plugin for both harnesses with sorted dependencies", () => {
    const fs = repositoryFileSystem([claudeAndCodex, claudeOnly]);
    const catalog = loadPlugins(ROOT, fs);
    expect(catalog.plugins.map((plugin) => plugin.name)).toEqual([
      "xonovex-core",
      "xonovex-workflow",
    ]);
    expect(catalog.plugins[1]).toEqual({
      directory: "packages/runtime/plugin/plugin-workflow",
      name: "xonovex-workflow",
      version: "5.3.0",
      dependencies: ["xonovex-agentic", "xonovex-core"],
      companions: [],
      harnesses: ["Claude Code", "Codex"],
    });
  });

  it("omits Codex when the Codex marketplace does not list the plugin", () => {
    const fs = repositoryFileSystem([{...claudeAndCodex, codex: false}]);
    expect(loadPlugins(ROOT, fs).plugins[0]?.harnesses).toEqual([
      "Claude Code",
    ]);
  });

  it("rejects differing marketplace names", () => {
    const fs = repositoryFileSystem([claudeOnly], {
      codexMarketplace: "other-marketplace",
    });
    expect(() => loadPlugins(ROOT, fs)).toThrow(
      "Harness marketplace names differ.",
    );
  });

  it("rejects a plugin whose manifest name differs from its marketplace entry", () => {
    const fs = memoryFileSystem({
      files: {
        [ROOT + "/.claude-plugin/marketplace.json"]: JSON.stringify({
          name: "xonovex-marketplace",
          plugins: [{name: "xonovex-core", source: "./plugins/core"}],
        }),
        [ROOT + "/.agents/plugins/marketplace.json"]: JSON.stringify({
          name: "xonovex-marketplace",
          plugins: [],
        }),
        [ROOT + "/plugins/core/plugin.json"]: JSON.stringify({
          name: "xonovex-renamed",
          version: "1.0.0",
        }),
      },
    });
    expect(() => loadPlugins(ROOT, fs)).toThrow(
      "Marketplace and plugin names differ: xonovex-core",
    );
  });

  it("rejects a plugin source that escapes the repository", () => {
    const fs = memoryFileSystem({
      files: {
        [ROOT + "/.claude-plugin/marketplace.json"]: JSON.stringify({
          name: "xonovex-marketplace",
          plugins: [{name: "xonovex-core", source: "../outside"}],
        }),
        [ROOT + "/.agents/plugins/marketplace.json"]: JSON.stringify({
          name: "xonovex-marketplace",
          plugins: [],
        }),
      },
    });
    expect(() => loadPlugins(ROOT, fs)).toThrow("Path escapes");
  });

  it("names each undeclared plugin that a skill hands off to", () => {
    const native: PluginFixture = {
      name: "xonovex-native",
      directory: "packages/runtime/plugin/plugin-native",
      version: "5.3.0",
      claude: true,
      codex: false,
      skills: {"skills/managing-memory/SKILL.md": "# Memory\n"},
    };
    const core: PluginFixture = {
      ...claudeOnly,
      skills: {
        "skills/using-git/SKILL.md": "# Git\n",
        "skills/reviewing-code/SKILL.md": "# Review\n",
      },
    };
    const engine: PluginFixture = {
      name: "xonovex-game-engine",
      directory: "packages/runtime/plugin/plugin-game-engine",
      version: "5.3.0",
      dependencies: ["xonovex-native"],
      claude: true,
      codex: false,
      skills: {
        "skills/mixing-audio/SKILL.md":
          "# Audio\n\nSee **managing-memory** and **mixing-audio**.\n",
        "skills/mixing-audio/references/a.md":
          "# A\n\nCommit with **using-git**, not using-docker or **using-unknown-tool**.\n",
      },
    };
    const catalog = loadPlugins(
      ROOT,
      repositoryFileSystem([native, core, engine]),
    );
    expect(
      catalog.plugins.map(({name, companions}) => [name, companions]),
    ).toEqual([
      ["xonovex-core", []],
      ["xonovex-game-engine", ["xonovex-core"]],
      ["xonovex-native", []],
    ]);
  });
});
