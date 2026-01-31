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
    if (prereleaseStr.startsWith(expectedPrefix) && type === "patch") {
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

export const updateDependencyVersions = (
  deps: Record<string, string> | undefined,
  name: string,
  newVersion: string,
): boolean => {
  if (!deps || !(name in deps)) return false;
  deps[name] = newVersion;
  return true;
};
