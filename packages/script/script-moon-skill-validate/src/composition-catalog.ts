import {createHash} from "node:crypto";
import {readdirSync, readFileSync} from "node:fs";
import {join, relative} from "node:path";
import {isDirectory, isFile} from "@xonovex/script-moon-common/fs";
import {major, satisfies, valid, validRange} from "semver";
import {z} from "zod";
import {type LinkReport} from "./reference-file-links.js";

const GUIDE_NAME_RE = /^[a-z0-9]+(?:-[a-z0-9]+)*-guide$/;
const PROVISION_ID_RE = /^[a-z][a-z0-9-]*:[a-z0-9]+(?:[.-][a-z0-9]+)*$/;
const SUPPORTED_CONTRACT_MAJOR = 1;

const LifecycleSchema = z.enum(["durable", "procedural"]);
const FunctionalRoleSchema = z.enum([
  "domain",
  "context",
  "preference",
  "procedure",
  "capability-use",
  "assurance",
  "recovery",
  "communication",
]);
const ProvisionSchema = z
  .object({
    id: z.string().regex(PROVISION_ID_RE),
    version: z.string().refine((value) => valid(value) !== null, {
      message: "must be a valid semantic version",
    }),
  })
  .strict();
const RequirementSchema = z
  .object({
    id: z.string().regex(PROVISION_ID_RE),
    range: z.string().refine((value) => validRange(value) !== null, {
      message: "must be a valid semantic-version range",
    }),
    strength: z.enum(["required", "preferred"]),
    reason: z.string().trim().min(1),
  })
  .strict();
const CatalogEntrySchema = z
  .object({
    name: z.string().regex(GUIDE_NAME_RE),
    classification: z
      .object({
        lifecycle: LifecycleSchema,
        functionalRole: FunctionalRoleSchema,
      })
      .strict(),
    provisions: z.array(ProvisionSchema).default([]),
    requirements: z.array(RequirementSchema).default([]),
  })
  .strict();
const CatalogFileSchema = z
  .object({
    contractVersion: z.string().refine((value) => valid(value) !== null, {
      message: "must be a valid semantic version",
    }),
    skills: z.array(CatalogEntrySchema),
  })
  .strict();

export type Lifecycle = z.infer<typeof LifecycleSchema>;
export type FunctionalRole = z.infer<typeof FunctionalRoleSchema>;
export type SemanticProvision = z.infer<typeof ProvisionSchema>;
export type SemanticRequirement = z.infer<typeof RequirementSchema>;
export type CompositionCatalogEntry = z.infer<typeof CatalogEntrySchema>;

export interface CompositionCatalog {
  readonly contractVersion: string;
  readonly digest: string;
  readonly skills: readonly CompositionCatalogEntry[];
}

export interface InstalledSkill {
  readonly guide: string;
  readonly implementationVersion: string;
  readonly packagePath: string;
  readonly plugin: string;
  readonly sourcesPath: string;
}

export interface SelectionProvenance {
  readonly kind:
    "explicit" | "exact-dependency" | "semantic-requirement" | "policy";
  readonly reason: string;
  readonly requestedBy?: string;
}

export interface SelectedSkill {
  readonly catalogContractVersion: string;
  readonly catalogDigest: string;
  readonly guide: string;
  readonly implementationVersion: string;
  readonly packagePath: string;
  readonly plugin: string;
  readonly provenance: SelectionProvenance;
  readonly provision?: SemanticProvision;
  readonly requestedRange?: string;
  readonly sourcesPath: string;
}

export type SelectionResult =
  | {
      readonly status: "selected";
      readonly selection: SelectedSkill;
    }
  | {
      readonly status: "ambiguous" | "incompatible" | "unavailable";
      readonly blocking: boolean;
      readonly candidates: readonly string[];
      readonly message: string;
    };

type ParseResult =
  | {readonly success: true; readonly data: CompositionCatalog}
  | {readonly success: false; readonly errors: readonly string[]};

const issueDetail = (issue: z.core.$ZodIssue): string => {
  const path =
    issue.path.length > 0 ? issue.path.map(String).join(".") : "root";
  return `${path}: ${issue.message}`;
};

export const parseCompositionCatalog = (
  input: unknown,
  sourceText = JSON.stringify(input),
): ParseResult => {
  const parsed = CatalogFileSchema.safeParse(input);
  if (!parsed.success) {
    return {
      success: false,
      errors: parsed.error.issues.map(issueDetail),
    };
  }
  if (major(parsed.data.contractVersion) !== SUPPORTED_CONTRACT_MAJOR) {
    return {
      success: false,
      errors: [
        `contractVersion: unsupported major ${String(major(parsed.data.contractVersion))}; expected ${String(SUPPORTED_CONTRACT_MAJOR)}`,
      ],
    };
  }
  return {
    success: true,
    data: {
      ...parsed.data,
      digest: createHash("sha256").update(sourceText).digest("hex"),
    },
  };
};

const selectedSkill = (
  catalog: CompositionCatalog,
  installed: InstalledSkill,
  provenance: SelectionProvenance,
  provision?: SemanticProvision,
  requestedRange?: string,
): SelectionResult => ({
  status: "selected",
  selection: {
    catalogContractVersion: catalog.contractVersion,
    catalogDigest: catalog.digest,
    guide: installed.guide,
    implementationVersion: installed.implementationVersion,
    packagePath: installed.packagePath,
    plugin: installed.plugin,
    provenance,
    ...(provision === undefined ? {} : {provision}),
    ...(requestedRange === undefined ? {} : {requestedRange}),
    sourcesPath: installed.sourcesPath,
  },
});

export const resolveExactSkill = (
  catalog: CompositionCatalog,
  installedSkills: readonly InstalledSkill[],
  guide: string,
  implementationVersion: string | undefined,
  provenance: SelectionProvenance,
): SelectionResult => {
  if (catalog.skills.every((entry) => entry.name !== guide)) {
    return {
      status: "unavailable",
      blocking: true,
      candidates: [],
      message: `exact skill ${guide} is not declared in catalog ${catalog.contractVersion}`,
    };
  }
  const installed = installedSkills.find((skill) => skill.guide === guide);
  if (installed === undefined) {
    return {
      status: "unavailable",
      blocking: true,
      candidates: [],
      message: `exact skill ${guide} is not installed`,
    };
  }
  if (
    implementationVersion !== undefined &&
    installed.implementationVersion !== implementationVersion
  ) {
    return {
      status: "incompatible",
      blocking: true,
      candidates: [`${guide}@${installed.implementationVersion}`],
      message: `exact skill ${guide}@${implementationVersion} does not match installed ${installed.implementationVersion}`,
    };
  }
  return selectedSkill(catalog, installed, provenance);
};

export const resolveSemanticRequirement = (
  catalog: CompositionCatalog,
  installedSkills: readonly InstalledSkill[],
  requirement: SemanticRequirement,
  provenance: SelectionProvenance,
): SelectionResult => {
  const installedByGuide = new Map(
    installedSkills.map((skill) => [skill.guide, skill]),
  );
  const declared = catalog.skills.flatMap((entry) =>
    entry.provisions
      .filter((provision) => provision.id === requirement.id)
      .map((provision) => ({entry, provision})),
  );
  const installed = declared.filter(({entry}) =>
    installedByGuide.has(entry.name),
  );
  const compatible = installed.filter(({provision}) =>
    satisfies(provision.version, requirement.range),
  );
  const blocking = requirement.strength === "required";

  if (compatible.length === 0) {
    const candidates = installed.map(
      ({entry, provision}) => `${entry.name}@${provision.version}`,
    );
    return {
      status: candidates.length > 0 ? "incompatible" : "unavailable",
      blocking,
      candidates,
      message:
        candidates.length > 0
          ? `${requirement.id} has no installed provision compatible with ${requirement.range}`
          : `${requirement.id} has no installed provider`,
    };
  }
  if (compatible.length > 1) {
    return {
      status: "ambiguous",
      blocking,
      candidates: compatible
        .map(({entry, provision}) => `${entry.name}@${provision.version}`)
        .toSorted(),
      message: `${requirement.id}@${requirement.range} has multiple compatible installed providers`,
    };
  }
  const candidate = compatible[0];
  if (candidate === undefined) {
    throw new Error("compatible semantic selection disappeared");
  }
  const installedCandidate = installedByGuide.get(candidate.entry.name);
  if (installedCandidate === undefined) {
    throw new Error("installed semantic selection disappeared");
  }
  return selectedSkill(
    catalog,
    installedCandidate,
    provenance,
    candidate.provision,
    requirement.range,
  );
};

const dependencyCycles = (
  edges: ReadonlyMap<string, ReadonlySet<string>>,
): readonly string[] => {
  const states = new Map<string, "visiting" | "visited">();
  const cycles = new Set<string>();
  const visit = (name: string, path: readonly string[]): void => {
    if (states.get(name) === "visited") return;
    if (states.get(name) === "visiting") {
      const start = path.indexOf(name);
      cycles.add([...path.slice(start), name].join(" → "));
      return;
    }
    states.set(name, "visiting");
    for (const dependency of edges.get(name) ?? []) {
      visit(dependency, [...path, name]);
    }
    states.set(name, "visited");
  };
  for (const name of edges.keys()) visit(name, []);
  return [...cycles].toSorted();
};

export const compositionCatalogErrors = (
  catalog: CompositionCatalog,
  installedSkills: readonly InstalledSkill[],
): readonly string[] => {
  const errors: string[] = [];
  const names = catalog.skills.map(({name}) => name);
  const duplicateNames = names.filter(
    (name, index) => names.indexOf(name) !== index,
  );
  for (const name of new Set(duplicateNames)) {
    errors.push(`duplicate skill entry ${name}`);
  }
  if (
    names.some((name, index) => index > 0 && name < (names[index - 1] ?? ""))
  ) {
    errors.push("skill entries must be sorted alphabetically by name");
  }

  const installedNames = new Set(installedSkills.map(({guide}) => guide));
  const catalogNames = new Set(names);
  for (const missing of [...installedNames]
    .filter((name) => !catalogNames.has(name))
    .toSorted()) {
    errors.push(`missing installed skill ${missing}`);
  }
  for (const unexpected of [...catalogNames]
    .filter((name) => !installedNames.has(name))
    .toSorted()) {
    errors.push(`catalog skill ${unexpected} is not installed`);
  }

  const edges = new Map<string, Set<string>>();
  for (const entry of catalog.skills) {
    const provisionIds = entry.provisions.map(({id}) => id);
    for (const duplicate of new Set(
      provisionIds.filter((id, index) => provisionIds.indexOf(id) !== index),
    )) {
      errors.push(
        `${entry.name} declares provision ${duplicate} more than once`,
      );
    }
    const requirementKeys = entry.requirements.map(
      ({id, range, strength}) => `${id}\0${range}\0${strength}`,
    );
    for (const duplicate of new Set(
      requirementKeys.filter(
        (key, index) => requirementKeys.indexOf(key) !== index,
      ),
    )) {
      errors.push(
        `${entry.name} declares requirement ${duplicate.replaceAll("\0", " ")} more than once`,
      );
    }

    for (const requirement of entry.requirements) {
      const resolution = resolveSemanticRequirement(
        catalog,
        installedSkills,
        requirement,
        {
          kind: "semantic-requirement",
          reason: requirement.reason,
          requestedBy: entry.name,
        },
      );
      if (requirement.strength !== "required") continue;
      if (resolution.status !== "selected") {
        errors.push(
          `${entry.name} required requirement is not deterministic: ${resolution.message}`,
        );
        continue;
      }
      const dependencies = edges.get(entry.name) ?? new Set<string>();
      dependencies.add(resolution.selection.guide);
      edges.set(entry.name, dependencies);
    }
  }
  for (const cycle of dependencyCycles(edges)) {
    errors.push(`required semantic dependency cycle ${cycle}`);
  }
  return errors;
};

export const compositionCatalogSnapshotErrors = (
  canonical: string,
  snapshot: string | undefined,
): readonly string[] => {
  if (snapshot === undefined) {
    return [
      "packaged workflow snapshot workflow-guide/assets/composition-catalog.json is missing",
    ];
  }
  if (snapshot !== canonical) {
    return [
      "packaged workflow snapshot differs from packages/skill/composition-catalog.json",
    ];
  }
  return [];
};

const installedSkillInventory = (
  repoRoot: string,
  report: LinkReport,
): {readonly skills: readonly InstalledSkill[]; readonly valid: boolean} => {
  const root = join(repoRoot, "packages", "skill");
  if (!isDirectory(root)) return {skills: [], valid: false};
  const installed: InstalledSkill[] = [];
  let inventoryValid = true;
  for (const packageName of readdirSync(root).toSorted()) {
    const packagePath = join(root, packageName);
    if (!packageName.startsWith("skill-") || !isDirectory(packagePath)) {
      continue;
    }
    const packageJsonPath = join(packagePath, "package.json");
    if (!isFile(packageJsonPath)) continue;
    let packageJson: unknown;
    try {
      packageJson = JSON.parse(
        readFileSync(packageJsonPath, "utf8"),
      ) as unknown;
    } catch (error) {
      inventoryValid = false;
      report.addFail(
        `composition catalog: invalid JSON in ${relative(repoRoot, packageJsonPath)}: ${error instanceof Error ? error.message : String(error)}`,
      );
      continue;
    }
    if (
      typeof packageJson !== "object" ||
      packageJson === null ||
      !("version" in packageJson) ||
      typeof packageJson.version !== "string" ||
      valid(packageJson.version) === null
    ) {
      inventoryValid = false;
      report.addFail(
        `composition catalog: ${relative(repoRoot, packageJsonPath)} needs a semantic version`,
      );
      continue;
    }
    for (const guide of readdirSync(packagePath).toSorted()) {
      const guidePath = join(packagePath, guide);
      if (!isFile(join(guidePath, "SKILL.md"))) continue;
      const sourcesPath = join(guidePath, "SOURCES.md");
      if (!isFile(sourcesPath)) {
        inventoryValid = false;
        report.addFail(
          `composition catalog: ${relative(repoRoot, sourcesPath)} is missing selection provenance`,
        );
      }
      installed.push({
        guide,
        implementationVersion: packageJson.version,
        packagePath: relative(repoRoot, packagePath),
        plugin: `xonovex-${packageName}`,
        sourcesPath: relative(repoRoot, sourcesPath),
      });
    }
  }
  return {skills: installed, valid: inventoryValid};
};

export const checkCompositionCatalog = (
  repoRoot: string,
  report: LinkReport,
): void => {
  const path = join(repoRoot, "packages", "skill", "composition-catalog.json");
  if (!isFile(path)) {
    report.addFail(
      "composition catalog: packages/skill/composition-catalog.json is missing",
    );
    return;
  }
  const sourceText = readFileSync(path, "utf8");
  let input: unknown;
  try {
    input = JSON.parse(sourceText) as unknown;
  } catch (error) {
    report.addFail(
      `composition catalog: invalid JSON: ${error instanceof Error ? error.message : String(error)}`,
    );
    return;
  }
  const parsed = parseCompositionCatalog(input, sourceText);
  if (!parsed.success) {
    for (const error of parsed.errors) {
      report.addFail(`composition catalog: ${error}`);
    }
    return;
  }
  const snapshotPath = join(
    repoRoot,
    "packages",
    "skill",
    "skill-workflow",
    "workflow-guide",
    "assets",
    "composition-catalog.json",
  );
  const snapshotText = isFile(snapshotPath)
    ? readFileSync(snapshotPath, "utf8")
    : undefined;
  const snapshotErrors = compositionCatalogSnapshotErrors(
    sourceText,
    snapshotText,
  );
  for (const error of snapshotErrors) {
    report.addFail(`composition catalog: ${error}`);
  }
  const installed = installedSkillInventory(repoRoot, report);
  const errors = compositionCatalogErrors(parsed.data, installed.skills);
  for (const error of errors) {
    report.addFail(`composition catalog: ${error}`);
  }
  if (
    errors.length === 0 &&
    snapshotErrors.length === 0 &&
    installed.valid &&
    installed.skills.length > 0
  ) {
    report.addPass(
      `composition catalog: ${String(installed.skills.length)} installed skill(s) have one validated classification`,
    );
  }
};
