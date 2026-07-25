import {readdirSync, readFileSync, statSync} from "node:fs";
import {basename, dirname, join, relative, resolve, sep} from "node:path";
import {
  MD_LINK_RE,
  relativeLinkTarget,
  type LinkReport,
} from "./reference-file-links.js";

// This module validates relative links that cross package boundaries.

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

const isStringArray = (value: unknown): value is readonly string[] =>
  Array.isArray(value) && value.every((entry) => typeof entry === "string");

const isStringRecord = (
  value: unknown,
): value is Readonly<Record<string, string>> =>
  isRecord(value) &&
  Object.values(value).every((entry) => typeof entry === "string");

const normalizedSkillPaths = (
  value: unknown,
): readonly string[] | undefined => {
  if (typeof value === "string") return [value];
  return isStringArray(value) ? value : undefined;
};

interface SkillPluginManifest {
  readonly name: string;
  readonly dependencies: readonly string[];
  readonly skills: readonly string[];
}

interface SkillPackageSurface {
  readonly guideNames: readonly string[];
  readonly markdown: string;
  readonly references: readonly {
    readonly path: string;
    readonly text: string;
  }[];
}

interface SkillPackageCheck {
  readonly valid: boolean;
  readonly pair: boolean;
  readonly manifest: SkillPluginManifest | undefined;
  readonly surface: SkillPackageSurface | undefined;
}

const NAMED_GUIDE_RE = /\*\*([a-z0-9][a-z0-9-]*-guide)\*\*/g;

const readSkillPluginManifest = (
  path: string,
  kind: "Claude" | "Codex",
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
  const skillsValue = value.skills;
  const skills = normalizedSkillPaths(skillsValue);
  const expectedShape = kind === "Claude" ? "a string array" : "a string";
  const hasExpectedShape =
    kind === "Claude"
      ? Array.isArray(skillsValue)
      : typeof skillsValue === "string";
  if (!hasExpectedShape || skills === undefined || skills.length === 0) {
    report.addFail(
      `skill packaging: ${relative(repoRoot, path)} skills must be ${expectedShape}`,
    );
    return undefined;
  }
  return {name: value.name, dependencies, skills};
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

const skillPackageSurface = (pkgDir: string): SkillPackageSurface => {
  const guideNames = readdirSync(pkgDir)
    .filter((entry) => isFile(join(pkgDir, entry, "SKILL.md")))
    .toSorted();
  const references = guideNames.flatMap((guide) =>
    markdownEntries(join(pkgDir, guide, "references")).map((path) => ({
      path: relative(pkgDir, path),
      text: readFileSync(path, "utf8"),
    })),
  );
  const files = guideNames.flatMap((guide) => {
    const guideDir = join(pkgDir, guide);
    return [
      join(guideDir, "SKILL.md"),
      ...markdownEntries(join(guideDir, "references")),
    ];
  });
  return {
    guideNames,
    markdown: files.map((file) => readFileSync(file, "utf8")).join("\n"),
    references,
  };
};

const contentShingles = (text: string): ReadonlySet<string> => {
  const words = text
    .toLowerCase()
    .replaceAll(/```[\s\S]*?```/g, " ")
    .replaceAll(/[^a-z0-9_]+/g, " ")
    .trim()
    .split(/\s+/)
    .filter(Boolean);
  const shingles = new Set<string>();
  for (let index = 0; index + 4 < words.length; index += 1) {
    shingles.add(words.slice(index, index + 5).join(" "));
  }
  return shingles;
};

const shingleContainment = (
  left: ReadonlySet<string>,
  right: ReadonlySet<string>,
): number => {
  const denominator = Math.min(left.size, right.size);
  if (denominator < 20) return 0;
  let shared = 0;
  for (const shingle of left) {
    if (right.has(shingle)) shared += 1;
  }
  return shared / denominator;
};

const checkDependencyReferenceOverlap = (
  sourceName: string,
  targetName: string,
  source: SkillPackageSurface,
  target: SkillPackageSurface,
  report: LinkReport,
): boolean => {
  let valid = true;
  for (const sourceReference of source.references) {
    const sourceShingles = contentShingles(sourceReference.text);
    for (const targetReference of target.references) {
      const containment = shingleContainment(
        sourceShingles,
        contentShingles(targetReference.text),
      );
      if (containment < 0.3) continue;
      valid = false;
      report.addFail(
        `skill ownership: ${sourceName} ${sourceReference.path} duplicates ${targetName} ${targetReference.path} (${String(Math.round(containment * 100))}% five-word-shingle containment); keep only the dependent skill's specialization`,
      );
    }
  }
  return valid;
};

const checkSkillPackage = (
  pkgDir: string,
  pkg: string,
  repoRoot: string,
  report: LinkReport,
): SkillPackageCheck | undefined => {
  const claudePath = join(pkgDir, ".claude-plugin", "plugin.json");
  const codexPath = join(pkgDir, ".codex-plugin", "plugin.json");
  const hasGuide = readdirSync(pkgDir).some((entry) =>
    isFile(join(pkgDir, entry, "SKILL.md")),
  );
  if (!hasGuide && !isFile(claudePath) && !isFile(codexPath)) return undefined;
  if (!isFile(claudePath) || !isFile(codexPath)) {
    report.addFail(
      `skill dependencies: ${relative(repoRoot, pkgDir)} needs both Claude and Codex manifests`,
    );
    return {valid: false, pair: false, manifest: undefined, surface: undefined};
  }
  const claude = readSkillPluginManifest(
    claudePath,
    "Claude",
    repoRoot,
    report,
  );
  const codex = readSkillPluginManifest(codexPath, "Codex", repoRoot, report);
  if (claude === undefined || codex === undefined) {
    return {valid: false, pair: false, manifest: undefined, surface: undefined};
  }

  let valid = true;
  const expectedName = `xonovex-${pkg}`;
  if (claude.name !== expectedName || codex.name !== expectedName) {
    valid = false;
    report.addFail(
      `skill dependencies: manifests in ${relative(repoRoot, pkgDir)} must be named ${expectedName}`,
    );
  }
  if (claude.name !== codex.name) {
    report.addFail(
      `skill dependencies: manifest names differ in ${relative(repoRoot, pkgDir)} (${claude.name} != ${codex.name})`,
    );
    return {valid: false, pair: true, manifest: undefined, surface: undefined};
  }
  if (!sameStrings(claude.dependencies, codex.dependencies)) {
    valid = false;
    report.addFail(
      `skill dependencies: manifest dependencies differ for ${claude.name}`,
    );
  }
  if (!sameStrings(claude.skills, codex.skills)) {
    valid = false;
    report.addFail(
      `skill packaging: manifest skill paths differ for ${claude.name}`,
    );
  }
  const surface = skillPackageSurface(pkgDir);
  const expectedSkills = surface.guideNames.map((guide) => `./${guide}`);
  for (const [kind, manifest] of [
    ["Claude", claude],
    ["Codex", codex],
  ] as const) {
    if (!sameStrings(manifest.skills, expectedSkills)) {
      valid = false;
      report.addFail(
        `skill packaging: ${kind} manifest in ${relative(repoRoot, pkgDir)} must point directly to ${expectedSkills.join(", ")}`,
      );
    }
  }
  return {
    valid,
    pair: true,
    manifest: codex,
    surface,
  };
};

const checkDeclaredDependencies = (
  manifests: ReadonlyMap<string, SkillPluginManifest>,
  surfaces: ReadonlyMap<string, SkillPackageSurface>,
  repoRoot: string,
  report: LinkReport,
): boolean => {
  let valid = true;
  for (const manifest of manifests.values()) {
    const packagePath = join(
      repoRoot,
      "packages",
      "skill",
      manifest.name.replace(/^xonovex-/u, ""),
      "package.json",
    );
    let packageDependencies: Readonly<Record<string, string>> | undefined;
    try {
      const packageJson = JSON.parse(
        readFileSync(packagePath, "utf8"),
      ) as unknown;
      if (isRecord(packageJson) && isStringRecord(packageJson.dependencies)) {
        packageDependencies = packageJson.dependencies;
      } else if (
        isRecord(packageJson) &&
        packageJson.dependencies === undefined
      ) {
        packageDependencies = {};
      } else {
        valid = false;
        report.addFail(
          `skill dependencies: ${relative(repoRoot, packagePath)} dependencies must map package names to versions`,
        );
      }
    } catch (error) {
      valid = false;
      report.addFail(
        `skill dependencies: cannot read ${relative(repoRoot, packagePath)}: ${error instanceof Error ? error.message : String(error)}`,
      );
    }
    for (const npmDependency of Object.keys(packageDependencies ?? {}).filter(
      (dependency) => dependency.startsWith("@xonovex/skill-"),
    )) {
      const pluginDependency = `xonovex-${npmDependency.slice("@xonovex/".length)}`;
      if (!manifest.dependencies.includes(pluginDependency)) {
        valid = false;
        report.addFail(
          `skill dependencies: ${manifest.name} declares ${npmDependency} in package.json but omits ${pluginDependency} from its plugin manifests`,
        );
      }
    }
    for (const dependency of manifest.dependencies) {
      if (!manifests.has(dependency)) {
        valid = false;
        report.addFail(
          `skill dependencies: ${manifest.name} depends on missing ${dependency}`,
        );
        continue;
      }
      const npmDependency = `@xonovex/${dependency.replace(/^xonovex-/u, "")}`;
      const declaredVersion = packageDependencies?.[npmDependency];
      if (declaredVersion === undefined) {
        valid = false;
        report.addFail(
          `skill dependencies: ${manifest.name} declares ${dependency} in its plugin manifests but omits ${npmDependency} from package.json`,
        );
      } else {
        const dependencyPackagePath = join(
          repoRoot,
          "packages",
          "skill",
          dependency.replace(/^xonovex-/u, ""),
          "package.json",
        );
        try {
          const dependencyPackage = JSON.parse(
            readFileSync(dependencyPackagePath, "utf8"),
          ) as unknown;
          const expectedVersion =
            isRecord(dependencyPackage) &&
            typeof dependencyPackage.version === "string"
              ? dependencyPackage.version
              : undefined;
          if (expectedVersion === undefined) {
            valid = false;
            report.addFail(
              `skill dependencies: ${relative(repoRoot, dependencyPackagePath)} needs a string version`,
            );
          } else if (declaredVersion !== expectedVersion) {
            valid = false;
            report.addFail(
              `skill dependencies: ${manifest.name} pins ${npmDependency}@${declaredVersion}; expected exact installed version ${expectedVersion}`,
            );
          }
        } catch (error) {
          valid = false;
          report.addFail(
            `skill dependencies: cannot read ${relative(repoRoot, dependencyPackagePath)}: ${error instanceof Error ? error.message : String(error)}`,
          );
        }
      }
      const sourceSurface = surfaces.get(manifest.name);
      const targetSurface = surfaces.get(dependency);
      if (sourceSurface !== undefined && targetSurface !== undefined) {
        valid =
          checkDependencyReferenceOverlap(
            manifest.name,
            dependency,
            sourceSurface,
            targetSurface,
            report,
          ) && valid;
      }
    }
  }
  return valid;
};

const dependencyCycles = (
  manifests: ReadonlyMap<string, SkillPluginManifest>,
): ReadonlySet<string> => {
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
  return cycles;
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
    const checked = checkSkillPackage(pkgDir, pkg, repoRoot, report);
    if (checked === undefined) continue;
    valid &&= checked.valid;
    if (checked.pair) pairs += 1;
    if (checked.manifest === undefined || checked.surface === undefined) {
      continue;
    }
    if (manifests.has(checked.manifest.name)) {
      valid = false;
      report.addFail(
        `skill dependencies: duplicate plugin name ${checked.manifest.name}`,
      );
    } else {
      manifests.set(checked.manifest.name, checked.manifest);
      surfaces.set(checked.manifest.name, checked.surface);
    }
  }

  valid =
    checkDeclaredDependencies(manifests, surfaces, repoRoot, report) && valid;
  for (const cycle of dependencyCycles(manifests)) {
    valid = false;
    report.addFail(`skill dependencies: dependency cycle ${cycle}`);
  }

  if (valid && pairs > 0) {
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

// collectSkillMarkdown returns guide and reference Markdown across skill packages.
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

// collectCommandMarkdown returns the published Markdown surface of every command
// package so cross-package references do not depend on one special command family.
const collectCommandMarkdown = (repoRoot: string): string[] => {
  const root = join(repoRoot, "packages", "command");
  if (!isDir(root)) return [];
  return readdirSync(root)
    .toSorted()
    .flatMap((pkg) => {
      const base = join(root, pkg);
      if (!isDir(base)) return [];
      return [
        join(base, "README.md"),
        ...markdownEntries(join(base, "docs")),
        ...markdownEntries(join(base, "commands")),
      ].filter(isFile);
    });
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
    ...collectCommandMarkdown(repoRoot),
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
