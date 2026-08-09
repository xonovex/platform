import {basename, dirname, join, relative, resolve, sep} from "node:path";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import {isRecord} from "@xonovex/script-moon-common/records";
import {
  MD_LINK_RE,
  relativeLinkTarget,
  type LinkReport,
} from "@xonovex/script-moon-skill-catalog-common/reference-file-links";

// This module validates relative links that cross package boundaries.

// A path of either kind, which is what a link target may legitimately be.
const exists = (path: string, fs: FileSystem): boolean =>
  fs.isFile(path) || fs.isDirectory(path);

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
  fs: FileSystem,
): SkillPluginManifest | undefined => {
  let value: unknown;
  try {
    value = JSON.parse(fs.readText(path)) as unknown;
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

const skillPackageSurface = (
  pkgDir: string,
  fs: FileSystem,
): SkillPackageSurface => {
  const guideNames = fs
    .readDirectory(pkgDir)
    .filter((entry) => fs.isFile(join(pkgDir, entry, "SKILL.md")))
    .toSorted();
  const references = guideNames.flatMap((guide) =>
    markdownEntries(join(pkgDir, guide, "references"), fs).map((path) => ({
      path: relative(pkgDir, path),
      text: fs.readText(path),
    })),
  );
  const files = guideNames.flatMap((guide) => {
    const guideDir = join(pkgDir, guide);
    return [
      join(guideDir, "SKILL.md"),
      ...markdownEntries(join(guideDir, "references"), fs),
    ];
  });
  return {
    guideNames,
    markdown: files.map((file) => fs.readText(file)).join("\n"),
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
  fs: FileSystem,
): SkillPackageCheck | undefined => {
  const claudePath = join(pkgDir, ".claude-plugin", "plugin.json");
  const codexPath = join(pkgDir, ".codex-plugin", "plugin.json");
  const hasGuide = fs
    .readDirectory(pkgDir)
    .some((entry) => fs.isFile(join(pkgDir, entry, "SKILL.md")));
  if (!hasGuide && !fs.isFile(claudePath) && !fs.isFile(codexPath)) {
    return undefined;
  }
  if (!fs.isFile(claudePath) || !fs.isFile(codexPath)) {
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
    fs,
  );
  const codex = readSkillPluginManifest(
    codexPath,
    "Codex",
    repoRoot,
    report,
    fs,
  );
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
  const surface = skillPackageSurface(pkgDir, fs);
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

// readPackageDependencies returns a skill package's dependencies, or undefined
// when package.json is unreadable or its dependencies are not a string map.
const readPackageDependencies = (
  packagePath: string,
  repoRoot: string,
  report: LinkReport,
  fs: FileSystem,
): Readonly<Record<string, string>> | undefined => {
  try {
    const packageJson = JSON.parse(fs.readText(packagePath)) as unknown;
    if (isRecord(packageJson) && isStringRecord(packageJson.dependencies)) {
      return packageJson.dependencies;
    }
    if (isRecord(packageJson) && packageJson.dependencies === undefined) {
      return {};
    }
    report.addFail(
      `skill dependencies: ${relative(repoRoot, packagePath)} dependencies must map package names to versions`,
    );
  } catch (error) {
    report.addFail(
      `skill dependencies: cannot read ${relative(repoRoot, packagePath)}: ${error instanceof Error ? error.message : String(error)}`,
    );
  }
  return undefined;
};

// checkDependencyVersion verifies a dependent pins the exact version the
// dependency's own package.json declares.
const checkDependencyVersion = (
  manifestName: string,
  dependency: string,
  npmDependency: string,
  declaredVersion: string,
  repoRoot: string,
  report: LinkReport,
  fs: FileSystem,
): boolean => {
  const dependencyPackagePath = join(
    repoRoot,
    "packages",
    "skill",
    dependency.replace(/^xonovex-/u, ""),
    "package.json",
  );
  try {
    const dependencyPackage = JSON.parse(
      fs.readText(dependencyPackagePath),
    ) as unknown;
    const expectedVersion =
      isRecord(dependencyPackage) &&
      typeof dependencyPackage.version === "string"
        ? dependencyPackage.version
        : undefined;
    if (expectedVersion === undefined) {
      report.addFail(
        `skill dependencies: ${relative(repoRoot, dependencyPackagePath)} needs a string version`,
      );
      return false;
    }
    if (declaredVersion !== expectedVersion) {
      report.addFail(
        `skill dependencies: ${manifestName} pins ${npmDependency}@${declaredVersion}; expected exact installed version ${expectedVersion}`,
      );
      return false;
    }
    return true;
  } catch (error) {
    report.addFail(
      `skill dependencies: cannot read ${relative(repoRoot, dependencyPackagePath)}: ${error instanceof Error ? error.message : String(error)}`,
    );
    return false;
  }
};

// checkManifestDependency verifies one plugin-manifest dependency resolves, is
// pinned in package.json, and shares reference surface with its dependent.
const checkManifestDependency = (
  manifest: SkillPluginManifest,
  dependency: string,
  packageDependencies: Readonly<Record<string, string>> | undefined,
  manifests: ReadonlyMap<string, SkillPluginManifest>,
  surfaces: ReadonlyMap<string, SkillPackageSurface>,
  repoRoot: string,
  report: LinkReport,
  fs: FileSystem,
): boolean => {
  if (!manifests.has(dependency)) {
    report.addFail(
      `skill dependencies: ${manifest.name} depends on missing ${dependency}`,
    );
    return false;
  }
  const npmDependency = `@xonovex/${dependency.replace(/^xonovex-/u, "")}`;
  const declaredVersion = packageDependencies?.[npmDependency];
  let valid: boolean;
  if (declaredVersion === undefined) {
    valid = false;
    report.addFail(
      `skill dependencies: ${manifest.name} declares ${dependency} in its plugin manifests but omits ${npmDependency} from package.json`,
    );
  } else {
    valid = checkDependencyVersion(
      manifest.name,
      dependency,
      npmDependency,
      declaredVersion,
      repoRoot,
      report,
      fs,
    );
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
  return valid;
};

const checkDeclaredDependencies = (
  manifests: ReadonlyMap<string, SkillPluginManifest>,
  surfaces: ReadonlyMap<string, SkillPackageSurface>,
  repoRoot: string,
  report: LinkReport,
  fs: FileSystem,
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
    const packageDependencies = readPackageDependencies(
      packagePath,
      repoRoot,
      report,
      fs,
    );
    if (packageDependencies === undefined) valid = false;
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
      valid =
        checkManifestDependency(
          manifest,
          dependency,
          packageDependencies,
          manifests,
          surfaces,
          repoRoot,
          report,
          fs,
        ) && valid;
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
  fs: FileSystem = nodeFileSystem,
): void => {
  const skillRoot = join(repoRoot, "packages", "skill");
  if (!fs.isDirectory(skillRoot)) return;
  const manifests = new Map<string, SkillPluginManifest>();
  const surfaces = new Map<string, SkillPackageSurface>();
  let pairs = 0;
  let valid = true;
  for (const pkg of fs.readDirectory(skillRoot).toSorted()) {
    const pkgDir = join(skillRoot, pkg);
    if (!pkg.startsWith("skill-") || !fs.isDirectory(pkgDir)) continue;
    const checked = checkSkillPackage(pkgDir, pkg, repoRoot, report, fs);
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
    checkDeclaredDependencies(manifests, surfaces, repoRoot, report, fs) &&
    valid;
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
  fs: FileSystem = nodeFileSystem,
): {resolved: number; broken: number} => {
  let resolved = 0;
  let broken = 0;
  for (const file of files) {
    const text = fs.readText(file);
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

const markdownEntries = (dir: string, fs: FileSystem): string[] =>
  fs.isDirectory(dir)
    ? fs
        .readDirectory(dir)
        .filter((entry) => entry.endsWith(".md"))
        .filter((entry) => fs.isFile(join(dir, entry)))
        .toSorted()
        .map((entry) => join(dir, entry))
    : [];

// collectSkillMarkdown returns guide and reference Markdown across skill packages.
const collectSkillMarkdown = (repoRoot: string, fs: FileSystem): string[] => {
  const skillRoot = join(repoRoot, "packages", "skill");
  if (!fs.isDirectory(skillRoot)) return [];
  const files: string[] = [];
  for (const pkg of fs.readDirectory(skillRoot).toSorted()) {
    const pkgDir = join(skillRoot, pkg);
    if (!fs.isDirectory(pkgDir)) continue;
    for (const guide of fs.readDirectory(pkgDir).toSorted()) {
      const guideDir = join(pkgDir, guide);
      const skillMd = join(guideDir, "SKILL.md");
      if (!fs.isFile(skillMd)) continue;
      files.push(skillMd, ...markdownEntries(join(guideDir, "references"), fs));
    }
  }
  return files;
};

// collectCommandMarkdown returns the published Markdown surface of every command
// package so cross-package references do not depend on one special command family.
const collectCommandMarkdown = (repoRoot: string, fs: FileSystem): string[] => {
  const root = join(repoRoot, "packages", "command");
  if (!fs.isDirectory(root)) return [];
  return fs
    .readDirectory(root)
    .toSorted()
    .flatMap((pkg) => {
      const base = join(root, pkg);
      if (!fs.isDirectory(base)) return [];
      return [
        join(base, "README.md"),
        ...markdownEntries(join(base, "docs"), fs),
        ...markdownEntries(join(base, "commands"), fs),
      ].filter((path) => fs.isFile(path));
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

// Blank out fenced blocks and inline spans, keeping line structure, so code
// samples cannot read as markdown links. A C call like `listeners[i](&cs)` and
// a documented link shape like `[name](url)` both match the link pattern.
const withoutCodeSpans = (text: string): string =>
  text
    .replaceAll(/```[\s\S]*?```/g, (block) => block.replaceAll(/[^\n]/g, " "))
    .replaceAll(/`[^`\n]*`/g, (span) => " ".repeat(span.length));

// Validate the boundary-crossing relative links in an explicit set of markdown
// files. Exposed for tests, which supply a synthetic package layout under a
// throwaway repo root.
export const checkMarkdownFilesForCrossPackageLinks = (
  files: readonly string[],
  repoRoot: string,
  report: LinkReport,
  fs: FileSystem = nodeFileSystem,
): LinkCounts => {
  let resolved = 0;
  let broken = 0;
  for (const file of files) {
    const dir = dirname(file);
    const text = withoutCodeSpans(fs.readText(file));
    for (const m of text.matchAll(MD_LINK_RE)) {
      const raw = m[1];
      if (raw === undefined) continue;
      const target = relativeLinkTarget(raw);
      if (target === null) continue;
      const resolvedTarget = resolve(dir, target);
      if (!crossesPackageBoundary(file, resolvedTarget, repoRoot)) continue;
      if (exists(resolvedTarget, fs)) {
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
  fs: FileSystem = nodeFileSystem,
): void => {
  const files = [
    ...collectSkillMarkdown(repoRoot, fs),
    ...collectCommandMarkdown(repoRoot, fs),
  ];
  const {resolved, broken} = checkMarkdownFilesForCrossPackageLinks(
    files,
    repoRoot,
    report,
    fs,
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
    fs,
  );
  if (handoffs.broken === 0 && handoffs.resolved > 0) {
    report.addPass(
      `skill handoffs: ${String(handoffs.resolved)}/${String(handoffs.resolved)} named handoff(s) resolve`,
    );
  }
  checkSkillDependencies(repoRoot, report);
};
