import {readdirSync, readFileSync, statSync} from "node:fs";
import {basename, dirname, join, relative, resolve, sep} from "node:path";
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

const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);

interface SkillPluginManifest {
  readonly name: string;
  readonly dependencies: readonly string[];
}

interface SkillPackageSurface {
  readonly guideNames: readonly string[];
  readonly markdown: string;
}

const NAMED_GUIDE_RE = /\*\*([a-z0-9][a-z0-9-]*-guide)\*\*/g;

const readSkillPluginManifest = (
  path: string,
  repoRoot: string,
  report: LinkReport,
): SkillPluginManifest | undefined => {
  let value: unknown;
  try {
    value = JSON.parse(readFileSync(path, "utf8")) as unknown;
  } catch (error) {
    report.addFail(
      `skill dependencies: invalid JSON in ${relative(repoRoot, path)}: ${error instanceof Error ? error.message : String(error)}`,
    );
    return undefined;
  }
  if (!isRecord(value) || typeof value.name !== "string") {
    report.addFail(
      `skill dependencies: ${relative(repoRoot, path)} needs a string name`,
    );
    return undefined;
  }
  const dependencies = value.dependencies ?? [];
  if (
    !Array.isArray(dependencies) ||
    dependencies.some((dependency) => typeof dependency !== "string")
  ) {
    report.addFail(
      `skill dependencies: ${relative(repoRoot, path)} dependencies must be strings`,
    );
    return undefined;
  }
  return {name: value.name, dependencies};
};

const sameStrings = (
  left: readonly string[],
  right: readonly string[],
): boolean => {
  const leftSorted = left.toSorted();
  const rightSorted = right.toSorted();
  return (
    leftSorted.length === rightSorted.length &&
    leftSorted.every((value, index) => value === rightSorted[index])
  );
};

const canonicalCycle = (cycle: readonly string[]): string => {
  const withoutRepeatedEnd = cycle.slice(0, -1);
  const rotations = withoutRepeatedEnd.map((_, index) => [
    ...withoutRepeatedEnd.slice(index),
    ...withoutRepeatedEnd.slice(0, index),
  ]);
  const canonical = rotations
    .toSorted((left, right) => left.join("\0").localeCompare(right.join("\0")))
    .at(0);
  const first = canonical?.at(0);
  return canonical === undefined || first === undefined
    ? ""
    : [...canonical, first].join(" → ");
};

export const checkSkillDependencies = (
  repoRoot: string,
  report: LinkReport,
): void => {
  const skillRoot = join(repoRoot, "packages", "skill");
  if (!isDir(skillRoot)) return;
  const manifests = new Map<string, SkillPluginManifest>();
  const surfaces = new Map<string, SkillPackageSurface>();
  let pairs = 0;
  let valid = true;
  for (const pkg of readdirSync(skillRoot).toSorted()) {
    const pkgDir = join(skillRoot, pkg);
    if (!pkg.startsWith("skill-") || !isDir(pkgDir)) continue;
    const claudePath = join(pkgDir, ".claude-plugin", "plugin.json");
    const codexPath = join(pkgDir, ".codex-plugin", "plugin.json");
    const hasGuide = readdirSync(pkgDir).some((entry) =>
      isFile(join(pkgDir, entry, "SKILL.md")),
    );
    if (!hasGuide && !isFile(claudePath) && !isFile(codexPath)) continue;
    if (!isFile(claudePath) || !isFile(codexPath)) {
      valid = false;
      report.addFail(
        `skill dependencies: ${relative(repoRoot, pkgDir)} needs both Claude and Codex manifests`,
      );
      continue;
    }
    const claude = readSkillPluginManifest(claudePath, repoRoot, report);
    const codex = readSkillPluginManifest(codexPath, repoRoot, report);
    if (claude === undefined || codex === undefined) {
      valid = false;
      continue;
    }
    pairs += 1;
    const expectedName = `xonovex-${pkg}`;
    if (claude.name !== expectedName || codex.name !== expectedName) {
      valid = false;
      report.addFail(
        `skill dependencies: manifests in ${relative(repoRoot, pkgDir)} must be named ${expectedName}`,
      );
    }
    if (claude.name !== codex.name) {
      valid = false;
      report.addFail(
        `skill dependencies: manifest names differ in ${relative(repoRoot, pkgDir)} (${claude.name} != ${codex.name})`,
      );
      continue;
    }
    if (!sameStrings(claude.dependencies, codex.dependencies)) {
      valid = false;
      report.addFail(
        `skill dependencies: manifest dependencies differ for ${claude.name}`,
      );
    }
    if (manifests.has(codex.name)) {
      valid = false;
      report.addFail(`skill dependencies: duplicate plugin name ${codex.name}`);
    } else {
      manifests.set(codex.name, codex);
      const guideNames = readdirSync(pkgDir)
        .filter((entry) => isFile(join(pkgDir, entry, "SKILL.md")))
        .toSorted();
      const files = guideNames.flatMap((guide) => {
        const guideDir = join(pkgDir, guide);
        return [
          join(guideDir, "SKILL.md"),
          ...markdownEntries(join(guideDir, "references")),
        ];
      });
      surfaces.set(codex.name, {
        guideNames,
        markdown: files.map((file) => readFileSync(file, "utf8")).join("\n"),
      });
    }
  }

  for (const manifest of manifests.values()) {
    for (const dependency of manifest.dependencies) {
      if (!manifests.has(dependency)) {
        valid = false;
        report.addFail(
          `skill dependencies: ${manifest.name} depends on missing ${dependency}`,
        );
        continue;
      }
      const targetGuides = surfaces.get(dependency)?.guideNames ?? [];
      const sourceMarkdown = surfaces.get(manifest.name)?.markdown ?? "";
      const namedTargets = targetGuides
        .map((guide) => `**${guide}**`)
        .join(" or ");
      if (
        targetGuides.length > 0 &&
        targetGuides.every((guide) => !sourceMarkdown.includes(`**${guide}**`))
      ) {
        valid = false;
        report.addFail(
          `skill dependencies: ${manifest.name} depends on ${dependency} but does not name ${namedTargets} in its guidance`,
        );
      }
    }
  }

  const state = new Map<string, "visiting" | "visited">();
  const cycles = new Set<string>();
  const visit = (name: string, path: readonly string[]): void => {
    if (state.get(name) === "visited") return;
    if (state.get(name) === "visiting") {
      const start = path.indexOf(name);
      cycles.add(canonicalCycle([...path.slice(start), name]));
      return;
    }
    state.set(name, "visiting");
    const manifest = manifests.get(name);
    for (const dependency of manifest?.dependencies ?? []) {
      if (manifests.has(dependency)) visit(dependency, [...path, name]);
    }
    state.set(name, "visited");
  };
  for (const name of manifests.keys()) visit(name, []);
  for (const cycle of cycles) {
    valid = false;
    report.addFail(`skill dependencies: dependency cycle ${cycle}`);
  }

  if (pairs > 0 && valid) {
    report.addPass(
      `skill dependencies: ${String(pairs)} manifest pair(s) agree with no dangling dependencies or cycles`,
    );
  }
};

export const checkNamedSkillHandoffs = (
  files: readonly string[],
  knownGuideNames: ReadonlySet<string>,
  repoRoot: string,
  report: LinkReport,
): {resolved: number; broken: number} => {
  let resolved = 0;
  let broken = 0;
  for (const file of files) {
    const text = readFileSync(file, "utf8");
    for (const match of text.matchAll(NAMED_GUIDE_RE)) {
      const guide = match[1];
      if (guide === undefined) continue;
      if (knownGuideNames.has(guide)) {
        resolved += 1;
      } else {
        broken += 1;
        report.addFail(
          `skill handoffs: ${relative(repoRoot, file)} names missing **${guide}**`,
        );
      }
    }
  }
  return {resolved, broken};
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
  const knownGuideNames = new Set(
    files
      .filter((file) => file.endsWith(`${sep}SKILL.md`))
      .map((file) => basename(dirname(file))),
  );
  const handoffs = checkNamedSkillHandoffs(
    files,
    knownGuideNames,
    repoRoot,
    report,
  );
  if (handoffs.broken === 0 && handoffs.resolved > 0) {
    report.addPass(
      `skill handoffs: ${String(handoffs.resolved)}/${String(handoffs.resolved)} named handoff(s) resolve`,
    );
  }
  checkSkillDependencies(repoRoot, report);
};
