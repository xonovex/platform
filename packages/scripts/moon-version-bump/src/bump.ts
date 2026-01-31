export type BumpType = "patch" | "minor" | "major";

export const bumpVersion = (version: string, type: BumpType): string => {
  const parts = version.split(".").map(Number);
  const major = parts[0] ?? 0;
  const minor = parts[1] ?? 0;
  const patch = parts[2] ?? 0;
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
