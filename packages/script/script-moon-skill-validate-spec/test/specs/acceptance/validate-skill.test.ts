import {cpSync, mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join, resolve} from "node:path";
import {afterEach, describe, expect, it, vi} from "vitest";
import {main as validateSkill} from "../../../src/validate-skill.js";

// These validate the catalog's own skill-code-quality package rather than a
// fixture, so they state what the repository's content must satisfy. The rules
// themselves are covered against built trees in the unit tier.
const temporaryDirectories: string[] = [];

afterEach(() => {
  vi.restoreAllMocks();
  for (const directory of temporaryDirectories) {
    rmSync(directory, {recursive: true, force: true});
  }
  temporaryDirectories.length = 0;
});

describe("the live skill catalog", () => {
  it("validates a complete repository skill", () => {
    const skill = resolve(
      import.meta.dirname,
      "../../../../../skill/skill-code-quality",
    );
    vi.spyOn(console, "log").mockImplementation(() => {});

    const result = validateSkill(["--strict", skill]);

    expect(result).toBe(0);
  });

  it("keeps drift findings advisory in warn mode and blocking in enforce mode", () => {
    // A copied skill in a throwaway workspace keeps the assertion independent of
    // whatever the real catalog currently measures.
    const root = mkdtempSync(join(tmpdir(), "skill-drift-"));
    temporaryDirectories.push(root);
    mkdirSync(join(root, ".moon"));
    const guide = join(root, "packages", "skill", "skill-code-quality");
    mkdirSync(join(root, "packages", "skill"), {recursive: true});
    cpSync(
      resolve(import.meta.dirname, "../../../../../skill/skill-code-quality"),
      guide,
      {recursive: true},
    );
    writeFileSync(
      join(root, "budgets.json"),
      JSON.stringify({
        "packages/skill/skill-code-quality/code-quality-guide/SKILL.md": 1,
      }),
      "utf8",
    );
    const log = vi.spyOn(console, "log").mockImplementation(() => {});

    expect(validateSkill(["--strict", guide])).toBe(0);
    expect(log.mock.calls.flat().join("\n")).toContain("[DRIFT] budget:");

    process.env.XONOVEX_LINT_MODE = "enforce";
    try {
      expect(validateSkill(["--strict", guide])).toBe(1);
    } finally {
      delete process.env.XONOVEX_LINT_MODE;
    }
  });
  it("allows the catalog's explicit Claude Code adapter name", () => {
    const skill = resolve(
      import.meta.dirname,
      "../../../../../skill/skill-claude-code",
    );
    vi.spyOn(console, "log").mockImplementation(() => {});

    const result = validateSkill(["--strict", skill]);

    expect(result).toBe(0);
  });
});
