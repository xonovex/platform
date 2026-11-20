import {dirname, join, relative, resolve, sep} from "node:path";
import {
  fail,
  pass,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";
import {
  MD_LINK_RE,
  relativeLinkTarget,
} from "@xonovex/script-moon-skill-catalog-common/reference-file-links";
import type {MarkdownDocument} from "./catalog-scan.js";

// The package root beneath a direct or nested package group, or null for a
// file outside every package, from which any link crosses out.
const packageRootOf = (path: string, repoRoot: string): string | null => {
  const parts = relative(repoRoot, path).split(sep);
  const depth = ["runtime", "tooling", "sandbox"].includes(parts[1] ?? "")
    ? 4
    : 3;
  if (parts[0] !== "packages" || parts.length <= depth) return null;
  return join(repoRoot, ...parts.slice(0, depth));
};

const crossesPackageBoundary = (
  sourceFile: string,
  resolvedTarget: string,
  repoRoot: string,
): boolean => {
  const root = packageRootOf(sourceFile, repoRoot);
  if (root === null) return true;
  return resolvedTarget !== root && !resolvedTarget.startsWith(root + sep);
};

// Blank out fenced blocks and inline spans, keeping line structure, so code
// samples cannot read as markdown links. A C call like `listeners[i](&cs)` and
// a documented link shape like `[name](url)` both match the link pattern.
const withoutCodeSpans = (text: string): string =>
  text
    .replaceAll(/```[\s\S]*?```/g, (block) => block.replaceAll(/[^\n]/g, " "))
    .replaceAll(/`[^`\n]*`/g, (span) => " ".repeat(span.length));

interface CrossingLink {
  readonly document: MarkdownDocument;
  readonly raw: string;
  readonly target: string;
}

const crossingLinks = (
  document: MarkdownDocument,
  repoRoot: string,
): readonly CrossingLink[] =>
  [...withoutCodeSpans(document.text).matchAll(MD_LINK_RE)].flatMap((match) => {
    const raw = match[1];
    const target = raw === undefined ? null : relativeLinkTarget(raw);
    if (raw === undefined || target === null) return [];
    const resolvedTarget = resolve(dirname(document.path), target);
    return crossesPackageBoundary(document.path, resolvedTarget, repoRoot)
      ? [{document, raw, target: resolvedTarget}]
      : [];
  });

/**
 * Fails each relative link that leaves its package for a path that does not
 * exist, and passes the set when every such link resolves.
 */
export const checkCrossPackageLinks = (
  documents: readonly MarkdownDocument[],
  repoRoot: string,
  pathExists: (path: string) => boolean,
): readonly Finding[] => {
  const links = documents.flatMap((document) =>
    crossingLinks(document, repoRoot),
  );
  const broken = links
    .filter(({target}) => !pathExists(target))
    .map(({document, raw}) =>
      fail(
        `cross-package links: broken link in ${relative(repoRoot, document.path)} → ${raw}`,
      ),
    );
  return broken.length === 0 && links.length > 0
    ? [
        pass(
          `cross-package links: ${String(links.length)}/${String(links.length)} link(s) resolve`,
        ),
      ]
    : broken;
};
