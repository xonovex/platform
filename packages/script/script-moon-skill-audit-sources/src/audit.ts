import {execFileSync} from "node:child_process";
import {readdirSync, readFileSync, writeFileSync} from "node:fs";
import {basename, join, resolve, sep} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {
  isDirectory,
  isFile,
  resolveGuideDirectory,
} from "@xonovex/script-moon-common/fs";
import {parseArgs, type ParsedArgs} from "./args.js";
import {
  hasReferenceMapping,
  parseSources,
  type Source,
} from "./source-parser.js";

const REVIEWED_RE = /\*\*Last reviewed:\*\*\s*(\d{4}-\d{2}-\d{2})/;
const HEADER_RE = /^##\s+(.*\S)\s*$/;

interface FetchReport {
  status: string;
  detail: string;
}

interface DriftReport {
  checkout: string;
  resolved: boolean;
  pull_failed: boolean;
  pinned_version: string | null;
  latest_version: string | null;
  behind: boolean;
  commit_count: number;
  commits: readonly string[];
  review_refs: readonly string[];
  note?: string;
}

interface SourceReport {
  title: string;
  url: string | undefined;
  urls: readonly string[];
  provenance: string | undefined;
  last_reviewed: string | null;
  age_days: number | null;
  stale: boolean;
  refs: readonly string[];
  covers_all_references: boolean;
  reference_mapping_missing: boolean;
  dangling_refs: readonly string[];
  version: string | null;
  commit: string | null;
  watch_count: number;
  fetch?: FetchReport;
  fetches?: readonly FetchReport[];
  drift?: DriftReport;
}

interface SkillReport {
  skill: string;
  sources_file: string;
  source_count: number;
  sources: readonly SourceReport[];
  uncovered_refs: readonly string[];
  problems: number;
}

const HELP = `Audit a skill's upstream sources (SOURCES.md) for drift.

Usage:
    moon-skill-audit-sources <skill-dir> [options]
    moon-skill-audit-sources --all [root] [options]

      skill-dir            path to a skill dir (containing SOURCES.md) or to a
                           SOURCES.md file directly
      --all [root]         audit every */SOURCES.md under root (default: cwd)
      --max-age DAYS       staleness threshold in days (default: 180)
      --fetch              HTTP-check each URL still resolves
      --pull               fetch tags in each source's **Checkout:** repo first
      --mark-reviewed [T]  stamp 'Last reviewed' to today for sources whose title
                           contains T (case-insensitive); omit T to stamp all
      --json               emit a JSON report instead of text
      -h, --help           show this help

A source block may add upstream-drift fields:
      **References:** all | a.md, b.md
                                      reference files this source supports
      **Checkout:** <path>            local source repo (relative to workspace root)
      **Version:** <semver>           version the skill is pinned to
      **Commit:** <hash>              commit the skill was distilled from
      **Watch:** <subpath> -> a.md    map a source subpath to the references it feeds
When present, the audit also reports the latest released tag vs the pinned
version and the commits since the pinned commit on watched paths.`;

// ensureAscii escapes every non-ASCII UTF-16 code unit as \uXXXX.
const ensureAscii = (text: string): string =>
  text.replaceAll(/[\u0080-\uFFFF]/g, (ch) => {
    // The non-Unicode regex visits each half of a surrogate pair separately.
    const codePoint = ch.codePointAt(0);
    if (codePoint === undefined) return ch;
    const code = codePoint.toString(16).padStart(4, "0");
    return String.raw`\u${code}`;
  });

const toIsoDate = (date: Date): string => {
  const year = date.getFullYear().toString().padStart(4, "0");
  const month = (date.getMonth() + 1).toString().padStart(2, "0");
  const day = date.getDate().toString().padStart(2, "0");
  return `${year}-${month}-${day}`;
};

const daysBetween = (today: Date, reviewed: Date): number => {
  const msPerDay = 24 * 60 * 60 * 1000;
  const a = Date.UTC(today.getFullYear(), today.getMonth(), today.getDate());
  const b = Date.UTC(
    reviewed.getFullYear(),
    reviewed.getMonth(),
    reviewed.getDate(),
  );
  return Math.round((a - b) / msPerDay);
};

const resolveSourcesFile = (target: string): string | undefined => {
  if (isFile(target) && basename(target) === "SOURCES.md") {
    return target;
  }
  if (isDirectory(target)) {
    if (isFile(join(target, "SOURCES.md"))) {
      return join(target, "SOURCES.md");
    }
    const guideDir = resolveGuideDirectory(target);
    if (isFile(join(guideDir, "SOURCES.md"))) {
      return join(guideDir, "SOURCES.md");
    }
  }
  return undefined;
};

const fetchStatus = async (url: string, timeout = 15): Promise<FetchReport> => {
  const controller = new AbortController();
  const timer = setTimeout(() => {
    controller.abort();
  }, timeout * 1000);
  try {
    const resp = await fetch(url, {
      headers: {"User-Agent": "skill-source-audit/1"},
      signal: controller.signal,
    });
    const code = resp.status;
    const lm = resp.headers.get("Last-Modified") ?? "";
    const detail =
      `HTTP ${String(code)}` + (lm ? `, Last-Modified: ${lm}` : "");
    if (code === 404) {
      return {status: "missing", detail: `HTTP ${String(code)}`};
    }
    return {status: code && code < 400 ? "ok" : "error", detail};
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    return {status: "error", detail: `unreachable: ${message}`};
  } finally {
    clearTimeout(timer);
  }
};

const listExistingRefs = (skillDir: string): Set<string> => {
  const refsDir = join(skillDir, "references");
  if (!isDirectory(refsDir)) return new Set<string>();
  const refs = new Set<string>();
  for (const entry of readdirSync(refsDir)) {
    if (entry.endsWith(".md") && isFile(join(refsDir, entry))) {
      refs.add(entry);
    }
  }
  return refs;
};

// findWorkspaceRoot resolves checkout paths from the nearest Moon workspace.
const findWorkspaceRoot = (start: string): string => {
  let dir = resolve(start);
  for (;;) {
    if (isDirectory(join(dir, ".moon"))) return dir;
    const parent = resolve(dir, "..");
    if (parent === dir) return resolve(start);
    dir = parent;
  }
};

const git = (cwd: string, args: readonly string[]): string | undefined => {
  try {
    return execFileSync(resolveExecutable("git"), [...args], {
      cwd,
      encoding: "utf8",
      stdio: ["ignore", "pipe", "ignore"],
      maxBuffer: 10 * 1024 * 1024,
    });
  } catch {
    return undefined;
  }
};

type Semver = readonly [number, number, number];

const parseSemver = (text: string): Semver | undefined => {
  const m = /(\d+)\.(\d+)\.(\d+)/.exec(text);
  return m?.[1] !== undefined && m[2] !== undefined && m[3] !== undefined
    ? [Number(m[1]), Number(m[2]), Number(m[3])]
    : undefined;
};

const cmpSemver = (a: Semver, b: Semver): number =>
  a[0] - b[0] || a[1] - b[1] || a[2] - b[2];

const latestSemverTag = (
  checkout: string,
): {version: Semver; raw: string} | undefined => {
  const out = git(checkout, ["tag", "--list"]);
  if (out === undefined) return undefined;
  let best: {version: Semver; raw: string} | undefined;
  for (const line of out.split("\n")) {
    const raw = line.trim();
    if (!/^v?\d+\.\d+\.\d+$/.test(raw)) continue;
    const version = parseSemver(raw);
    if (version && (!best || cmpSemver(version, best.version) > 0)) {
      best = {version, raw};
    }
  }
  return best;
};

const computeDrift = (
  workspaceRoot: string,
  source: Source,
  pull: boolean,
): DriftReport => {
  const checkout = resolve(workspaceRoot, source.checkout ?? "");
  const report: DriftReport = {
    checkout,
    resolved: false,
    pull_failed: false,
    pinned_version: source.version ?? null,
    latest_version: null,
    behind: false,
    commit_count: 0,
    commits: [],
    review_refs: [],
  };
  if (!isDirectory(checkout) || !isDirectory(join(checkout, ".git"))) {
    return {...report, note: "checkout not found or not a git repo"};
  }
  report.resolved = true;
  if (pull && git(checkout, ["fetch", "--tags", "--quiet"]) === undefined) {
    return {
      ...report,
      pull_failed: true,
      note: "failed to fetch upstream tags",
    };
  }

  const latest = latestSemverTag(checkout);
  const pinnedV = source.version ? parseSemver(source.version) : undefined;
  if (latest) {
    report.latest_version = latest.raw;
    if (pinnedV) report.behind = cmpSemver(latest.version, pinnedV) > 0;
  }

  // Diff base: the pinned commit (exact) else the pinned version tag.
  const ref = source.commit ?? source.version;
  if (ref === undefined) {
    return {...report, note: "no **Commit:** or **Version:** to diff from"};
  }
  if (
    git(checkout, ["rev-parse", "--verify", "--quiet", `${ref}^{commit}`]) ===
    undefined
  ) {
    return {...report, note: `pinned ref '${ref}' not found in checkout`};
  }

  // Upper bound: the latest released tag (deterministic after --pull) else HEAD.
  const upper = latest ? latest.raw : "HEAD";
  const paths = source.watches.map((w) => w.path);
  const range = `${ref}..${upper}`;

  const logOut =
    git(checkout, [
      "log",
      "--oneline",
      range,
      ...(paths.length > 0 ? ["--", ...paths] : []),
    ]) ?? "";
  const commits = logOut.split("\n").filter((l) => l.trim());
  report.commit_count = commits.length;
  report.commits = commits.slice(0, 20);

  const diffOut =
    git(checkout, [
      "diff",
      "--name-only",
      range,
      ...(paths.length > 0 ? ["--", ...paths] : []),
    ]) ?? "";
  const reviewRefs = new Set<string>();
  for (const file of diffOut.split("\n")) {
    const f = file.trim();
    if (!f) continue;
    for (const w of source.watches) {
      if (f.startsWith(w.path)) for (const r of w.refs) reviewRefs.add(r);
    }
  }
  report.review_refs = [...reviewRefs].toSorted();
  return report;
};

const auditSkill = async (
  sourcesFile: string,
  maxAge: number,
  doFetch: boolean,
  pull: boolean,
  today: Date,
): Promise<SkillReport> => {
  const skillDir = resolve(sourcesFile, "..");
  const workspaceRoot = findWorkspaceRoot(skillDir);
  const text = readFileSync(sourcesFile, "utf8");
  const sources = parseSources(text);

  const existingRefs = listExistingRefs(skillDir);
  const covered = new Set<string>();

  const srcReports: SourceReport[] = [];
  let problems = 0;
  for (const s of sources) {
    const age = s.reviewed ? daysBetween(today, s.reviewed) : null;
    const stale = age !== null && age > maxAge;
    const referenceMappingMissing = !hasReferenceMapping(s);
    const dangling = [...s.refs].filter((r) => !existingRefs.has(r)).toSorted();
    for (const r of s.refs) covered.add(r);
    const report: SourceReport = {
      title: s.title,
      url: s.url,
      urls: s.urls,
      provenance: s.provenance,
      last_reviewed: s.reviewedRaw ?? null,
      age_days: age,
      stale,
      refs: [...s.refs].toSorted(),
      covers_all_references: s.coversAllReferences,
      reference_mapping_missing: referenceMappingMissing,
      dangling_refs: dangling,
      version: s.version ?? null,
      commit: s.commit ?? null,
      watch_count: s.watches.length,
    };
    if (s.coversAllReferences) {
      for (const reference of existingRefs) covered.add(reference);
    }
    if (age === null || stale || dangling.length > 0) {
      problems += 1;
    }
    if (referenceMappingMissing) problems += 1;
    if (doFetch && s.urls.length > 0) {
      const fetches = await Promise.all(s.urls.map((url) => fetchStatus(url)));
      report.fetch = fetches[0];
      report.fetches = fetches;
      problems += fetches.filter((fetch) => fetch.status !== "ok").length;
    }
    if (s.checkout !== undefined) {
      const drift = computeDrift(workspaceRoot, s, pull);
      report.drift = drift;
      if (drift.pull_failed || drift.behind || drift.commit_count > 0) {
        problems += 1;
      }
    }
    srcReports.push(report);
  }

  const uncovered = [...existingRefs].filter((r) => !covered.has(r)).toSorted();
  problems += uncovered.length;
  if (sources.length === 0) problems += 1;
  return {
    skill: basename(resolve(skillDir)),
    sources_file: sourcesFile,
    source_count: sources.length,
    sources: srcReports,
    uncovered_refs: uncovered,
    problems,
  };
};

const markReviewed = (
  sourcesFile: string,
  titleFilter: string | undefined,
  today: Date,
): string[] => {
  const text = readFileSync(sourcesFile, "utf8");
  const lines = text.split(/(?<=\n)/);
  const needle = titleFilter ? titleFilter.toLowerCase() : undefined;
  const edited: string[] = [];
  let currentTitle = "";
  let matched = false;
  for (const [i, line] of lines.entries()) {
    const header = HEADER_RE.exec(line.replace(/\n$/, ""));
    if (header) {
      currentTitle = header[1] ?? "";
      matched =
        needle === undefined || currentTitle.toLowerCase().includes(needle);
      continue;
    }
    if (matched && REVIEWED_RE.test(line)) {
      lines[i] = line.replace(
        REVIEWED_RE,
        `**Last reviewed:** ${toIsoDate(today)}`,
      );
      if (currentTitle && !edited.includes(currentTitle)) {
        edited.push(currentTitle);
      }
    }
  }
  if (edited.length > 0) {
    writeFileSync(sourcesFile, lines.join(""), "utf8");
  }
  return edited;
};

const sourceFlags = (source: SourceReport): readonly string[] => {
  const flags: string[] = [];
  if (source.drift?.pull_failed) flags.push("PULL FAILED");
  if (source.stale) flags.push(`STALE (${String(source.age_days)}d)`);
  if (source.last_reviewed === null) flags.push("MISSING REVIEW DATE");
  if (source.reference_mapping_missing) {
    flags.push("MISSING REFERENCE MAPPING");
  }
  if (source.dangling_refs.length > 0) {
    flags.push(`DANGLING: ${source.dangling_refs.join(", ")}`);
  }
  const failedFetch = (
    source.fetches ?? (source.fetch ? [source.fetch] : [])
  ).find((fetch) => fetch.status !== "ok");
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

const printDrift = (source: SourceReport): void => {
  const drift = source.drift;
  if (drift === undefined) return;
  if (!drift.resolved) {
    console.log(`    upstream      : ${drift.note ?? "unresolved"}`);
    return;
  }

  const latest = drift.latest_version ?? "(no tags)";
  const state = drift.behind
    ? `behind — latest released ${latest}`
    : `up to date with latest released ${latest}`;
  console.log(
    `    upstream      : pinned ${drift.pinned_version ?? "?"}, ${state}`,
  );
  if (drift.note) console.log(`                    note: ${drift.note}`);
  if (drift.commit_count === 0) return;

  console.log(
    `    commits since : ${String(drift.commit_count)} on watched paths`,
  );
  for (const commit of drift.commits)
    console.log(`                    ${commit}`);
  const review = drift.review_refs.length > 0 ? drift.review_refs : source.refs;
  if (review.length > 0) {
    const paths = review.map((ref) => "references/" + ref).join(", ");
    console.log(`    review        : ${paths}`);
  }
};

const sourceFeeds = (source: SourceReport): string => {
  if (source.covers_all_references) return "all references";
  if (source.refs.length === 0) return "(missing)";
  return source.refs.map((ref) => `references/${ref}`).join(", ");
};

const printSource = (source: SourceReport): void => {
  const flags = sourceFlags(source);
  const marker = flags.length > 0 ? flags.join("  ") : "ok";
  const fetches = source.fetches ?? (source.fetch ? [source.fetch] : []);
  console.log(`\n  [${marker}] ${source.title}`);
  console.log(
    `    urls          : ${source.urls.length > 0 ? source.urls.join(", ") : "(none)"}`,
  );
  if (source.provenance) {
    console.log(`    provenance    : ${source.provenance}`);
  }
  const age =
    source.age_days === null ? "" : `  (${String(source.age_days)}d ago)`;
  console.log(`    last reviewed : ${source.last_reviewed ?? "(none)"}${age}`);
  for (const [index, fetch] of fetches.entries()) {
    console.log(`    fetch ${String(index + 1).padEnd(7)}: ${fetch.detail}`);
  }
  console.log(`    feeds         : ${sourceFeeds(source)}`);
  printDrift(source);
};

const printTextReport = (report: SkillReport, maxAge: number): void => {
  console.log(`skill: ${report.skill}  (${report.sources_file})`);
  console.log(
    `sources: ${String(report.source_count)}  max-age: ${String(maxAge)}d`,
  );
  for (const source of report.sources) printSource(source);
  if (report.uncovered_refs.length > 0) {
    console.log(
      "\n  reference files with no declared source or repository-original provenance:",
    );
    for (const ref of report.uncovered_refs) {
      console.log(`    references/${ref}`);
    }
  }
  console.log(
    "\nNote: distilled prose is not auto-rewritten. Review the reference files a " +
      "stale/changed source feeds, then re-run with --mark-reviewed to stamp the date.",
  );
};

const collectTargets = (args: ParsedArgs): string[] => {
  if (args.all !== undefined) {
    const root = args.all;
    if (!isDirectory(root)) {
      process.stderr.write(`error: --all root not a directory: ${root}\n`);
      process.exit(2);
    }
    const found = walkSourcesFiles(root).toSorted(comparePaths);
    if (found.length === 0) {
      process.stderr.write(`error: no SOURCES.md found under ${root}\n`);
      process.exit(2);
    }
    return found;
  }
  // collectTargets uses the current directory when no target is supplied.
  const target = args.target ?? process.cwd();
  const sf = resolveSourcesFile(target);
  if (sf === undefined) {
    process.stderr.write(`error: no SOURCES.md at ${target}\n`);
    process.exit(2);
  }
  return [sf];
};

// comparePaths orders path components before comparing path depth.
const comparePaths = (a: string, b: string): number => {
  const partsA = a.split(sep);
  const partsB = b.split(sep);
  const len = Math.min(partsA.length, partsB.length);
  for (let i = 0; i < len; i++) {
    const ca = partsA[i] ?? "";
    const cb = partsB[i] ?? "";
    if (ca < cb) return -1;
    if (ca > cb) return 1;
  }
  return partsA.length - partsB.length;
};

const walkSourcesFiles = (root: string): string[] => {
  const found: string[] = [];
  const walk = (dir: string): void => {
    let entries: string[];
    try {
      entries = readdirSync(dir);
    } catch {
      return;
    }
    for (const entry of entries) {
      const full = join(dir, entry);
      if (isDirectory(full)) {
        if (entry === "node_modules") continue;
        walk(full);
      } else if (entry === "SOURCES.md" && isFile(join(dir, "SKILL.md"))) {
        found.push(full);
      }
    }
  };
  walk(root);
  return found;
};

export const main = async (argv: readonly string[]): Promise<number> => {
  const args = parseArgs(argv);
  if (args.help) {
    console.log(HELP);
    return 0;
  }
  const today = new Date();
  const targets = collectTargets(args);

  // --mark-reviewed is a single-skill write action; refuse it for --all.
  if (args.markReviewed !== undefined) {
    const single = targets[0];
    if (single === undefined || targets.length !== 1) {
      process.stderr.write(
        "error: --mark-reviewed operates on a single skill\n",
      );
      return 2;
    }
    const edited = markReviewed(single, args.markReviewed || undefined, today);
    if (edited.length > 0) {
      process.stderr.write(
        `stamped 'Last reviewed' = ${toIsoDate(today)} for: ` +
          edited.join(", ") +
          "\n",
      );
    } else {
      process.stderr.write("no matching source blocks to stamp\n");
    }
    return 0;
  }

  const reports: SkillReport[] = [];
  for (const sf of targets) {
    reports.push(
      await auditSkill(sf, args.maxAge, args.fetch, args.pull, today),
    );
  }
  const totalProblems = reports.reduce((sum, r) => sum + r.problems, 0);

  if (args.json) {
    const out = reports.length === 1 ? reports[0] : reports;
    console.log(ensureAscii(JSON.stringify(out, null, 2)));
  } else {
    for (const [i, rep] of reports.entries()) {
      if (i) {
        console.log("\n" + "=".repeat(60));
      }
      printTextReport(rep, args.maxAge);
    }
    if (reports.length > 1) {
      const stale = reports.reduce(
        (sum, r) => sum + r.sources.filter((s) => s.stale).length,
        0,
      );
      process.stderr.write(
        `\n--- ${String(reports.length)} skills, ${String(totalProblems)} problem(s), ` +
          `${String(stale)} stale source(s) ---\n`,
      );
    }
  }

  return totalProblems ? 1 : 0;
};
