export type BumpType = "patch" | "minor" | "major";

export const bumpVersion = (
  version: string,
  type: BumpType,
  preid?: string,
): string => {
  const [corePart, ...prereleaseParts] = version.split("-");
  const parts = (corePart ?? "").split(".").map(Number);
  const major = parts[0] ?? 0;
  const minor = parts[1] ?? 0;
  const patch = parts[2] ?? 0;
  const prereleaseStr = prereleaseParts.join("-");

  if (preid) {
    const expectedPrefix = `${preid}.`;
    if (type === "patch" && prereleaseStr.startsWith(expectedPrefix)) {
      const preNum = Number(prereleaseStr.slice(expectedPrefix.length));
      return `${String(major)}.${String(minor)}.${String(patch)}-${preid}.${String(preNum + 1)}`;
    }
    const bumped = bumpVersion(corePart ?? "", type);
    return `${bumped}-${preid}.0`;
  }

  if (prereleaseStr) {
    return `${String(major)}.${String(minor)}.${String(patch + 1)}`;
  }

  switch (type) {
    case "major": {
      return `${String(major + 1)}.0.0`;
    }
    case "minor": {
      return `${String(major)}.${String(minor + 1)}.0`;
    }
    case "patch": {
      return `${String(major)}.${String(minor)}.${String(patch + 1)}`;
    }
  }
};

interface ParsedVersion {
  readonly core: readonly number[];
  readonly prerelease: readonly string[];
}

const parseVersion = (version: string): ParsedVersion => {
  const [corePart, ...prereleaseParts] = version.split("-");
  return {
    core: (corePart ?? "").split(".").map(Number),
    prerelease: prereleaseParts.join("-").split(".").filter(Boolean),
  };
};

const compareIdentifiers = (left: string, right: string): number => {
  const leftNumeric = /^\d+$/.test(left);
  const rightNumeric = /^\d+$/.test(right);
  if (leftNumeric && rightNumeric) return Number(left) - Number(right);
  if (leftNumeric) return -1;
  if (rightNumeric) return 1;
  if (left < right) return -1;
  return left > right ? 1 : 0;
};

const comparePrerelease = (
  left: readonly string[],
  right: readonly string[],
): number => {
  if (left.length === 0 && right.length === 0) return 0;
  // A release outranks any prerelease that shares its core version.
  if (left.length === 0) return 1;
  if (right.length === 0) return -1;
  for (let index = 0; index < Math.max(left.length, right.length); index += 1) {
    const leftId = left[index];
    const rightId = right[index];
    if (leftId === undefined) return -1;
    if (rightId === undefined) return 1;
    const difference = compareIdentifiers(leftId, rightId);
    if (difference !== 0) return difference;
  }
  return 0;
};

export const compareVersions = (left: string, right: string): number => {
  const parsedLeft = parseVersion(left);
  const parsedRight = parseVersion(right);
  for (let index = 0; index < 3; index += 1) {
    const difference =
      (parsedLeft.core[index] ?? 0) - (parsedRight.core[index] ?? 0);
    if (difference !== 0) return difference;
  }
  return comparePrerelease(parsedLeft.prerelease, parsedRight.prerelease);
};

export const updateDependencyVersions = (
  deps: Record<string, string> | undefined,
  name: string,
  newVersion: string,
): boolean => {
  if (!deps || !(name in deps)) return false;
  deps[name] = newVersion;
  return true;
};
