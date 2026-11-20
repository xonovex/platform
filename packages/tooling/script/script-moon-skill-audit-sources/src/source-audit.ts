import {basename, join, resolve} from "node:path";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {daysBetween} from "./dates.js";
import {
  aggregateContentSha256,
  fetchStatus,
  type FetchReport,
} from "./fetch-check.js";
import type {HttpClient} from "./http.js";
import {findWorkspaceRoot, listExistingRefs} from "./skill-sources.js";
import {
  hasReferenceMapping,
  parseSources,
  type Source,
} from "./source-parser.js";
import {computeDrift, type DriftReport} from "./upstream-drift.js";
import type {UpstreamRepository} from "./upstream.js";

export interface SourceReport {
  title: string;
  url: string | undefined;
  urls: readonly string[];
  provenance: string | undefined;
  last_reviewed: string | null;
  age_days: number | null;
  review_max_age_days: number;
  stale: boolean;
  refs: readonly string[];
  covers_all_references: boolean;
  reference_mapping_missing: boolean;
  dangling_refs: readonly string[];
  version: string | null;
  commit: string | null;
  watch_count: number;
  content_sha256: string | null;
  fetched_content_sha256: string | null;
  content_changed: boolean;
  drift_anchor_missing: boolean;
  fetch?: FetchReport;
  fetches?: readonly FetchReport[];
  drift?: DriftReport;
}

export interface SkillReport {
  skill: string;
  sources_file: string;
  source_count: number;
  sources: readonly SourceReport[];
  uncovered_refs: readonly string[];
  problems: number;
}

export interface AuditOptions {
  readonly maxAge: number;
  readonly versionMaxAge: number;
  readonly fetch: boolean;
  readonly pull: boolean;
}

export interface AuditDependencies {
  readonly fs: FileSystem;
  readonly http: HttpClient;
  readonly openUpstream: (directory: string) => UpstreamRepository;
  readonly today: Date;
}

interface AuditContext {
  readonly existingRefs: ReadonlySet<string>;
  readonly workspaceRoot: string;
  readonly options: AuditOptions;
  readonly dependencies: AuditDependencies;
}

const upstreamDrift = (source: Source, context: AuditContext): DriftReport => {
  const {fs, openUpstream} = context.dependencies;
  const checkout = resolve(context.workspaceRoot, source.checkout ?? "");
  const isRepository =
    fs.isDirectory(checkout) && fs.isDirectory(join(checkout, ".git"));
  return computeDrift(
    checkout,
    source,
    context.options.pull,
    isRepository ? openUpstream(checkout) : undefined,
  );
};

// auditSource reports one source entry and counts the problems it carries.
const auditSource = async (
  s: Source,
  context: AuditContext,
): Promise<{readonly report: SourceReport; readonly problems: number}> => {
  const {existingRefs, options} = context;
  const {http, today} = context.dependencies;
  let problems = 0;
  const age = s.reviewed ? daysBetween(today, s.reviewed) : null;
  const reviewMaxAge =
    s.version === undefined ? options.maxAge : options.versionMaxAge;
  const stale = age !== null && age > reviewMaxAge;
  const referenceMappingMissing = !hasReferenceMapping(s);
  const driftAnchorMissing =
    s.version !== undefined &&
    s.urls.length > 0 &&
    s.contentSha256 === undefined &&
    !(
      s.checkout !== undefined &&
      s.commit !== undefined &&
      s.watches.length > 0
    );
  const dangling = [...s.refs].filter((r) => !existingRefs.has(r)).toSorted();
  const report: SourceReport = {
    title: s.title,
    url: s.url,
    urls: s.urls,
    provenance: s.provenance,
    last_reviewed: s.reviewedRaw ?? null,
    age_days: age,
    review_max_age_days: reviewMaxAge,
    stale,
    refs: [...s.refs].toSorted(),
    covers_all_references: s.coversAllReferences,
    reference_mapping_missing: referenceMappingMissing,
    dangling_refs: dangling,
    version: s.version ?? null,
    commit: s.commit ?? null,
    watch_count: s.watches.length,
    content_sha256: s.contentSha256 ?? null,
    fetched_content_sha256: null,
    content_changed: false,
    drift_anchor_missing: driftAnchorMissing,
  };
  if (age === null || stale || dangling.length > 0) {
    problems += 1;
  }
  if (referenceMappingMissing) problems += 1;
  if (driftAnchorMissing) problems += 1;
  if (options.fetch && s.urls.length > 0) {
    const fetches = await Promise.all(
      s.urls.map((url) => fetchStatus(http, url)),
    );
    report.fetch = fetches[0];
    report.fetches = fetches;
    problems += fetches.filter(
      (fetch) => fetch.status !== "ok" && fetch.status !== "restricted",
    ).length;
    report.fetched_content_sha256 = aggregateContentSha256(s.urls, fetches);
    if (s.contentSha256 !== undefined) {
      if (report.fetched_content_sha256 === null) {
        problems += 1;
      } else if (report.fetched_content_sha256 !== s.contentSha256) {
        report.content_changed = true;
        problems += 1;
      }
    }
  }
  if (s.checkout !== undefined) {
    const drift = upstreamDrift(s, context);
    report.drift = drift;
    if (drift.pull_failed || drift.behind || drift.commit_count > 0) {
      problems += 1;
    }
  }
  return {problems, report};
};

/** Audits the sources of one skill; sourcesFile is reported as given and read against cwd. */
export const auditSkill = async (
  sourcesFile: string,
  cwd: string,
  options: AuditOptions,
  dependencies: AuditDependencies,
): Promise<SkillReport> => {
  const {fs} = dependencies;
  const skillDir = resolve(cwd, sourcesFile, "..");
  const sources = parseSources(fs.readText(resolve(cwd, sourcesFile)));
  const existingRefs = listExistingRefs(skillDir, fs);
  const context: AuditContext = {
    existingRefs,
    workspaceRoot: findWorkspaceRoot(skillDir, fs),
    options,
    dependencies,
  };

  const covered = new Set<string>();
  const srcReports: SourceReport[] = [];
  let problems = 0;
  for (const s of sources) {
    const audited = await auditSource(s, context);
    problems += audited.problems;
    for (const r of s.refs) covered.add(r);
    if (s.coversAllReferences) {
      for (const reference of existingRefs) covered.add(reference);
    }
    srcReports.push(audited.report);
  }

  const uncovered = [...existingRefs].filter((r) => !covered.has(r)).toSorted();
  problems += uncovered.length;
  if (sources.length === 0) problems += 1;
  return {
    skill: basename(skillDir),
    sources_file: sourcesFile,
    source_count: sources.length,
    sources: srcReports,
    uncovered_refs: uncovered,
    problems,
  };
};
