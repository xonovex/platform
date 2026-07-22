import {basename} from "node:path";

const HEADER_RE = /^##\s+(.*\S)\s*$/;
const URL_FIELD_RE = /\*\*URLs?:\*\*/;
const HTTP_URL_RE = /https?:\/\/\S+/g;
const REVIEWED_RE = /\*\*Last reviewed:\*\*\s*(\d{4}-\d{2}-\d{2})/;
const PROVENANCE_RE = /\*\*Provenance:\*\*\s*(.+\S)\s*$/;
const REF_RE = /references\/([a-z0-9][a-z0-9-]*\.md)/g;
const ARROW_REF_RE = /(?:->|→)\s*`?([a-z0-9][a-z0-9-]*\.md)`?/g;
const REFERENCES_RE = /\*\*References:\*\*\s*(.*\S)?\s*$/;
const LEGACY_ALL_REFERENCES_RE =
  /\*\*Used for:\*\*.*\b(?:all `references\/`|all references|every [^\n]* reference)/i;
const CHECKOUT_RE = /\*\*Checkout:\*\*\s*(\S+)/;
const VERSION_RE = /\*\*Version:\*\*\s*(.+\S)\s*$/;
const COMMIT_RE = /\*\*Commit:\*\*\s*([0-9a-f]{7,40})/;
const WATCH_RE = /\*\*Watch:\*\*\s*(.+?)\s*(?:->|→)\s*(.+\S)\s*$/;

interface Watch {
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
  readonly coversAllReferences: boolean;
  readonly lineNo: number;
  readonly checkout: string | undefined;
  readonly version: string | undefined;
  readonly commit: string | undefined;
  readonly watches: readonly Watch[];
}

export const hasReferenceMapping = (
  source: Pick<Source, "refs" | "coversAllReferences">,
): boolean => source.coversAllReferences || source.refs.size > 0;

interface MutableSource {
  title: string;
  urls: string[];
  provenance: string | undefined;
  reviewed: Date | undefined;
  reviewedRaw: string | undefined;
  refs: Set<string>;
  coversAllReferences: boolean;
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

const findArrowRefs = (line: string): string[] => {
  ARROW_REF_RE.lastIndex = 0;
  return [...line.matchAll(ARROW_REF_RE)]
    .map((match) => match[1])
    .filter((reference): reference is string => reference !== undefined);
};

const parseReferencesField = (
  value: string,
): {all: boolean; refs: readonly string[]} => {
  const normalized = value.trim().replace(/^[-:]\s*/, "");
  if (/^all(?:\s+references)?$/i.test(normalized)) {
    return {all: true, refs: []};
  }
  const refs = normalized
    .split(",")
    .map((reference) => basename(reference.trim().replaceAll("`", "")))
    .filter((reference) => /^[a-z0-9][a-z0-9-]*\.md$/.test(reference));
  return {all: false, refs};
};

const findUrls = (line: string): string[] => {
  HTTP_URL_RE.lastIndex = 0;
  return [...line.matchAll(HTTP_URL_RE)]
    .map((match) => match[0].replace(/[.,;:·]+$/, ""))
    .filter((url) => url.length > 0);
};

const createSource = (title: string, lineNo: number): MutableSource => ({
  title,
  urls: [],
  provenance: undefined,
  reviewed: undefined,
  reviewedRaw: undefined,
  refs: new Set<string>(),
  coversAllReferences: false,
  lineNo,
  checkout: undefined,
  version: undefined,
  commit: undefined,
  watches: [],
});

const collectUrls = (
  source: MutableSource,
  line: string,
  readingList: boolean,
): boolean => {
  if (URL_FIELD_RE.test(line)) {
    source.urls.push(...findUrls(line));
    return true;
  }
  if (readingList && /^\s{2,}-\s+/.test(line)) {
    source.urls.push(...findUrls(line));
    return true;
  }
  return false;
};

const collectReferences = (
  source: MutableSource,
  line: string,
  readingList: boolean,
): boolean => {
  const field = REFERENCES_RE.exec(line)?.[1];
  const listItem = readingList && /^\s{2,}-\s+/.test(line);
  if (field !== undefined || listItem) {
    const parsed = parseReferencesField(field ?? line);
    source.coversAllReferences ||= parsed.all;
    for (const reference of parsed.refs) source.refs.add(reference);
  }
  if (LEGACY_ALL_REFERENCES_RE.test(line)) {
    source.coversAllReferences = true;
  }
  return field !== undefined || listItem;
};

const collectMetadata = (source: MutableSource, line: string): void => {
  const provenance = PROVENANCE_RE.exec(line)?.[1];
  if (provenance !== undefined) source.provenance = provenance;
  const reviewed = REVIEWED_RE.exec(line)?.[1];
  if (reviewed !== undefined) {
    source.reviewedRaw = reviewed;
    source.reviewed = parseIsoDate(reviewed);
  }
  const checkout = CHECKOUT_RE.exec(line)?.[1];
  if (checkout !== undefined) source.checkout = checkout;
  const version = VERSION_RE.exec(line)?.[1];
  if (version !== undefined) {
    const trimmed = version.trim();
    source.version = trimmed.startsWith("`")
      ? /^`([^`]+)`/.exec(trimmed)?.[1]
      : trimmed;
  }
  const commit = COMMIT_RE.exec(line)?.[1];
  if (commit !== undefined) source.commit = commit;

  const watch = WATCH_RE.exec(line);
  if (watch?.[1] !== undefined && watch[2] !== undefined) {
    const refs = watch[2]
      .split(",")
      .map((reference) => basename(reference.trim()))
      .filter((reference) => reference.endsWith(".md"));
    source.watches.push({path: watch[1], refs});
    for (const reference of refs) source.refs.add(reference);
  }
  for (const reference of findRefs(line)) source.refs.add(reference);
  for (const reference of findArrowRefs(line)) source.refs.add(reference);
};

export const parseSources = (text: string): Source[] => {
  const sources: MutableSource[] = [];
  let current: MutableSource | undefined;
  let readingUrlList = false;
  let readingReferenceList = false;
  const lines = text.split("\n");
  for (const [index, line] of lines.entries()) {
    const header = HEADER_RE.exec(line);
    if (header) {
      current = createSource(header[1] ?? "", index);
      sources.push(current);
      readingUrlList = false;
      readingReferenceList = false;
      continue;
    }
    if (current === undefined) continue;

    readingUrlList = collectUrls(current, line, readingUrlList);
    readingReferenceList = collectReferences(
      current,
      line,
      readingReferenceList,
    );
    collectMetadata(current, line);
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
