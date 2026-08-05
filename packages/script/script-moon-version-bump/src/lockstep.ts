import {basename, dirname} from "node:path";
import type {PackageJson} from "@xonovex/script-moon-common/package-json";
import {bumpVersion, compareVersions, type BumpType} from "./bump.js";

interface WorkspacePackage {
  readonly path: string;
  readonly pkg: PackageJson;
  readonly headVersion: string | undefined;
}

interface MemberPlan {
  readonly name: string;
  readonly path: string;
  readonly pkg: PackageJson;
  readonly baseVersion: string;
  readonly previousVersion: string;
}

interface DependentPlan {
  readonly name: string;
  readonly path: string;
  readonly pkg: PackageJson;
  readonly referenceCount: number;
  readonly previousVersion: string | undefined;
  readonly newVersion: string | undefined;
}

interface LockstepPlan {
  readonly baseVersion: string;
  readonly targetVersion: string;
  readonly members: readonly MemberPlan[];
  readonly dependents: readonly DependentPlan[];
}

interface LockstepRequest {
  readonly packages: readonly WorkspacePackage[];
  readonly names: readonly string[];
  readonly bumpType: BumpType;
  readonly preid: string | undefined;
  readonly exact: string | undefined;
}

interface ResolvedMember {
  readonly name: string;
  readonly version: string;
  readonly entry: WorkspacePackage;
}

interface FieldRewrite {
  readonly deps: Record<string, string> | undefined;
  readonly changed: number;
}

interface PackageRewrite {
  readonly pkg: PackageJson;
  readonly changed: number;
}

const parseLockstepNames = (raw: string): readonly string[] => {
  const names = raw
    .split(",")
    .map((name) => name.trim())
    .filter(Boolean);
  if (names.length === 0) {
    throw new Error("--lockstep needs at least one package name");
  }
  const duplicate = names.find((name, index) => names.indexOf(name) !== index);
  if (duplicate !== undefined) {
    throw new Error(`duplicate lockstep package ${duplicate}`);
  }
  return names;
};

const resolveMember = (
  packages: readonly WorkspacePackage[],
  request: string,
): ResolvedMember => {
  const matches = packages.filter(
    (entry) =>
      entry.pkg.name === request || basename(dirname(entry.path)) === request,
  );
  const entry = matches[0];
  if (entry === undefined) {
    throw new Error(`unknown lockstep package ${request}`);
  }
  if (matches.length > 1) {
    throw new Error(
      `ambiguous lockstep package ${request}: ${matches.map((match) => match.path).join(", ")}`,
    );
  }
  if (!entry.pkg.name || !entry.pkg.version) {
    throw new Error(`${entry.path} is missing a name or version`);
  }
  return {name: entry.pkg.name, version: entry.pkg.version, entry};
};

const rewriteField = (
  deps: Record<string, string> | undefined,
  versions: ReadonlyMap<string, string>,
): FieldRewrite => {
  if (deps === undefined) return {deps: undefined, changed: 0};
  let changed = 0;
  const next: Record<string, string> = {};
  for (const [name, range] of Object.entries(deps)) {
    const target = versions.get(name);
    if (target !== undefined && target !== range) {
      next[name] = target;
      changed += 1;
      continue;
    }
    next[name] = range;
  }
  return {deps: changed === 0 ? deps : next, changed};
};

const rewriteReferences = (
  pkg: PackageJson,
  versions: ReadonlyMap<string, string>,
): PackageRewrite => {
  const dependencies = rewriteField(pkg.dependencies, versions);
  const devDependencies = rewriteField(pkg.devDependencies, versions);
  const peerDependencies = rewriteField(pkg.peerDependencies, versions);
  const optionalDependencies = rewriteField(pkg.optionalDependencies, versions);
  const changed =
    dependencies.changed +
    devDependencies.changed +
    peerDependencies.changed +
    optionalDependencies.changed;
  if (changed === 0) return {pkg, changed};
  return {
    pkg: {
      ...pkg,
      ...(dependencies.deps === undefined
        ? {}
        : {dependencies: dependencies.deps}),
      ...(devDependencies.deps === undefined
        ? {}
        : {devDependencies: devDependencies.deps}),
      ...(peerDependencies.deps === undefined
        ? {}
        : {peerDependencies: peerDependencies.deps}),
      ...(optionalDependencies.deps === undefined
        ? {}
        : {optionalDependencies: optionalDependencies.deps}),
    },
    changed,
  };
};

const baselineOf = (member: ResolvedMember): string =>
  member.entry.headVersion ?? member.version;

// The release baseline is the highest committed version in the set, so a member
// a peer already patch-bumped in the worktree cannot drag the line backwards.
const selectBaseVersion = (members: readonly ResolvedMember[]): string => {
  const [first, ...rest] = members;
  if (first === undefined) {
    throw new Error("a lockstep set needs at least one package");
  }
  return rest.reduce(
    (highest, member) =>
      compareVersions(baselineOf(member), highest) > 0
        ? baselineOf(member)
        : highest,
    baselineOf(first),
  );
};

const planDependent = (
  entry: WorkspacePackage,
  versions: ReadonlyMap<string, string>,
): DependentPlan | undefined => {
  const rewritten = rewriteReferences(entry.pkg, versions);
  if (rewritten.changed === 0) return undefined;
  const currentVersion = entry.pkg.version;
  const newVersion =
    currentVersion !== undefined &&
    entry.pkg.private !== true &&
    entry.headVersion === currentVersion
      ? bumpVersion(currentVersion, "patch")
      : undefined;
  return {
    name: entry.pkg.name ?? entry.path,
    path: entry.path,
    pkg:
      newVersion === undefined
        ? rewritten.pkg
        : {...rewritten.pkg, version: newVersion},
    referenceCount: rewritten.changed,
    previousVersion: newVersion === undefined ? undefined : currentVersion,
    newVersion,
  };
};

const planLockstep = (request: LockstepRequest): LockstepPlan => {
  const members = request.names.map((name) =>
    resolveMember(request.packages, name),
  );
  const baseVersion = selectBaseVersion(members);
  const targetVersion =
    request.exact ?? bumpVersion(baseVersion, request.bumpType, request.preid);
  const versions = new Map(
    members.map((member) => [member.name, targetVersion]),
  );
  const memberPaths = new Set(members.map((member) => member.entry.path));

  const memberPlans = members.map((member) => ({
    name: member.name,
    path: member.entry.path,
    pkg: {
      ...rewriteReferences(member.entry.pkg, versions).pkg,
      version: targetVersion,
    },
    baseVersion: baselineOf(member),
    previousVersion: member.version,
  }));

  const dependentPlans = request.packages
    .filter((entry) => !memberPaths.has(entry.path))
    .map((entry) => planDependent(entry, versions))
    .filter((plan) => plan !== undefined);

  return {
    baseVersion,
    targetVersion,
    members: memberPlans,
    dependents: dependentPlans,
  };
};

export {parseLockstepNames, planLockstep};
export type {
  DependentPlan,
  LockstepPlan,
  LockstepRequest,
  MemberPlan,
  WorkspacePackage,
};
