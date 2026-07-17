import {readdirSync, readFileSync, statSync} from "node:fs";
import {dirname, join, relative, resolve, sep} from "node:path";
import {
  MD_LINK_RE,
  relativeLinkTarget,
  type LinkReport,
} from "./reference-file-links.js";

// A composition doc's link to a contract in another skill or package is a fact:
// the target file must exist. checkReferenceFileLinks already validates the
// intra-skill links inside each references/*.md, and command-workflow's own
// documentation check validates links that stay inside command-workflow/docs.
// The gap this guard closes is the boundary-crossing links — a SKILL.md or a
// command delegation pointing at ../../../skill/<other>/…/contract.md — which no
// per-skill run and no command-package run resolves. It scans the composition
// surface once and fails on any cross-package link whose target has moved or been
// renamed. Intra-package links are left to the checks that already own them, so a
// link is validated here only when it escapes its own packages/<layer>/<pkg> root.

const exists = (path: string): boolean => {
  try {
    statSync(path);
    return true;
  } catch {
    return false;
  }
};

const isDir = (path: string): boolean => {
  try {
    return statSync(path).isDirectory();
  } catch {
    return false;
  }
};

const isFile = (path: string): boolean => {
  try {
    return statSync(path).isFile();
  } catch {
    return false;
  }
};

const markdownEntries = (dir: string): string[] =>
  isDir(dir)
    ? readdirSync(dir)
        .filter((entry) => entry.endsWith(".md"))
        .filter((entry) => isFile(join(dir, entry)))
        .toSorted()
        .map((entry) => join(dir, entry))
    : [];

// Every skill's SKILL.md and its references/*.md, across the two-level skill
// package layout (packages/skill/<package>/<guide>/).
const collectSkillMarkdown = (repoRoot: string): string[] => {
  const skillRoot = join(repoRoot, "packages", "skill");
  if (!isDir(skillRoot)) return [];
  const files: string[] = [];
  for (const pkg of readdirSync(skillRoot).toSorted()) {
    const pkgDir = join(skillRoot, pkg);
    if (!isDir(pkgDir)) continue;
    for (const guide of readdirSync(pkgDir).toSorted()) {
      const guideDir = join(pkgDir, guide);
      const skillMd = join(guideDir, "SKILL.md");
      if (!isFile(skillMd)) continue;
      files.push(skillMd, ...markdownEntries(join(guideDir, "references")));
    }
  }
  return files;
};

// command-workflow's docs and command delegations. Its README/MIGRATION/docs
// links are already validated by the package's own documentation check; the
// commands/*.md delegations are not, and neither run resolves cross-package
// targets, which this guard does.
const collectCommandWorkflowMarkdown = (repoRoot: string): string[] => {
  const base = join(repoRoot, "packages", "command", "command-workflow");
  return [
    ...markdownEntries(join(base, "docs")),
    ...markdownEntries(join(base, "commands")),
  ];
};

// The packages/<layer>/<package> root a file lives under, or null when the file
// is not inside a package (in which case any link is treated as crossing out).
const packageRootOf = (absFile: string, repoRoot: string): string | null => {
  const [top, layer, pkg] = relative(repoRoot, absFile).split(sep);
  if (top !== "packages" || layer === undefined || pkg === undefined) {
    return null;
  }
  return join(repoRoot, top, layer, pkg);
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

export interface LinkCounts {
  resolved: number;
  broken: number;
}

// Validate the boundary-crossing relative links in an explicit set of markdown
// files. Exposed for tests, which supply a synthetic package layout under a
// throwaway repo root.
export const checkMarkdownFilesForCrossPackageLinks = (
  files: readonly string[],
  repoRoot: string,
  report: LinkReport,
): LinkCounts => {
  let resolved = 0;
  let broken = 0;
  for (const file of files) {
    const dir = dirname(file);
    const text = readFileSync(file, "utf8");
    for (const m of text.matchAll(MD_LINK_RE)) {
      const raw = m[1];
      if (raw === undefined) continue;
      const target = relativeLinkTarget(raw);
      if (target === null) continue;
      const resolvedTarget = resolve(dir, target);
      if (!crossesPackageBoundary(file, resolvedTarget, repoRoot)) continue;
      if (exists(resolvedTarget)) {
        resolved += 1;
      } else {
        report.addFail(
          `cross-package links: broken link in ${relative(repoRoot, file)} → ${raw}`,
        );
        broken += 1;
      }
    }
  }
  return {resolved, broken};
};

// Scan the whole composition surface for broken cross-package links.
export const checkCrossPackageLinks = (
  repoRoot: string,
  report: LinkReport,
): void => {
  const files = [
    ...collectSkillMarkdown(repoRoot),
    ...collectCommandWorkflowMarkdown(repoRoot),
  ];
  const {resolved, broken} = checkMarkdownFilesForCrossPackageLinks(
    files,
    repoRoot,
    report,
  );
  if (broken === 0 && resolved > 0) {
    report.addPass(
      `cross-package links: ${String(resolved)}/${String(resolved)} link(s) resolve`,
    );
  }
};
