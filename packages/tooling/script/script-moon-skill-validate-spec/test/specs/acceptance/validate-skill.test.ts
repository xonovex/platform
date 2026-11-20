import {cpSync, mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join, resolve} from "node:path";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import type {DriftLintMode} from "@xonovex/script-moon-common/drift-budgets";
import {afterEach, describe, expect, it} from "vitest";
import {main} from "../../../src/validate-skill.js";

// Validates against the real disk and returns the exit code with the report.
const validateSkill = (
  argv: readonly string[],
  driftMode: DriftLintMode = "warn",
) => {
  const output = memoryOutput();
  const code = main(argv, process.cwd(), {
    fs: nodeFileSystem,
    output,
    driftMode,
  });
  return {code, report: output.stdoutLines().join("\n")};
};

// These validate the catalog's own auditing-code-quality skill rather than a
// fixture, so they state what the repository's content must satisfy. The rules
// themselves are covered against built trees in the unit tier.
const temporaryDirectories: string[] = [];

afterEach(() => {
  for (const directory of temporaryDirectories) {
    rmSync(directory, {recursive: true, force: true});
  }
  temporaryDirectories.length = 0;
});

describe("the live skill catalog", () => {
  it("validates a complete repository skill", () => {
    const skill = resolve(
      import.meta.dirname,
      "../../../../../../runtime/plugin/plugin-core/skills/auditing-code-quality",
    );
    expect(validateSkill(["--strict", skill]).code).toBe(0);
  });

  it("keeps drift findings advisory in warn mode and blocking in enforce mode", () => {
    // A copied skill in a throwaway workspace keeps the assertion independent of
    // whatever the real catalog currently measures.
    const root = mkdtempSync(join(tmpdir(), "skill-drift-"));
    temporaryDirectories.push(root);
    mkdirSync(join(root, ".moon"));
    const skillDirectory = join(
      root,
      "packages",
      "runtime",
      "plugin",
      "plugin-core",
      "skills",
      "auditing-code-quality",
    );
    mkdirSync(join(root, "packages", "runtime", "plugin"), {recursive: true});
    cpSync(
      resolve(
        import.meta.dirname,
        "../../../../../../runtime/plugin/plugin-core/skills/auditing-code-quality",
      ),
      skillDirectory,
      {recursive: true},
    );
    writeFileSync(
      join(root, "budgets.json"),
      JSON.stringify({
        "packages/runtime/plugin/plugin-core/skills/auditing-code-quality/SKILL.md": 1,
      }),
      "utf8",
    );
    const advisory = validateSkill(["--strict", skillDirectory]);
    expect(advisory.code).toBe(0);
    expect(advisory.report).toContain("[DRIFT] budget:");

    expect(validateSkill(["--strict", skillDirectory], "enforce").code).toBe(1);
  });
  it("allows the catalog's explicit Claude Code adapter name", () => {
    const skill = resolve(
      import.meta.dirname,
      "../../../../../../runtime/plugin/plugin-agentic/skills/configuring-claude-code",
    );
    expect(validateSkill(["--strict", skill]).code).toBe(0);
  });
});
