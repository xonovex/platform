import {basename, dirname} from "node:path";
import type {PackageJson} from "@xonovex/script-moon-common/package-json";
import {bumpVersion, compareVersions, type BumpType} from "./bump.js";
import type {ReleasePredicate} from "./release.js";

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
  /** Whether the version moves with a release line of more than one package. */
  readonly withReleaseLine: boolean;
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
  readonly isReleased: ReleasePredicate;
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

// Two public packages share a release line when one pins the other exactly in
// optionalDependencies and both hold that version, as a launcher pins its
// platform packages.
const pinsInLine = (
  holder: WorkspacePackage,
  held: WorkspacePackage,
): boolean =>
  holder.pkg.private !== true &&
  held.pkg.private !== true &&
  held.pkg.name !== undefined &&
  held.pkg.version !== undefined &&
  holder.pkg.version === held.pkg.version &&
  holder.pkg.optionalDependencies?.[held.pkg.name] === held.pkg.version;

/** The release line of entry: entry first, then every package it reaches. */
const releaseLineOf = (
  packages: readonly WorkspacePackage[],
  entry: WorkspacePackage,
): readonly WorkspacePackage[] => {
  const grow = (
    line: readonly WorkspacePackage[],
  ): readonly WorkspacePackage[] => {
    const reached = packages.filter(
      (other) =>
        line.every((member) => member.path !== other.path) &&
        line.some(
          (member) => pinsInLine(member, other) || pinsInLine(other, member),
        ),
    );
    return reached.length === 0 ? line : grow([...line, ...reached]);
  };
  return grow([entry]);
};

// A dependent earns a version of its own only when it is publishable, the
// worktree has not already bumped it away from its committed version, and its
// release line was released: a never-released line ships the rewritten
// reference in its first release at its current version. Every package of a
// line holds the same version, so the whole line moves one patch past it and a
// launcher never leaves its platform packages behind.
const lineVersionOf = (
  entry: WorkspacePackage,
  line: readonly WorkspacePackage[],
  isReleased: ReleasePredicate,
): string | undefined => {
  const currentVersion = entry.pkg.version;
  if (
    currentVersion === undefined ||
    entry.pkg.private === true ||
    entry.headVersion !== currentVersion
  )
    return undefined;
  const released = line.some(
    (member) =>
      member.pkg.version !== undefined &&
      isReleased(member.path, member.pkg.version),
  );
  return released ? bumpVersion(currentVersion, "patch") : undefined;
};

// A patch-bumped dependent is itself held by exact references elsewhere, so the
// version map is driven to a fixed point before any package is planned. Each
// pass that changes anything adds a name, which bounds the pass count.
const resolveVersions = (
  packages: readonly WorkspacePackage[],
  memberPaths: ReadonlySet<string>,
  memberVersions: ReadonlyMap<string, string>,
  isReleased: ReleasePredicate,
): ReadonlyMap<string, string> => {
  const resolved = new Map(memberVersions);
  let settled = false;
  let remainingPasses = packages.length;
  while (!settled && remainingPasses > 0) {
    remainingPasses -= 1;
    settled = true;
    for (const entry of packages) {
      const name = entry.pkg.name;
      if (
        name === undefined ||
        memberPaths.has(entry.path) ||
        resolved.has(name) ||
        rewriteReferences(entry.pkg, resolved).changed === 0
      )
        continue;
      const line = releaseLineOf(packages, entry);
      const newVersion = lineVersionOf(entry, line, isReleased);
      if (newVersion === undefined) continue;
      for (const member of line) {
        if (member.pkg.name !== undefined) {
          resolved.set(member.pkg.name, newVersion);
        }
      }
      settled = false;
    }
  }
  return resolved;
};

const planDependent = (
  packages: readonly WorkspacePackage[],
  entry: WorkspacePackage,
  versions: ReadonlyMap<string, string>,
): DependentPlan | undefined => {
  const rewritten = rewriteReferences(entry.pkg, versions);
  const newVersion =
    entry.pkg.name === undefined ? undefined : versions.get(entry.pkg.name);
  if (newVersion === undefined && rewritten.changed === 0) return undefined;
  return {
    name: entry.pkg.name ?? entry.path,
    path: entry.path,
    pkg:
      newVersion === undefined
        ? rewritten.pkg
        : {...rewritten.pkg, version: newVersion},
    referenceCount: rewritten.changed,
    previousVersion: newVersion === undefined ? undefined : entry.pkg.version,
    newVersion,
    withReleaseLine:
      newVersion !== undefined && releaseLineOf(packages, entry).length > 1,
  };
};

// A release line moves as one, so a set that names part of a line would leave
// the rest of it on the old version.
const assertWholeLines = (
  packages: readonly WorkspacePackage[],
  members: readonly ResolvedMember[],
  memberPaths: ReadonlySet<string>,
): void => {
  for (const member of members) {
    const missing = releaseLineOf(packages, member.entry).filter(
      (entry) => !memberPaths.has(entry.path),
    );
    if (missing.length > 0) {
      throw new Error(
        `the lockstep set splits the release line of ${member.name}; add ${missing.map((entry) => entry.pkg.name ?? entry.path).join(", ")}`,
      );
    }
  }
};

const planLockstep = (request: LockstepRequest): LockstepPlan => {
  const members = request.names.map((name) =>
    resolveMember(request.packages, name),
  );
  const baseVersion = selectBaseVersion(members);
  const targetVersion =
    request.exact ?? bumpVersion(baseVersion, request.bumpType, request.preid);
  const memberPaths = new Set(members.map((member) => member.entry.path));
  assertWholeLines(request.packages, members, memberPaths);
  const versions = resolveVersions(
    request.packages,
    memberPaths,
    new Map(members.map((member) => [member.name, targetVersion])),
    request.isReleased,
  );

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
    .map((entry) => planDependent(request.packages, entry, versions))
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
