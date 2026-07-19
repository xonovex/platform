import {basename} from "node:path";

const HEADER_RE = /^##\s+(.*\S)\s*$/;
const URL_FIELD_RE = /\*\*URLs?:\*\*/;
const HTTP_URL_RE = /https?:\/\/\S+/g;
const REVIEWED_RE = /\*\*Last reviewed:\*\*\s*(\d{4}-\d{2}-\d{2})/;
const PROVENANCE_RE = /\*\*Provenance:\*\*\s*(.+\S)\s*$/;
const REF_RE = /references\/([a-z0-9][a-z0-9-]*\.md)/g;
const CHECKOUT_RE = /\*\*Checkout:\*\*\s*(\S+)/;
const VERSION_RE = /\*\*Version:\*\*\s*v?(\d+\.\d+\.\d+)/;
const COMMIT_RE = /\*\*Commit:\*\*\s*([0-9a-f]{7,40})/;
const WATCH_RE = /\*\*Watch:\*\*\s*(.+?)\s*(?:->|→)\s*(.+\S)\s*$/;

export interface Watch {
  readonly path: string;
  readonly refs: readonly string[];
}

export interface Source {
  readonly title: string;
  readonly urls: readonly string[];
  readonly url: string | undefined;
  readonly provenance: string | undefined;
  readonly reviewed: Date | undefined;
  readonly reviewedRaw: string | undefined;
  readonly refs: ReadonlySet<string>;
  readonly lineNo: number;
  readonly checkout: string | undefined;
  readonly version: string | undefined;
  readonly commit: string | undefined;
  readonly watches: readonly Watch[];
}

interface MutableSource {
  title: string;
  urls: string[];
  provenance: string | undefined;
  reviewed: Date | undefined;
  reviewedRaw: string | undefined;
  refs: Set<string>;
  lineNo: number;
  checkout: string | undefined;
  version: string | undefined;
  commit: string | undefined;
  watches: Watch[];
}

const parseIsoDate = (raw: string): Date | undefined => {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(raw);
  if (
    match?.[1] === undefined ||
    match[2] === undefined ||
    match[3] === undefined
  ) {
    return undefined;
  }
  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
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

const findRefs = (line: string): string[] => {
  const found: string[] = [];
  REF_RE.lastIndex = 0;
  let match: RegExpExecArray | null;
  while ((match = REF_RE.exec(line)) !== null) {
    if (match[1] !== undefined) found.push(match[1]);
  }
  return found;
};

const findUrls = (line: string): string[] => {
  HTTP_URL_RE.lastIndex = 0;
  return [...line.matchAll(HTTP_URL_RE)]
    .map((match) => match[0].replace(/[.,;:·]+$/, ""))
    .filter((url) => url.length > 0);
};

export const parseSources = (text: string): Source[] => {
  const sources: MutableSource[] = [];
  let current: MutableSource | undefined;
  let readingUrlList = false;
  const lines = text.split("\n");
  for (const [index, line] of lines.entries()) {
    const header = HEADER_RE.exec(line);
    if (header) {
      current = {
        title: header[1] ?? "",
        urls: [],
        provenance: undefined,
        reviewed: undefined,
        reviewedRaw: undefined,
        refs: new Set<string>(),
        lineNo: index,
        checkout: undefined,
        version: undefined,
        commit: undefined,
        watches: [],
      };
      sources.push(current);
      readingUrlList = false;
      continue;
    }
    if (current === undefined) continue;

    if (URL_FIELD_RE.test(line)) {
      readingUrlList = true;
      current.urls.push(...findUrls(line));
    } else if (readingUrlList && /^\s{2,}-\s+/.test(line)) {
      current.urls.push(...findUrls(line));
    } else {
      readingUrlList = false;
    }

    const provenanceMatch = PROVENANCE_RE.exec(line);
    if (provenanceMatch?.[1] !== undefined) {
      current.provenance = provenanceMatch[1];
    }
    const reviewedMatch = REVIEWED_RE.exec(line);
    if (reviewedMatch?.[1] !== undefined) {
      current.reviewedRaw = reviewedMatch[1];
      current.reviewed = parseIsoDate(reviewedMatch[1]);
    }
    const checkoutMatch = CHECKOUT_RE.exec(line);
    if (checkoutMatch?.[1] !== undefined) {
      current.checkout = checkoutMatch[1];
    }
    const versionMatch = VERSION_RE.exec(line);
    if (versionMatch?.[1] !== undefined) {
      current.version = versionMatch[1];
    }
    const commitMatch = COMMIT_RE.exec(line);
    if (commitMatch?.[1] !== undefined) {
      current.commit = commitMatch[1];
    }
    const watchMatch = WATCH_RE.exec(line);
    if (watchMatch?.[1] !== undefined && watchMatch[2] !== undefined) {
      const refs = watchMatch[2]
        .split(",")
        .map((reference) => basename(reference.trim()))
        .filter((reference) => reference.endsWith(".md"));
      current.watches.push({path: watchMatch[1], refs});
      for (const reference of refs) current.refs.add(reference);
    }
    for (const reference of findRefs(line)) current.refs.add(reference);
  }

  return sources
    .filter(
      (source) => source.urls.length > 0 || source.provenance !== undefined,
    )
    .map((source) => ({
      ...source,
      urls: [...new Set(source.urls)],
      url: source.urls[0],
    }));
};
