import {describe, expect, it} from "vitest";
import {buildIsolatedCodexArgs} from "../../../../src/codex/args.js";

// Every Codex eval invocation across the trigger, routing, and output harnesses is
// built from this argument set. A flag dropped here silently lets a run read the
// machine's real settings, so the whole set is pinned.
describe("Codex harness isolation", () => {
  it("denies a Codex run user config, rules, and write access", () => {
    expect(buildIsolatedCodexArgs({model: ""})).toEqual([
      "exec",
      "--json",
      "--ephemeral",
      "--sandbox",
      "read-only",
      "--ignore-user-config",
      "--ignore-rules",
      "--skip-git-repo-check",
    ]);
  });

  it("appends only a non-empty Codex model", () => {
    expect(buildIsolatedCodexArgs({model: "gpt-5"}).slice(-2)).toEqual([
      "--model",
      "gpt-5",
    ]);
  });
});
