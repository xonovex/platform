import {execFileSync} from "node:child_process";
import {
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {afterEach, describe, expect, it, vi} from "vitest";
import {main} from "./audit.js";

describe("main", () => {
  const temporaryDirectories: string[] = [];

  afterEach(() => {
    vi.restoreAllMocks();
    for (const directory of temporaryDirectories) {
      rmSync(directory, {recursive: true, force: true});
    }
    temporaryDirectories.length = 0;
  });

  function skillDirectory(name: string, sources: string): string {
    const root = mkdtempSync(join(tmpdir(), "skill-audit-sources-"));
    temporaryDirectories.push(root);
    const skill = join(root, name);
    mkdirSync(join(skill, "references"), {recursive: true});
    writeFileSync(join(skill, "SKILL.md"), `# ${name}\n`);
    writeFileSync(join(skill, "SOURCES.md"), sources);
    return skill;
  }

  it("renders help without auditing the filesystem", async () => {
    await expect(main(["--help"])).resolves.toBe(0);
  });

  it("reports a healthy skill as JSON", async () => {
    const skill = skillDirectory(
      "healthy-skill",
      `# Sources

## Primary source
- **URL:** https://example.com/guide
- **Version:** 2.4.0
- **References:** all
- **Last reviewed:** 2099-01-01
`,
    );
    writeFileSync(join(skill, "references", "details.md"), "# Details\n");
    const log = vi.spyOn(console, "log").mockImplementation(vi.fn());

    const result = await main([skill, "--json"]);

    expect(result).toBe(0);
    expect(JSON.parse(String(log.mock.calls[0]?.[0]))).toMatchObject({
      skill: "healthy-skill",
      source_count: 1,
      sources: [{version: "2.4.0", commit: null, watch_count: 0}],
      problems: 0,
    });
  });

  it("reports stale, unmapped, and uncovered sources", async () => {
    const skill = skillDirectory(
      "stale-skill",
      `# Sources

## Old source
- **URL:** https://example.com/old
- **Last reviewed:** 2020-01-01
`,
    );
    writeFileSync(join(skill, "references", "uncovered.md"), "# Uncovered\n");
    const log = vi.spyOn(console, "log").mockImplementation(vi.fn());

    const result = await main([skill, "--max-age=1"]);

    expect(result).toBe(1);
    const output = log.mock.calls.flat().join("\n");
    expect(output).toContain("STALE");
    expect(output).toContain("MISSING REFERENCE MAPPING");
    expect(output).toContain("references/uncovered.md");
  });

  it("marks a matching source as reviewed", async () => {
    const skill = skillDirectory(
      "reviewed-skill",
      `# Sources

## Primary source
- **URL:** https://example.com/guide
- **References:** all
- **Last reviewed:** 2020-01-01

## Secondary source
- **URL:** https://example.com/other
- **References:** all
- **Last reviewed:** 2020-01-01
`,
    );
    vi.spyOn(process.stderr, "write").mockImplementation(() => true);

    const result = await main([skill, "--mark-reviewed=Primary"]);

    expect(result).toBe(0);
    const updated = readFileSync(join(skill, "SOURCES.md"), "utf8");
    expect(updated).not.toContain(
      "## Primary source\n- **URL:** https://example.com/guide\n- **References:** all\n- **Last reviewed:** 2020-01-01",
    );
    expect(updated).toContain(
      "## Secondary source\n- **URL:** https://example.com/other\n- **References:** all\n- **Last reviewed:** 2020-01-01",
    );
  });

  it("audits every skill below a root", async () => {
    const first = skillDirectory(
      "first-skill",
      `# Sources

## Primary source
- **Provenance:** Repository-original guidance
- **References:** all
- **Last reviewed:** 2099-01-01
`,
    );
    const root = join(first, "..");
    const second = join(root, "second-skill");
    mkdirSync(second);
    writeFileSync(join(second, "SKILL.md"), "# second-skill\n");
    writeFileSync(
      join(second, "SOURCES.md"),
      `# Sources

## Primary source
- **Provenance:** Repository-original guidance
- **References:** all
- **Last reviewed:** 2099-01-01
`,
    );
    const log = vi.spyOn(console, "log").mockImplementation(vi.fn());
    vi.spyOn(process.stderr, "write").mockImplementation(() => true);

    const result = await main(["--all", root, "--json"]);

    expect(result).toBe(0);
    const report = JSON.parse(String(log.mock.calls[0]?.[0])) as unknown[];
    expect(report).toHaveLength(2);
  });

  it("reports a failed explicit upstream pull", async () => {
    const skill = skillDirectory(
      "pull-failure-skill",
      `# Sources

## Primary source
- **Provenance:** Repository-original guidance
- **References:** all
- **Checkout:** upstream
- **Last reviewed:** 2099-01-01
`,
    );
    const checkout = join(skill, "upstream");
    mkdirSync(checkout);
    execFileSync(resolveExecutable("git"), ["init", "--quiet"], {
      cwd: checkout,
    });
    execFileSync(
      resolveExecutable("git"),
      ["remote", "add", "origin", "file:///definitely/missing-upstream"],
      {cwd: checkout},
    );
    const log = vi.spyOn(console, "log").mockImplementation(vi.fn());

    const result = await main([skill, "--pull", "--json"]);

    expect(result).toBe(1);
    expect(JSON.parse(String(log.mock.calls[0]?.[0]))).toMatchObject({
      problems: 1,
      sources: [{drift: {pull_failed: true}}],
    });
  });
});
