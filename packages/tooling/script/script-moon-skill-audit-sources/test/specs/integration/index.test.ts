import {execFileSync, spawnSync} from "node:child_process";
import {mkdirSync, writeFileSync} from "node:fs";
import {join, resolve} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/adapters/executable-node";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import {describe, expect, it} from "vitest";
import {fetchHttpClient} from "../../../src/adapters/http-fetch.js";
import {gitUpstreamRepository} from "../../../src/adapters/upstream-git.js";
import {main} from "../../../src/audit.js";
import {skillDirectories} from "../../util/skill-directory.js";

// Audits on the real disk with the Node adapters and returns the exit code
// with the standard output.
const audit = async (argv: readonly string[]) => {
  const output = memoryOutput();
  const code = await main(argv, process.cwd(), {
    fs: nodeFileSystem,
    output,
    http: fetchHttpClient,
    openUpstream: gitUpstreamRepository,
    today: new Date(),
  });
  return {code, stdout: output.stdoutLines().join("\n")};
};

describe("main", () => {
  const skillDirectory = skillDirectories();

  it("keeps the shipped portable auditor aligned with the canonical schema", async () => {
    const skill = skillDirectory(
      "portable-parity-skill",
      `# Sources

## Versioned docs
- **URLs:**
  - https://example.com/one
  - https://example.com/two
- **Version:** 2.4.0
- **Content SHA256:** ${"c".repeat(64)}
- **References:** all
- **Last reviewed:** 2099-01-01

## Repository synthesis
- **Provenance:** Repository-original guidance
- **References:** all
- **Last reviewed:** 2099-01-01
`,
    );
    writeFileSync(join(skill, "references", "details.md"), "# Details\n");
    const canonicalRun = await audit([skill, "--json"]);
    const canonicalResult = canonicalRun.code;
    const canonical = JSON.parse(canonicalRun.stdout) as {
      problems: number;
      source_count: number;
      sources: readonly Record<string, unknown>[];
      uncovered_refs: readonly string[];
    };
    const portable = resolve(
      import.meta.dirname,
      "../../../../../../runtime/plugin/plugin-agentic/skills/authoring-skills/scripts/audit-sources.py",
    );
    // The script declares no dependencies and imports only the standard library,
    // so uv has nothing to fetch: --offline and a refused interpreter download
    // turn any future dependency into a visible failure rather than a silent
    // download.
    const portableResult = spawnSync(
      resolveExecutable("uv"),
      ["run", "--offline", "--script", portable, skill, "--json"],
      {
        encoding: "utf8",
        env: {...process.env, UV_PYTHON_DOWNLOADS: "never"},
      },
    );
    const portableReport = JSON.parse(
      portableResult.stdout,
    ) as typeof canonical;
    const comparable = (report: typeof canonical) => ({
      problems: report.problems,
      source_count: report.source_count,
      uncovered_refs: report.uncovered_refs,
      sources: report.sources.map((source) => ({
        title: source.title,
        urls: source.urls,
        provenance: source.provenance ?? null,
        version: source.version,
        content_sha256: source.content_sha256,
        covers_all_references: source.covers_all_references,
        reference_mapping_missing: source.reference_mapping_missing,
        drift_anchor_missing: source.drift_anchor_missing,
      })),
    });

    expect(canonicalResult).toBe(0);
    expect(
      portableResult.status,
      `${portableResult.stdout}\n${portableResult.stderr}`,
    ).toBe(0);
    expect(comparable(portableReport)).toEqual(comparable(canonical));
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
    const result = await audit([skill, "--pull", "--json"]);

    expect(result.code).toBe(1);
    expect(JSON.parse(result.stdout)).toMatchObject({
      problems: 1,
      sources: [{drift: {pull_failed: true}}],
    });
  });
});
