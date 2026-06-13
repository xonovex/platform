#!/usr/bin/env node
import {readdirSync, readFileSync, statSync, writeFileSync} from "node:fs";
import {basename, join, resolve, sep} from "node:path";

const URL_RE = /\*\*URL:\*\*\s*(\S+)/;
const REVIEWED_RE = /\*\*Last reviewed:\*\*\s*(\d{4}-\d{2}-\d{2})/;
const REF_RE = /references\/([a-z0-9][a-z0-9-]*\.md)/g;
const HEADER_RE = /^##\s+(.*\S)\s*$/;

interface Source {
  title: string;
  url: string | undefined;
  reviewed: Date | undefined;
  reviewedRaw: string | undefined;
  refs: Set<string>;
  lineNo: number;
}

interface FetchReport {
  status: string;
  detail: string;
}

interface SourceReport {
  title: string;
  url: string | undefined;
  last_reviewed: string | null;
  age_days: number | null;
  stale: boolean;
  refs: readonly string[];
  dangling_refs: readonly string[];
  fetch?: FetchReport;
}

interface SkillReport {
  skill: string;
  sources_file: string;
  source_count: number;
  sources: readonly SourceReport[];
  uncovered_refs: readonly string[];
  problems: number;
}

interface ParsedArgs {
  target: string | undefined;
  all: string | undefined;
  maxAge: number;
  fetch: boolean;
  markReviewed: string | undefined;
  json: boolean;
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
      --mark-reviewed [T]  stamp 'Last reviewed' to today for sources whose title
                           contains T (case-insensitive); omit T to stamp all
      --json               emit a JSON report instead of text
      -h, --help           show this help`;

// Match Python json.dumps(ensure_ascii=True): escape every non-ASCII UTF-16
// code unit as \uXXXX (surrogate pairs become two escapes, as in CPython).
const ensureAscii = (text: string): string =>
  text.replaceAll(/[\u0080-\uFFFF]/g, (ch) => {
    // charCodeAt (UTF-16 code unit), not codePointAt: the regex matches one code
    // unit at a time, so astral chars must emit two \uXXXX surrogate escapes to
    // match CPython, which codePointAt's combined code point would not produce.
    // eslint-disable-next-line unicorn/prefer-code-point
    const code = ch.charCodeAt(0).toString(16).padStart(4, "0");
    return String.raw`\u${code}`;
  });

const toIsoDate = (date: Date): string => {
  const year = date.getFullYear().toString().padStart(4, "0");
  const month = (date.getMonth() + 1).toString().padStart(2, "0");
  const day = date.getDate().toString().padStart(2, "0");
  return `${year}-${month}-${day}`;
};

const parseIsoDate = (raw: string): Date | undefined => {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(raw);
  if (m?.[1] === undefined || m[2] === undefined || m[3] === undefined) {
    return undefined;
  }
  const year = Number(m[1]);
  const month = Number(m[2]);
  const day = Number(m[3]);
  const date = new Date(year, month - 1, day);
  if (
    date.getFullYear() !== year ||
    date.getMonth() !== month - 1 ||
    date.getDate() !== day
  ) {
    return undefined;
  }
  return date;
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

const findRefs = (line: string): string[] => {
  const found: string[] = [];
  REF_RE.lastIndex = 0;
  let m: RegExpExecArray | null;
  while ((m = REF_RE.exec(line)) !== null) {
    if (m[1] !== undefined) found.push(m[1]);
  }
  return found;
};

const parseSources = (text: string): Source[] => {
  const sources: Source[] = [];
  let current: Source | undefined;
  const lines = text.split("\n");
  for (const [i, line] of lines.entries()) {
    const header = HEADER_RE.exec(line);
    if (header) {
      current = {
        title: header[1] ?? "",
        url: undefined,
        reviewed: undefined,
        reviewedRaw: undefined,
        refs: new Set<string>(),
        lineNo: i,
      };
      sources.push(current);
      continue;
    }
    if (current === undefined) continue;
    const urlMatch = URL_RE.exec(line);
    if (urlMatch?.[1] !== undefined) {
      current.url = urlMatch[1];
    }
    const reviewedMatch = REVIEWED_RE.exec(line);
    if (reviewedMatch?.[1] !== undefined) {
      current.reviewedRaw = reviewedMatch[1];
      current.reviewed = parseIsoDate(reviewedMatch[1]);
    }
    for (const ref of findRefs(line)) {
      current.refs.add(ref);
    }
  }
  // A block is a real source only if it carries a URL (others are prose notes).
  return sources.filter((s) => s.url);
};

const isFile = (target: string): boolean => {
  try {
    return statSync(target).isFile();
  } catch {
    return false;
  }
};

const isDir = (target: string): boolean => {
  try {
    return statSync(target).isDirectory();
  } catch {
    return false;
  }
};

// The skill/guide dir for a base dir: base itself if base/SKILL.md exists,
// else the single immediate subdir containing SKILL.md (skill-package layout,
// e.g. skill-c99/c99-guide/SKILL.md). >1 match is ambiguous; 0 falls back to
// base so the existing not-found error still fires.
const resolveGuideDir = (base: string): string => {
  if (isFile(join(base, "SKILL.md"))) {
    return base;
  }
  const nested = readdirSync(base)
    .map((entry) => join(base, entry))
    .filter((dir) => isFile(join(dir, "SKILL.md")));
  if (nested.length > 1) {
    process.stderr.write(
      `error: multiple SKILL.md found under ${base}; pass one explicitly\n`,
    );
    process.exit(2);
  }
  return nested[0] ?? base;
};

const resolveSourcesFile = (target: string): string | undefined => {
  if (isFile(target) && basename(target) === "SOURCES.md") {
    return target;
  }
  if (isDir(target)) {
    if (isFile(join(target, "SOURCES.md"))) {
      return join(target, "SOURCES.md");
    }
    // No SOURCES.md directly: descend into the single guide subdir (identified
    // by SKILL.md, per the skill-package layout) and use its SOURCES.md.
    const guideDir = resolveGuideDir(target);
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
      method: "GET",
      headers: {"User-Agent": "skill-source-audit/1"},
      redirect: "follow",
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
  if (!isDir(refsDir)) return new Set<string>();
  const refs = new Set<string>();
  for (const entry of readdirSync(refsDir)) {
    if (entry.endsWith(".md") && isFile(join(refsDir, entry))) {
      refs.add(entry);
    }
  }
  return refs;
};

const auditSkill = async (
  sourcesFile: string,
  maxAge: number,
  doFetch: boolean,
  today: Date,
): Promise<SkillReport> => {
  const skillDir = resolve(sourcesFile, "..");
  const text = readFileSync(sourcesFile, "utf8");
  const sources = parseSources(text);

  const existingRefs = listExistingRefs(skillDir);
  const covered = new Set<string>();

  const srcReports: SourceReport[] = [];
  let problems = 0;
  for (const s of sources) {
    const age = s.reviewed ? daysBetween(today, s.reviewed) : null;
    const stale = age !== null && age > maxAge;
    const dangling = [...s.refs].filter((r) => !existingRefs.has(r)).toSorted();
    for (const r of s.refs) covered.add(r);
    const report: SourceReport = {
      title: s.title,
      url: s.url,
      last_reviewed: s.reviewedRaw ?? null,
      age_days: age,
      stale,
      refs: [...s.refs].toSorted(),
      dangling_refs: dangling,
    };
    if (stale || dangling.length > 0) {
      problems += 1;
    }
    if (doFetch && s.url) {
      const fetchReport = await fetchStatus(s.url);
      report.fetch = fetchReport;
      if (fetchReport.status !== "ok") {
        problems += 1;
      }
    }
    srcReports.push(report);
  }

  const uncovered = [...existingRefs].filter((r) => !covered.has(r)).toSorted();
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

const printTextReport = (rep: SkillReport, maxAge: number): void => {
  console.log(`skill: ${rep.skill}  (${rep.sources_file})`);
  console.log(
    `sources: ${String(rep.source_count)}  max-age: ${String(maxAge)}d`,
  );
  for (const s of rep.sources) {
    const flags: string[] = [];
    if (s.stale) {
      flags.push(`STALE (${String(s.age_days)}d)`);
    }
    if (s.dangling_refs.length > 0) {
      flags.push(`DANGLING: ${s.dangling_refs.join(", ")}`);
    }
    const fetch = s.fetch;
    if (fetch && fetch.status !== "ok") {
      flags.push(`URL ${fetch.status}`);
    }
    const marker = flags.length > 0 ? flags.join("  ") : "ok";
    console.log(`\n  [${marker}] ${s.title}`);
    console.log(`    url           : ${String(s.url)}`);
    console.log(
      `    last reviewed : ${s.last_reviewed ?? "(none)"}` +
        (s.age_days === null ? "" : `  (${String(s.age_days)}d ago)`),
    );
    if (fetch) {
      console.log(`    fetch         : ${fetch.detail}`);
    }
    if (s.refs.length > 0) {
      console.log(
        `    feeds         : ${s.refs.map((r) => "references/" + r).join(", ")}`,
      );
    }
  }
  if (rep.uncovered_refs.length > 0) {
    console.log(
      "\n  reference files with no upstream source (repo-original or provenance gap):",
    );
    for (const r of rep.uncovered_refs) {
      console.log(`    references/${r}`);
    }
  }
  console.log(
    "\nNote: distilled prose is not auto-rewritten. Review the reference files a " +
      "stale/changed source feeds, then re-run with --mark-reviewed to stamp the date.",
  );
};

const isOptionToken = (token: string | undefined): boolean =>
  token !== undefined && token.startsWith("-") && token !== "-";

const parseArgv = (argv: readonly string[]): ParsedArgs => {
  const parsed: ParsedArgs = {
    target: undefined,
    all: undefined,
    maxAge: 180,
    fetch: false,
    markReviewed: undefined,
    json: false,
  };
  const positionals: string[] = [];

  // Manual scan to reproduce argparse semantics, including the optional-value
  // flags --all [ROOT] and --mark-reviewed [TITLE] that node:util cannot model.
  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i];
    if (arg === undefined) continue;
    switch (arg) {
      case "-h":
      case "--help": {
        console.log(HELP);
        process.exit(0);
      }
      case "--fetch": {
        parsed.fetch = true;
        break;
      }
      case "--json": {
        parsed.json = true;
        break;
      }
      default: {
        if (arg === "--max-age" || arg.startsWith("--max-age=")) {
          let raw: string | undefined;
          if (arg.startsWith("--max-age=")) {
            raw = arg.slice("--max-age=".length);
          } else {
            raw = argv[i + 1];
            i += 1;
          }
          if (raw === undefined || isOptionToken(raw)) {
            process.stderr.write(
              "error: argument --max-age: expected one argument\n",
            );
            process.exit(2);
          }
          const value = Number(raw);
          if (!Number.isInteger(value)) {
            process.stderr.write(
              `error: argument --max-age: invalid int value: '${raw}'\n`,
            );
            process.exit(2);
          }
          parsed.maxAge = value;
        } else if (arg === "--all" || arg.startsWith("--all=")) {
          if (arg.startsWith("--all=")) {
            parsed.all = arg.slice("--all=".length);
          } else {
            const next = argv[i + 1];
            if (next !== undefined && !isOptionToken(next)) {
              parsed.all = next;
              i += 1;
            } else {
              parsed.all = ".";
            }
          }
        } else if (
          arg === "--mark-reviewed" ||
          arg.startsWith("--mark-reviewed=")
        ) {
          if (arg.startsWith("--mark-reviewed=")) {
            parsed.markReviewed = arg.slice("--mark-reviewed=".length);
          } else {
            const next = argv[i + 1];
            if (next !== undefined && !isOptionToken(next)) {
              parsed.markReviewed = next;
              i += 1;
            } else {
              parsed.markReviewed = "";
            }
          }
        } else if (isOptionToken(arg)) {
          process.stderr.write(`error: unrecognized arguments: ${arg}\n`);
          process.exit(2);
        } else {
          positionals.push(arg);
        }
      }
    }
  }

  if (positionals.length > 1) {
    process.stderr.write(
      `error: unrecognized arguments: ${positionals.slice(1).join(" ")}\n`,
    );
    process.exit(2);
  }
  parsed.target = positionals[0];
  return parsed;
};

const collectTargets = (args: ParsedArgs): string[] => {
  if (args.all !== undefined) {
    const root = args.all;
    if (!isDir(root)) {
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
  // Ergonomic addition: default the target to cwd when neither a target nor
  // --all is supplied (the Python CLI errors here instead).
  const target = args.target ?? process.cwd();
  const sf = resolveSourcesFile(target);
  if (sf === undefined) {
    process.stderr.write(`error: no SOURCES.md at ${target}\n`);
    process.exit(2);
  }
  return [sf];
};

// Mirror Python's pathlib.Path ordering, which compares path components one by
// one (so "skill-c99" sorts before "skill-c99-opinionated") rather than the raw
// string (where the path separator would flip that order).
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
      if (isDir(full)) {
        if (entry === "node_modules") continue;
        walk(full);
      } else if (entry === "SOURCES.md") {
        found.push(full);
      }
    }
  };
  walk(root);
  return found;
};

const main = async (argv: readonly string[]): Promise<number> => {
  const args = parseArgv(argv);
  const today = new Date();
  const targets = collectTargets(args);

  // --mark-reviewed is a single-skill write action; refuse it for --all.
  if (args.markReviewed !== undefined) {
    const single = targets[0];
    if (targets.length !== 1 || single === undefined) {
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
    reports.push(await auditSkill(sf, args.maxAge, args.fetch, today));
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

try {
  const code = await main(process.argv.slice(2));
  process.exit(code);
} catch (error) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`error: ${message}\n`);
  process.exit(2);
}
