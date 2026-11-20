import {describe, expect, it} from "vitest";
import {
  buildIsolatedClaudeArgs,
  buildTriggerClaudeArgs,
  TRIGGER_MAX_TURNS,
} from "../../../../src/claude/args.js";

// Every Claude eval invocation across the trigger, routing, and output harnesses
// is built from this argument set. A flag dropped here silently lets a run read
// the machine's real settings, so the whole set is pinned.
describe("Claude harness isolation", () => {
  it("denies a Claude run the machine's settings, MCP servers, and session", () => {
    const args = buildIsolatedClaudeArgs("stream-json");

    expect(args).toEqual([
      "-p",
      "--output-format",
      "stream-json",
      "--setting-sources",
      "",
      "--strict-mcp-config",
      "--mcp-config",
      '{"mcpServers":{}}',
      "--no-session-persistence",
      "--no-chrome",
    ]);
  });

  it("carries the requested output format", () => {
    const args = buildIsolatedClaudeArgs("json");

    expect(args.slice(0, 3)).toEqual(["-p", "--output-format", "json"]);
  });
});

describe("trigger process isolation", () => {
  it("exposes only Skill with bounded isolated execution", () => {
    const args = buildTriggerClaudeArgs({
      model: "haiku",
      budget: 0.05,
      pluginDirectories: ["dependency-directory", "plugin-directory"],
    });
    const toolIndex = args.indexOf("--tools");
    const settingIndex = args.indexOf("--setting-sources");

    expect(args.slice(toolIndex, toolIndex + 2)).toEqual(["--tools", "Skill"]);
    expect(args.slice(settingIndex, settingIndex + 2)).toEqual([
      "--setting-sources",
      "",
    ]);
    expect(args).toContain("--strict-mcp-config");
    expect(args).toContain("--no-session-persistence");
    expect(args).toContain("--max-budget-usd");
    expect(args).toContain("0.05");
    const turnsIndex = args.indexOf("--max-turns");
    expect(args.slice(turnsIndex, turnsIndex + 2)).toEqual([
      "--max-turns",
      String(TRIGGER_MAX_TURNS),
    ]);
    expect(args).not.toContain("Read");
    expect(args).not.toContain("Bash");
    expect(args.filter((argument) => argument === "--plugin-dir")).toHaveLength(
      2,
    );
  });
});
