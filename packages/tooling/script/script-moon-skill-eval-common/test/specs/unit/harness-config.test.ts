import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {describe, expect, it} from "vitest";
import {
  declaredSkillName,
  harnessInput,
  locateHarness,
  parseHarness,
  resolveHarnessModel,
  shortSkillName,
} from "../../../src/harness-config.js";

describe("harness selection", () => {
  it("takes the flag, then SKILL_EVAL_HARNESS, then claude", () => {
    expect(harnessInput("codex", {SKILL_EVAL_HARNESS: "claude"})).toBe("codex");
    expect(harnessInput(undefined, {SKILL_EVAL_HARNESS: "codex"})).toBe(
      "codex",
    );
    expect(harnessInput(undefined, {})).toBe("claude");
  });

  it("accepts only a known harness", () => {
    expect(parseHarness("codex")).toBe("codex");
    expect(parseHarness("gemini")).toBeUndefined();
  });
});

describe("locateHarness", () => {
  const fs = memoryFileSystem({
    files: {"/opt/bin/claude": "", "/opt/win/codex.EXE": ""},
    executables: ["/opt/bin/claude", "/opt/win/codex.EXE"],
  });

  it("finds the harness on the PATH of the environment it is given", () => {
    expect(
      locateHarness("claude", {PATH: "/usr/bin:/opt/bin"}, "linux", fs),
    ).toBe("/opt/bin/claude");
    expect(
      locateHarness("codex", {PATH: "/opt/win", PATHEXT: ".EXE"}, "win32", fs),
    ).toMatch(/codex\.EXE$/);
  });

  it("reports an absent harness or an empty PATH as undefined", () => {
    expect(
      locateHarness("codex", {PATH: "/opt/bin"}, "linux", fs),
    ).toBeUndefined();
    expect(locateHarness("claude", {}, "linux", fs)).toBeUndefined();
  });
});

describe("resolveHarnessModel", () => {
  it("takes the flag, then the variable of the harness role, then the default", () => {
    expect(
      resolveHarnessModel("claude", "generation", "flag-model", {
        CLAUDE_MODEL: "env-model",
      }),
    ).toBe("flag-model");
    expect(
      resolveHarnessModel("codex", "judge", undefined, {
        CODEX_JUDGE_MODEL: "codex-judge",
        JUDGE_MODEL: "claude-judge",
      }),
    ).toBe("codex-judge");
    expect(
      resolveHarnessModel("claude", "judge", undefined, {
        JUDGE_MODEL: "claude-judge",
      }),
    ).toBe("claude-judge");
    expect(resolveHarnessModel("codex", "generation", " ", {})).toBe(
      "gpt-5.3-codex",
    );
  });
});

describe("skill names", () => {
  it("drops a plugin prefix", () => {
    expect(shortSkillName("plugin:using-git")).toBe("using-git");
    expect(shortSkillName("using-git")).toBe("using-git");
  });

  it("reads the name the SKILL.md frontmatter declares", () => {
    const fs = memoryFileSystem({
      files: {"/skill/SKILL.md": "---\nname: using-git\n---\n# Git\n"},
    });

    expect(declaredSkillName("/skill", fs)).toBe("using-git");
    expect(declaredSkillName("/other", fs)).toBeUndefined();
  });
});
