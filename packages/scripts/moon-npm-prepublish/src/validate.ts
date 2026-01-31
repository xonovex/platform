import type {PackageJson} from "@xonovex/moon-scripts-common";

const REQUIRED_FIELDS = [
  "name",
  "version",
  "license",
  "repository",
  "files",
] as const;

export const validatePackage = (pkg: PackageJson): readonly string[] => {
  const errors: string[] = [];

  for (const field of REQUIRED_FIELDS) {
    if (!(field in pkg) || pkg[field as keyof PackageJson] === undefined) {
      errors.push(`Missing required field: ${field}`);
    }
  }

  if (pkg.repository) {
    if (!pkg.repository.type) errors.push("repository.type is missing");
    if (!pkg.repository.url) errors.push("repository.url is missing");
  }

  if (!pkg.publishConfig?.access) {
    errors.push("publishConfig.access is not set");
  }

  return errors;
};
