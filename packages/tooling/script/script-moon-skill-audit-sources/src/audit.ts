import {resolve} from "node:path";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import type {Output} from "@xonovex/script-moon-common/output";
import {parseArgs, type ParsedArgs} from "./args.js";
import {toIsoDate} from "./dates.js";
import type {HttpClient} from "./http.js";
import {stampReviewed} from "./mark-reviewed.js";
import {findSourcesFiles, resolveSourcesFile} from "./skill-sources.js";
import {auditSkill, type SkillReport} from "./source-audit.js";
import {renderJsonReport, renderTextReport} from "./text-report.js";
import type {UpstreamRepository} from "./upstream.js";

const HELP = `Audit a skill's upstream sources (SOURCES.md) for drift.

Usage:
    moon-skill-audit-sources <skill-dir> [options]
    moon-skill-audit-sources --all [root] [options]

      skill-dir            path to a skill dir (containing SOURCES.md) or to a
                           SOURCES.md file directly
      --all [root]         audit every */SOURCES.md under root (default: cwd)
      --max-age DAYS       staleness threshold in days (default: 180)
      --version-max-age D  shorter threshold for versioned sources (default: 90)
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
      **Content SHA256:** <digest>    pin ordered fetched web content when no checkout exists
When present, the audit also reports the latest released tag vs the pinned
version and the commits since the pinned commit on watched paths.`;

export interface AuditSourcesDependencies {
  readonly fs: FileSystem;
  readonly output: Output;
  readonly http: HttpClient;
  readonly openUpstream: (directory: string) => UpstreamRepository;
  readonly today: Date;
}

type Targets =
  | {readonly kind: "found"; readonly files: readonly string[]}
  | {readonly kind: "missing"; readonly message: string};

// Without a target the audit reads the working directory.
const collectTargets = (
  args: ParsedArgs,
  cwd: string,
  fs: FileSystem,
): Targets => {
  if (args.all !== undefined) {
    const root = args.all;
    if (!fs.isDirectory(resolve(cwd, root))) {
      return {kind: "missing", message: `--all root not a directory: ${root}`};
    }
    const files = findSourcesFiles(root, cwd, fs);
    return files.length === 0
      ? {kind: "missing", message: `no SOURCES.md found under ${root}`}
      : {kind: "found", files};
  }
  const target = args.target ?? cwd;
  const file = resolveSourcesFile(target, cwd, fs);
  return file === undefined
    ? {kind: "missing", message: `no SOURCES.md at ${target}`}
    : {kind: "found", files: [file]};
};

// --mark-reviewed is a single-skill write action, refused for --all.
const markReviewed = (
  targets: readonly string[],
  titleFilter: string,
  cwd: string,
  dependencies: AuditSourcesDependencies,
): number => {
  const {fs, output, today} = dependencies;
  const single = targets[0];
  if (single === undefined || targets.length !== 1) {
    output.stderr("error: --mark-reviewed operates on a single skill");
    return 2;
  }
  const path = resolve(cwd, single);
  const stamped = stampReviewed(
    fs.readText(path),
    titleFilter || undefined,
    today,
  );
  if (stamped.edited.length === 0) {
    output.stderr("no matching source blocks to stamp");
    return 0;
  }
  fs.writeFile(path, stamped.text);
  output.stderr(
    `stamped 'Last reviewed' = ${toIsoDate(today)} for: ` +
      stamped.edited.join(", "),
  );
  return 0;
};

const renderReports = (
  reports: readonly SkillReport[],
  args: ParsedArgs,
  output: Output,
): void => {
  if (args.json) {
    output.stdout(renderJsonReport(reports));
    return;
  }
  for (const [i, report] of reports.entries()) {
    if (i) output.stdout("\n" + "=".repeat(60));
    for (const line of renderTextReport(
      report,
      args.maxAge,
      args.versionMaxAge,
    )) {
      output.stdout(line);
    }
  }
  if (reports.length > 1) {
    const problems = reports.reduce((sum, r) => sum + r.problems, 0);
    const stale = reports.reduce(
      (sum, r) => sum + r.sources.filter((s) => s.stale).length,
      0,
    );
    output.stderr(
      `\n--- ${String(reports.length)} skills, ${String(problems)} problem(s), ` +
        `${String(stale)} stale source(s) ---`,
    );
  }
};

/** Audits the sources of one skill or of every skill below a root and returns the exit code. */
export const main = async (
  argv: readonly string[],
  cwd: string,
  dependencies: AuditSourcesDependencies,
): Promise<number> => {
  const {fs, output} = dependencies;
  let args: ParsedArgs;
  try {
    args = parseArgs(argv);
  } catch (error) {
    output.stderr(
      `error: ${error instanceof Error ? error.message : String(error)}`,
    );
    return 2;
  }
  if (args.help) {
    output.stdout(HELP);
    return 0;
  }
  const targets = collectTargets(args, cwd, fs);
  if (targets.kind === "missing") {
    output.stderr(`error: ${targets.message}`);
    return 2;
  }
  if (args.markReviewed !== undefined) {
    return markReviewed(targets.files, args.markReviewed, cwd, dependencies);
  }

  const options = {
    maxAge: args.maxAge,
    versionMaxAge: args.versionMaxAge,
    fetch: args.fetch,
    pull: args.pull,
  };
  const reports: SkillReport[] = [];
  for (const file of targets.files) {
    reports.push(await auditSkill(file, cwd, options, dependencies));
  }
  renderReports(reports, args, output);
  return reports.some((report) => report.problems > 0) ? 1 : 0;
};
