import {join} from "node:path";
import {fail, pass, type Finding} from "./finding.js";

export const MD_LINK_RE = /\[[^\]]*\]\(([^)\s]+)\)/g;
const EXTERNAL_LINK_RE = /^(?:https?|mailto):/i;

/** One Markdown file directly under a skill's references directory. */
export interface ReferenceDocument {
  /** The file name inside the references directory. */
  readonly name: string;
  readonly text: string;
}

// The file a markdown link points at, with any in-page fragment stripped, or
// null when the link names no real sibling file: an external scheme
// (http/https/mailto), a placeholder or elided illustrative form, or a pure
// #anchor.
export const relativeLinkTarget = (raw: string): string | null => {
  if (
    EXTERNAL_LINK_RE.test(raw) ||
    raw.includes("<") ||
    raw.includes("{") ||
    raw.includes("…")
  )
    return null;
  const target = raw.split("#", 1)[0] ?? "";
  return target === "" ? null : target;
};

// checkReferenceFileLinks resolves links relative to their containing document,
// which sits directly in referencesDirectory.
export const checkReferenceFileLinks = (
  referencesDirectory: string,
  documents: readonly ReferenceDocument[],
  isFile: (path: string) => boolean,
): readonly Finding[] => {
  const links = documents.flatMap(({name, text}) =>
    [...text.matchAll(MD_LINK_RE)].flatMap((match) => {
      const raw = match[1];
      const target = raw === undefined ? null : relativeLinkTarget(raw);
      if (raw === undefined || target === null) return [];
      return [{raw, name, resolves: isFile(join(referencesDirectory, target))}];
    }),
  );
  const failures = links
    .filter(({resolves}) => !resolves)
    .map(({name, raw}) =>
      fail(`reference links: broken link in references/${name} → ${raw}`),
    );
  const resolved = links.length - failures.length;
  return failures.length === 0 && resolved > 0
    ? [
        pass(
          `reference links: ${String(resolved)}/${String(resolved)} link(s) resolve`,
        ),
      ]
    : failures;
};
