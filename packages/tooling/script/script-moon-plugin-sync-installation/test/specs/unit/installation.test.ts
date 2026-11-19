import {describe, expect, it} from "vitest";
import {installation} from "../../../src/installation.js";

describe("installation", () => {
  it("renders one harness section without dependencies", () => {
    const text = installation(
      {
        directory: "packages/runtime/plugin/plugin-core",
        name: "xonovex-core",
        version: "5.3.0",
        dependencies: [],
        companions: [],
        harnesses: ["Claude Code"],
      },
      "xonovex-marketplace",
    );
    expect(text).toBe(
      [
        "## Install",
        "",
        "Add the Xonovex marketplace once, then install `xonovex-core` in the selected harness. The bundle version is `5.3.0`.",
        "",
        "### Claude Code",
        "",
        "Install this bundle in Claude Code. The harness discovers its bundled skills.",
        "",
        "```bash",
        "claude plugin marketplace add xonovex/platform",
        "claude plugin install xonovex-core@xonovex-marketplace",
        "```",
      ].join("\n"),
    );
  });

  it("renders both harness sections, the dependencies and the companions", () => {
    const text = installation(
      {
        directory: "packages/runtime/plugin/plugin-workflow",
        name: "xonovex-workflow",
        version: "5.3.0",
        dependencies: ["xonovex-core", "xonovex-agentic"],
        companions: ["xonovex-native", "xonovex-platform"],
        harnesses: ["Claude Code", "Codex"],
      },
      "xonovex-marketplace",
    );
    expect(text).toBe(
      [
        "## Install",
        "",
        "Add the Xonovex marketplace once, then install `xonovex-workflow` in the selected harness. The bundle version is `5.3.0`.",
        "",
        "### Claude Code",
        "",
        "Install this bundle in Claude Code. The harness discovers its bundled skills.",
        "",
        "```bash",
        "claude plugin marketplace add xonovex/platform",
        "claude plugin install xonovex-workflow@xonovex-marketplace",
        "```",
        "",
        "### Codex",
        "",
        "Install this bundle in Codex. Codex loads the skills declared in its plugin manifest.",
        "",
        "```bash",
        "codex plugin marketplace add xonovex/platform",
        "codex plugin add xonovex-workflow@xonovex-marketplace",
        "```",
        "",
        "Plugin dependencies: `xonovex-core`, `xonovex-agentic`.",
        "",
        "Optional companions: `xonovex-native`, `xonovex-platform`.",
      ].join("\n"),
    );
  });
});
