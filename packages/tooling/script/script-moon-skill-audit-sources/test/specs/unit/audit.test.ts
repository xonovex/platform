import {createHash} from "node:crypto";
import {join} from "node:path";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {describe, expect, it, vi} from "vitest";
import {main} from "../../../src/audit.js";
import type {HttpClient, HttpResponse} from "../../../src/http.js";
import type {UpstreamRepository} from "../../../src/upstream.js";
import {
  MEMORY_ROOT,
  memorySkillDirectories,
} from "../../util/skill-directory.js";

const TODAY = new Date(2026, 4, 20);

const response = (status: number, body = ""): HttpResponse => ({
  status,
  lastModified: undefined,
  body: new TextEncoder().encode(body),
});

// An HTTP client that answers each request with the next response in turn.
const httpClient = (...responses: readonly HttpResponse[]) => {
  const get = vi.fn<HttpClient["get"]>();
  for (const answer of responses) get.mockResolvedValueOnce(answer);
  return {get};
};

const noUpstream = (): UpstreamRepository => {
  throw new Error("no upstream checkout in this case");
};

// Runs the auditor from the memory root and returns the exit code with its output.
const run = async (
  argv: readonly string[],
  fs: FileSystem,
  http: HttpClient = httpClient(),
) => {
  const output = memoryOutput();
  const code = await main(argv, MEMORY_ROOT, {
    fs,
    output,
    http,
    openUpstream: noUpstream,
    today: TODAY,
  });
  return {code, stdout: output.stdoutLines(), stderr: output.stderrLines()};
};

const jsonReport = (lines: readonly string[]): unknown =>
  JSON.parse(lines.join("\n"));

describe("main", () => {
  const skillDirectory = memorySkillDirectories();

  it("renders help without auditing the filesystem", async () => {
    const result = await run(["--help"], skillDirectory("x", "").fs);

    expect(result.code).toBe(0);
    expect(result.stdout[0]).toContain("Audit a skill's upstream sources");
  });

  it("reports an invalid argument", async () => {
    expect(
      await run(["--max-age=x"], skillDirectory("x", "").fs),
    ).toMatchObject({
      code: 2,
      stderr: [
        "error: argument --max-age: invalid non-negative int value: 'x'",
      ],
    });
  });

  it("reports a healthy skill as JSON", async () => {
    const {fs, skill} = skillDirectory(
      "healthy-skill",
      `# Sources

## Primary source
- **URL:** https://example.com/guide
- **Version:** 2.4.0
- **Content SHA256:** ${"a".repeat(64)}
- **References:** all
- **Last reviewed:** 2099-01-01
`,
    );
    fs.writeFile(join(skill, "references", "details.md"), "# Details\n");

    const result = await run([skill, "--json"], fs);

    expect(result.code).toBe(0);
    expect(jsonReport(result.stdout)).toMatchObject({
      skill: "healthy-skill",
      source_count: 1,
      sources: [
        {
          version: "2.4.0",
          commit: null,
          watch_count: 0,
          content_sha256: "a".repeat(64),
          drift_anchor_missing: false,
          review_max_age_days: 90,
        },
      ],
      problems: 0,
    });
  });

  it("uses the shorter review cadence for versioned sources", async () => {
    const {fs, skill} = skillDirectory(
      "versioned-skill",
      `# Sources

## Versioned source
- **URL:** https://example.com/versioned
- **Version:** 1.2.3
- **Content SHA256:** ${"b".repeat(64)}
- **References:** all
- **Last reviewed:** 2026-05-13
`,
    );

    const result = await run(
      [skill, "--max-age=180", "--version-max-age=1", "--json"],
      fs,
    );

    expect(result.code).toBe(1);
    expect(jsonReport(result.stdout)).toMatchObject({
      sources: [{stale: true, review_max_age_days: 1}],
    });
  });

  it("reports stale, unmapped, and uncovered sources", async () => {
    const {fs, skill} = skillDirectory(
      "stale-skill",
      `# Sources

## Old source
- **URL:** https://example.com/old
- **Last reviewed:** 2020-01-01
`,
    );
    fs.writeFile(join(skill, "references", "uncovered.md"), "# Uncovered\n");

    const result = await run([skill, "--max-age=1"], fs);

    expect(result.code).toBe(1);
    const output = result.stdout.join("\n");
    expect(output).toContain("STALE");
    expect(output).toContain("MISSING REFERENCE MAPPING");
    expect(output).toContain("references/uncovered.md");
  });

  it("requires a drift anchor for versioned web documentation", async () => {
    const {fs, skill} = skillDirectory(
      "unanchored-skill",
      `# Sources

## Versioned source
- **URL:** https://example.com/versioned
- **Version:** 1.2.3
- **References:** all
- **Last reviewed:** 2099-01-01
`,
    );

    const result = await run([skill, "--json"], fs);

    expect(result.code).toBe(1);
    expect(jsonReport(result.stdout)).toMatchObject({
      problems: 1,
      sources: [{drift_anchor_missing: true}],
    });
  });

  it.each([401, 403, 429])(
    "does not fail a live check solely because a source returns HTTP %i",
    async (status) => {
      const {fs, skill} = skillDirectory(
        "restricted-skill",
        `# Sources

## Restricted source
- **URL:** https://example.com/restricted
- **References:** all
- **Last reviewed:** 2099-01-01
`,
      );
      const result = await run(
        [skill, "--fetch", "--json"],
        fs,
        httpClient(response(status, "restricted")),
      );

      expect(result.code).toBe(0);
      expect(jsonReport(result.stdout)).toMatchObject({
        problems: 0,
        sources: [
          {fetches: [{status: "restricted", detail: `HTTP ${String(status)}`}]},
        ],
      });
    },
  );

  it("retries a transient live-check error once", async () => {
    const {fs, skill} = skillDirectory(
      "retry-skill",
      `# Sources

## Retry source
- **URL:** https://example.com/retry
- **References:** all
- **Last reviewed:** 2099-01-01
`,
    );
    const http = httpClient(response(503), response(200, "available"));

    const result = await run([skill, "--fetch", "--json"], fs, http);

    expect(result.code).toBe(0);
    expect(http.get).toHaveBeenCalledTimes(2);
    expect(jsonReport(result.stdout)).toMatchObject({
      problems: 0,
      sources: [{fetches: [{status: "ok", detail: "HTTP 200"}]}],
    });
  });

  it("verifies the ordered content digest for versioned web documentation", async () => {
    const url = "https://example.com/versioned";
    const bodyDigest = createHash("sha256").update("stable body").digest("hex");
    const contentDigest = createHash("sha256")
      .update(url)
      .update("\0")
      .update(bodyDigest)
      .update("\n")
      .digest("hex");
    const {fs, skill} = skillDirectory(
      "snapshot-skill",
      `# Sources

## Versioned source
- **URL:** ${url}
- **Version:** 1.2.3
- **Content SHA256:** ${contentDigest}
- **References:** all
- **Last reviewed:** 2099-01-01
`,
    );
    const result = await run(
      [skill, "--fetch", "--json"],
      fs,
      httpClient(response(200, "stable body")),
    );

    expect(result.code).toBe(0);
    expect(jsonReport(result.stdout)).toMatchObject({
      problems: 0,
      sources: [
        {
          content_sha256: contentDigest,
          fetched_content_sha256: contentDigest,
          content_changed: false,
        },
      ],
    });
  });

  it("marks a matching source as reviewed", async () => {
    const {fs, skill} = skillDirectory(
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

    const result = await run([skill, "--mark-reviewed=Primary"], fs);

    expect(result).toMatchObject({
      code: 0,
      stderr: ["stamped 'Last reviewed' = 2026-05-20 for: Primary source"],
    });
    const updated = fs.readText(join(skill, "SOURCES.md"));
    expect(updated).not.toContain(
      "## Primary source\n- **URL:** https://example.com/guide\n- **References:** all\n- **Last reviewed:** 2020-01-01",
    );
    expect(updated).toContain(
      "## Secondary source\n- **URL:** https://example.com/other\n- **References:** all\n- **Last reviewed:** 2020-01-01",
    );
  });

  it("audits every skill below a root", async () => {
    const {fs, skill: first} = skillDirectory(
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
    fs.writeFile(join(second, "SKILL.md"), "# second-skill\n");
    fs.writeFile(
      join(second, "SOURCES.md"),
      `# Sources

## Primary source
- **Provenance:** Repository-original guidance
- **References:** all
- **Last reviewed:** 2099-01-01
`,
    );

    const result = await run(["--all", root, "--json"], fs);

    expect(result.code).toBe(0);
    const report = jsonReport(result.stdout) as unknown[];
    expect(report).toHaveLength(2);
  });

  it("prints a text report per skill and a summary for several", async () => {
    const {fs, skill: first} = skillDirectory(
      "first-skill",
      "# Sources\n\n## Old\n- **URL:** https://example.com/old\n- **Last reviewed:** 2020-01-01\n",
    );
    const second = join(first, "..", "second-skill");
    fs.writeFile(join(second, "SKILL.md"), "# second-skill\n");
    fs.writeFile(join(second, "SOURCES.md"), "# Sources\n");

    const result = await run(["--all"], fs);

    expect(result.code).toBe(1);
    expect(result.stdout).toContain("\n" + "=".repeat(60));
    expect(result.stdout[0]).toBe(
      "skill: first-skill  (first-skill/SOURCES.md)",
    );
    expect(result.stderr).toEqual([
      "\n--- 2 skills, 3 problem(s), 1 stale source(s) ---",
    ]);
  });

  it.each([
    [["--all", "absent"], "error: --all root not a directory: absent"],
    [["--all", "empty"], "error: no SOURCES.md found under empty"],
    [["absent"], "error: no SOURCES.md at absent"],
  ])("reports the missing target of %j", async (argv, message) => {
    const {fs} = skillDirectory("x", "# Sources\n");
    fs.writeFile(join(MEMORY_ROOT, "empty", "notes.md"), "");

    expect(await run(argv, fs)).toMatchObject({code: 2, stderr: [message]});
  });

  it("finds the skill nested in a target directory", async () => {
    const {fs} = skillDirectory("nested-skill", "# Sources\n");

    const result = await run([".", "--json"], fs);

    expect(jsonReport(result.stdout)).toMatchObject({
      sources_file: "nested-skill/SOURCES.md",
    });
  });

  it("names an ambiguous target as written", async () => {
    const {fs} = skillDirectory("one-skill", "# Sources\n");
    fs.writeFile(join(MEMORY_ROOT, "two-skill", "SKILL.md"), "# two\n");

    await expect(run(["./"], fs)).rejects.toThrow(
      "multiple SKILL.md found under ./; pass one explicitly",
    );
  });

  it("refuses to stamp several skills and reports when nothing matches", async () => {
    const {fs, skill} = skillDirectory(
      "one-skill",
      "# Sources\n\n## Only\n- **Last reviewed:** 2020-01-01\n",
    );
    fs.writeFile(join(MEMORY_ROOT, "two-skill", "SKILL.md"), "# two\n");
    fs.writeFile(join(MEMORY_ROOT, "two-skill", "SOURCES.md"), "# Sources\n");

    expect(await run(["--all", "--mark-reviewed"], fs)).toMatchObject({
      code: 2,
      stderr: ["error: --mark-reviewed operates on a single skill"],
    });
    expect(await run([skill, "--mark-reviewed=none"], fs)).toMatchObject({
      code: 0,
      stderr: ["no matching source blocks to stamp"],
    });
  });

  it("compares a pinned checkout with its upstream", async () => {
    const {fs, skill} = skillDirectory(
      "pinned-skill",
      "# Sources\n\n## Library\n- **Provenance:** Upstream library\n- **References:** all\n- **Checkout:** upstream\n- **Version:** 1.0.0\n- **Last reviewed:** 2099-01-01\n",
    );
    fs.writeFile(join(MEMORY_ROOT, ".moon", "workspace.yml"), "");
    fs.writeFile(join(MEMORY_ROOT, "upstream", ".git", "HEAD"), "");
    const upstream: UpstreamRepository = {
      fetchTags: () => true,
      tags: () => ["v1.0.0", "v1.1.0"],
      hasCommit: () => true,
      commits: () => ["abc123 change"],
      changedFiles: () => [],
    };
    const output = memoryOutput();

    const code = await main([skill], MEMORY_ROOT, {
      fs,
      output,
      http: httpClient(),
      openUpstream: () => upstream,
      today: TODAY,
    });

    expect(code).toBe(1);
    expect(output.stdoutLines()).toContain(
      "\n  [BEHIND 1.0.0 -> v1.1.0] Library",
    );
  });
});
