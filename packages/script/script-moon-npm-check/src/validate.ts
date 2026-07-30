import type {PackageJson} from "@xonovex/script-moon-common/package-json";

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
    const value = pkg[field as keyof PackageJson];
    if (value === undefined || !(field in pkg)) {
      errors.push(`Missing required field: ${field}`);
    } else if (
      (typeof value === "string" && value.trim().length === 0) ||
      (Array.isArray(value) && value.length === 0)
    ) {
      errors.push(`Required field is empty: ${field}`);
    }
  }

  if (pkg.repository) {
    if (!pkg.repository.type?.trim()) errors.push("repository.type is missing");
    if (!pkg.repository.url?.trim()) errors.push("repository.url is missing");
  }

  if (!pkg.publishConfig?.access?.trim()) {
    errors.push("publishConfig.access is not set");
  }

  return errors;
};

export const validateDeclaredFiles = (
  pkg: PackageJson,
  exists: (path: string) => boolean,
): readonly string[] =>
  Array.isArray(pkg.files)
    ? pkg.files.flatMap((file) =>
        exists(file) ? [] : [`Declared package file does not exist: "${file}"`],
      )
    : [];
