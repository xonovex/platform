import type {SkillReport, SourceReport} from "./source-audit.js";

const sourceFetches = (source: SourceReport) =>
  source.fetches ?? (source.fetch ? [source.fetch] : []);

const sourceFlags = (source: SourceReport): readonly string[] => {
  const flags: string[] = [];
  if (source.drift?.pull_failed) flags.push("PULL FAILED");
  if (source.stale) flags.push(`STALE (${String(source.age_days)}d)`);
  if (source.last_reviewed === null) flags.push("MISSING REVIEW DATE");
  if (source.reference_mapping_missing) {
    flags.push("MISSING REFERENCE MAPPING");
  }
  if (source.drift_anchor_missing) flags.push("MISSING DRIFT ANCHOR");
  if (source.content_changed) flags.push("CONTENT CHANGED");
  if (source.dangling_refs.length > 0) {
    flags.push(`DANGLING: ${source.dangling_refs.join(", ")}`);
  }
  const failedFetch = sourceFetches(source).find(
    (fetch) => fetch.status !== "ok",
  );
  if (failedFetch !== undefined) flags.push(`URL ${failedFetch.status}`);
  if (source.drift?.behind) {
    flags.push(
      `BEHIND ${source.drift.pinned_version ?? "?"} -> ${source.drift.latest_version ?? "?"}`,
    );
  } else if (source.drift !== undefined && source.drift.commit_count > 0) {
    flags.push(`DRIFT (${String(source.drift.commit_count)} commit(s))`);
  }
  return flags;
};

const driftLines = (source: SourceReport): readonly string[] => {
  const drift = source.drift;
  if (drift === undefined) return [];
  if (!drift.resolved) {
    return [`    upstream      : ${drift.note ?? "unresolved"}`];
  }
  const latest = drift.latest_version ?? "(no tags)";
  const state = drift.behind
    ? `behind — latest released ${latest}`
    : `up to date with latest released ${latest}`;
  const header = [
    `    upstream      : pinned ${drift.pinned_version ?? "?"}, ${state}`,
    ...(drift.note ? [`                    note: ${drift.note}`] : []),
  ];
  if (drift.commit_count === 0) return header;
  const review = drift.review_refs.length > 0 ? drift.review_refs : source.refs;
  return [
    ...header,
    `    commits since : ${String(drift.commit_count)} on watched paths`,
    ...drift.commits.map((commit) => `                    ${commit}`),
    ...(review.length > 0
      ? [
          `    review        : ${review.map((ref) => "references/" + ref).join(", ")}`,
        ]
      : []),
  ];
};

const sourceFeeds = (source: SourceReport): string => {
  if (source.covers_all_references) return "all references";
  if (source.refs.length === 0) return "(missing)";
  return source.refs.map((ref) => `references/${ref}`).join(", ");
};

const sourceLines = (source: SourceReport): readonly string[] => {
  const flags = sourceFlags(source);
  const marker = flags.length > 0 ? flags.join("  ") : "ok";
  const age =
    source.age_days === null ? "" : `  (${String(source.age_days)}d ago)`;
  return [
    `\n  [${marker}] ${source.title}`,
    `    urls          : ${source.urls.length > 0 ? source.urls.join(", ") : "(none)"}`,
    ...(source.provenance ? [`    provenance    : ${source.provenance}`] : []),
    `    last reviewed : ${source.last_reviewed ?? "(none)"}${age}`,
    ...sourceFetches(source).map(
      (fetch, index) =>
        `    fetch ${String(index + 1).padEnd(7)}: ${fetch.detail}`,
    ),
    ...(source.content_sha256 === null
      ? []
      : [
          `    content pin    : ${source.content_sha256}`,
          `    fetched hash   : ${source.fetched_content_sha256 ?? "(unavailable)"}`,
        ]),
    `    feeds         : ${sourceFeeds(source)}`,
    ...driftLines(source),
  ];
};

/** The text report of one skill. */
export const renderTextReport = (
  report: SkillReport,
  maxAge: number,
  versionMaxAge: number,
): readonly string[] => [
  `skill: ${report.skill}  (${report.sources_file})`,
  `sources: ${String(report.source_count)}  max-age: ${String(maxAge)}d  ` +
    `version-max-age: ${String(versionMaxAge)}d`,
  ...report.sources.flatMap((source) => sourceLines(source)),
  ...(report.uncovered_refs.length > 0
    ? [
        "\n  reference files with no declared source or repository-original provenance:",
        ...report.uncovered_refs.map((ref) => `    references/${ref}`),
      ]
    : []),
  "\nNote: distilled prose is not auto-rewritten. Review the reference files a " +
    "stale/changed source feeds, then re-run with --mark-reviewed to stamp the date.",
];

// The JSON report matches the escaping of Python's json.dumps, which the
// portable auditor uses: every non-ASCII UTF-16 code unit becomes \uXXXX.
const ensureAscii = (text: string): string =>
  text.replaceAll(/[\u0080-\uFFFF]/g, (ch) => {
    // The non-Unicode regex visits each half of a surrogate pair separately.
    const codePoint = ch.codePointAt(0);
    if (codePoint === undefined) return ch;
    const code = codePoint.toString(16).padStart(4, "0");
    return String.raw`\u${code}`;
  });

/** The JSON report: one object for one skill, else an array. */
export const renderJsonReport = (reports: readonly SkillReport[]): string =>
  ensureAscii(
    JSON.stringify(reports.length === 1 ? reports[0] : reports, null, 2),
  );
