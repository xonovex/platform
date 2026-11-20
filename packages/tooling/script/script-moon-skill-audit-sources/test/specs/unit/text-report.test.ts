import {describe, expect, it} from "vitest";
import type {SourceReport} from "../../../src/source-audit.js";
import {renderJsonReport, renderTextReport} from "../../../src/text-report.js";
import type {DriftReport} from "../../../src/upstream-drift.js";

const DRIFT: DriftReport = {
  checkout: "/up",
  resolved: true,
  pull_failed: true,
  pinned_version: "1.0.0",
  latest_version: null,
  behind: false,
  commit_count: 2,
  commits: ["c1", "c2"],
  review_refs: [],
  note: "partial",
};

const SOURCE: SourceReport = {
  title: "Library",
  url: "https://example.com",
  urls: ["https://example.com"],
  provenance: "Upstream docs",
  last_reviewed: "2020-01-01",
  age_days: 400,
  review_max_age_days: 90,
  stale: true,
  refs: ["api.md"],
  covers_all_references: false,
  reference_mapping_missing: false,
  dangling_refs: ["gone.md"],
  version: "1.0.0",
  commit: null,
  watch_count: 1,
  content_sha256: "a".repeat(64),
  fetched_content_sha256: null,
  content_changed: true,
  drift_anchor_missing: true,
  fetches: [{status: "missing", detail: "HTTP 404"}],
  drift: DRIFT,
};

const render = (source: SourceReport) =>
  renderTextReport(
    {
      skill: "s",
      sources_file: "s/SOURCES.md",
      source_count: 1,
      sources: [source],
      uncovered_refs: [],
      problems: 1,
    },
    180,
    90,
  );

describe("renderTextReport", () => {
  it("flags every problem of a source and lists its upstream drift", () => {
    expect(render(SOURCE).slice(2, -1)).toEqual([
      "\n  [PULL FAILED  STALE (400d)  MISSING DRIFT ANCHOR  CONTENT CHANGED  DANGLING: gone.md  URL missing  DRIFT (2 commit(s))] Library",
      "    urls          : https://example.com",
      "    provenance    : Upstream docs",
      "    last reviewed : 2020-01-01  (400d ago)",
      "    fetch 1      : HTTP 404",
      `    content pin    : ${"a".repeat(64)}`,
      "    fetched hash   : (unavailable)",
      "    feeds         : references/api.md",
      "    upstream      : pinned 1.0.0, up to date with latest released (no tags)",
      "                    note: partial",
      "    commits since : 2 on watched paths",
      "                    c1",
      "                    c2",
      "    review        : references/api.md",
    ]);
  });

  it("names a source without a review date, mapping or resolved checkout", () => {
    const lines = render({
      ...SOURCE,
      urls: [],
      provenance: undefined,
      last_reviewed: null,
      age_days: null,
      stale: false,
      refs: [],
      reference_mapping_missing: true,
      dangling_refs: [],
      content_sha256: null,
      content_changed: false,
      drift_anchor_missing: false,
      fetches: undefined,
      fetch: {status: "ok", detail: "HTTP 200"},
      drift: {
        ...DRIFT,
        resolved: false,
        note: undefined,
        pull_failed: false,
      },
    });

    expect(lines.slice(2, -1)).toEqual([
      "\n  [MISSING REVIEW DATE  MISSING REFERENCE MAPPING  DRIFT (2 commit(s))] Library",
      "    urls          : (none)",
      "    last reviewed : (none)",
      "    fetch 1      : HTTP 200",
      "    feeds         : (missing)",
      "    upstream      : unresolved",
    ]);
  });
});

describe("renderJsonReport", () => {
  it("escapes non-ASCII text the way the portable auditor does", () => {
    expect(
      renderJsonReport([
        {
          skill: "café → 😀",
          sources_file: "x",
          source_count: 0,
          sources: [],
          uncovered_refs: [],
          problems: 1,
        },
      ]),
    ).toContain('"skill": "caf\\u00e9 \\u2192 \\ud83d\\ude00"');
  });
});
