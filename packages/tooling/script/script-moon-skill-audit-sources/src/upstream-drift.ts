import type {Source} from "./source-parser.js";
import type {UpstreamRepository} from "./upstream.js";

export interface DriftReport {
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

type Semver = readonly [number, number, number];

const MAX_LISTED_COMMITS = 20;

const parseSemver = (text: string): Semver | undefined => {
  const m = /(\d+)\.(\d+)\.(\d+)/.exec(text);
  return m?.[1] === undefined || m[2] === undefined || m[3] === undefined
    ? undefined
    : [Number(m[1]), Number(m[2]), Number(m[3])];
};

const cmpSemver = (a: Semver, b: Semver): number =>
  a[0] - b[0] || a[1] - b[1] || a[2] - b[2];

const latestSemverTag = (
  tags: readonly string[],
): {version: Semver; raw: string} | undefined => {
  let best: {version: Semver; raw: string} | undefined;
  for (const line of tags) {
    const raw = line.trim();
    if (!/^v?\d+\.\d+\.\d+$/.test(raw)) continue;
    const version = parseSemver(raw);
    if (version && (!best || cmpSemver(version, best.version) > 0)) {
      best = {version, raw};
    }
  }
  return best;
};

/**
 * Compares a source pinned to an upstream checkout with that checkout: the
 * latest released tag against the pinned version, and the commits and files
 * changed on the watched paths since the pinned commit or version. An absent
 * repository is a checkout that is not a git repository.
 */
export const computeDrift = (
  checkout: string,
  source: Source,
  pull: boolean,
  repository?: UpstreamRepository,
): DriftReport => {
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
  if (repository === undefined) {
    return {...report, note: "checkout not found or not a git repo"};
  }
  report.resolved = true;
  if (pull && !repository.fetchTags()) {
    return {
      ...report,
      pull_failed: true,
      note: "failed to fetch upstream tags",
    };
  }

  const latest = latestSemverTag(repository.tags());
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
  if (!repository.hasCommit(ref)) {
    return {...report, note: `pinned ref '${ref}' not found in checkout`};
  }

  // Upper bound: the latest released tag (deterministic after --pull) else HEAD.
  const range = `${ref}..${latest ? latest.raw : "HEAD"}`;
  const paths = source.watches.map((w) => w.path);
  const commits = repository.commits(range, paths);
  report.commit_count = commits.length;
  report.commits = commits.slice(0, MAX_LISTED_COMMITS);

  const reviewRefs = new Set<string>();
  for (const file of repository.changedFiles(range, paths)) {
    for (const w of source.watches) {
      if (file.startsWith(w.path)) for (const r of w.refs) reviewRefs.add(r);
    }
  }
  report.review_refs = [...reviewRefs].toSorted();
  return report;
};
