import {spawnSync} from "node:child_process";
import {resolve} from "node:path";
import {describe, expect, it} from "vitest";

const workspaceRoot = resolve(import.meta.dirname, "../../../../..");
const validator = resolve(import.meta.dirname, "validate.py");

describe("portable skill validator", () => {
  it("accepts the same explicit product adapter as the canonical validator", () => {
    const skill = resolve(
      workspaceRoot,
      "packages/skill/skill-claude-code/claude-code-guide",
    );

    const result = spawnSync("uv", ["run", validator, "--strict", skill], {
      encoding: "utf8",
    });

    expect(result.status, `${result.stdout}\n${result.stderr}`).toBe(0);
  });
});
