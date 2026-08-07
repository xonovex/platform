import {createHash} from "node:crypto";
import {mkdirSync, readFileSync, writeFileSync} from "node:fs";
import {join} from "node:path";
import {afterEach, describe, expect, it, vi} from "vitest";
import {main} from "../../../src/audit.js";
import {skillDirectories} from "../../util/skill-directory.js";

describe("main", () => {
  const skillDirectory = skillDirectories();

  afterEach(() => {
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

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
- **Content SHA256:** ${"a".repeat(64)}
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
    const skill = skillDirectory(
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
    const log = vi.spyOn(console, "log").mockImplementation(vi.fn());

    const result = await main([
      skill,
      "--max-age=180",
      "--version-max-age=1",
      "--json",
    ]);

    expect(result).toBe(1);
    expect(JSON.parse(String(log.mock.calls[0]?.[0]))).toMatchObject({
      sources: [{stale: true, review_max_age_days: 1}],
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

  it("requires a drift anchor for versioned web documentation", async () => {
    const skill = skillDirectory(
      "unanchored-skill",
      `# Sources

## Versioned source
- **URL:** https://example.com/versioned
- **Version:** 1.2.3
- **References:** all
- **Last reviewed:** 2099-01-01
`,
    );
    const log = vi.spyOn(console, "log").mockImplementation(vi.fn());

    const result = await main([skill, "--json"]);

    expect(result).toBe(1);
    expect(JSON.parse(String(log.mock.calls[0]?.[0]))).toMatchObject({
      problems: 1,
      sources: [{drift_anchor_missing: true}],
    });
  });

  it.each([401, 403, 429])(
    "does not fail a live check solely because a source returns HTTP %i",
    async (status) => {
      const skill = skillDirectory(
        "restricted-skill",
        `# Sources

## Restricted source
- **URL:** https://example.com/restricted
- **References:** all
- **Last reviewed:** 2099-01-01
`,
      );
      vi.stubGlobal(
        "fetch",
        vi.fn().mockResolvedValue(new Response("restricted", {status})),
      );
      const log = vi.spyOn(console, "log").mockImplementation(vi.fn());

      const result = await main([skill, "--fetch", "--json"]);

      expect(result).toBe(0);
      expect(JSON.parse(String(log.mock.calls[0]?.[0]))).toMatchObject({
        problems: 0,
        sources: [
          {fetches: [{status: "restricted", detail: `HTTP ${String(status)}`}]},
        ],
      });
    },
  );

  it("retries a transient live-check error once", async () => {
    const skill = skillDirectory(
      "retry-skill",
      `# Sources

## Retry source
- **URL:** https://example.com/retry
- **References:** all
- **Last reviewed:** 2099-01-01
`,
    );
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(new Response("unavailable", {status: 503}))
      .mockResolvedValueOnce(new Response("available", {status: 200}));
    vi.stubGlobal("fetch", fetch);
    const log = vi.spyOn(console, "log").mockImplementation(vi.fn());

    const result = await main([skill, "--fetch", "--json"]);

    expect(result).toBe(0);
    expect(fetch).toHaveBeenCalledTimes(2);
    expect(JSON.parse(String(log.mock.calls[0]?.[0]))).toMatchObject({
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
    const skill = skillDirectory(
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
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue(new Response("stable body", {status: 200})),
    );
    const log = vi.spyOn(console, "log").mockImplementation(vi.fn());

    const result = await main([skill, "--fetch", "--json"]);

    expect(result).toBe(0);
    expect(JSON.parse(String(log.mock.calls[0]?.[0]))).toMatchObject({
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
});
