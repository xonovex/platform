import {existsSync, readdirSync, readFileSync, statSync} from "node:fs";
import {dirname, join, resolve} from "node:path";
import {fileURLToPath} from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const ROOT_DIR = resolve(__dirname, "..");

interface PackageJson {
  name?: string;
  version?: string;
  license?: string;
  repository?: {type?: string; url?: string; directory?: string};
  files?: readonly string[];
  publishConfig?: {access?: string; registry?: string};
  private?: boolean;
}

interface ValidationError {
  package: string;
  errors: readonly string[];
}

const REQUIRED_FIELDS = [
  "name",
  "version",
  "license",
  "repository",
  "files",
] as const;

const findPackageJsonFiles = (dir: string): readonly string[] => {
  const packages: string[] = [];
  const configDir = join(dir, "packages", "config");
  const cliDir = join(dir, "packages", "cli");
  const libDir = join(dir, "packages", "lib");
  const pluginsDir = join(dir, "packages", "plugins");

  const addPackagesFromDir = (searchDir: string): void => {
    if (!existsSync(searchDir)) return;

    for (const entry of readdirSync(searchDir)) {
      const entryPath = join(searchDir, entry);
      if (!statSync(entryPath).isDirectory()) continue;

      const packageJsonPath = join(entryPath, "package.json");
      if (existsSync(packageJsonPath)) {
        packages.push(packageJsonPath);
      }

      // Handle nested platform packages
      if (entry === "agent-cli-go-platforms") {
        for (const platform of readdirSync(entryPath)) {
          const platformPath = join(entryPath, platform);
          if (!statSync(platformPath).isDirectory()) continue;
          const platformPackageJson = join(platformPath, "package.json");
          if (existsSync(platformPackageJson)) {
            packages.push(platformPackageJson);
          }
        }
      }
    }
  };

  addPackagesFromDir(configDir);
  addPackagesFromDir(cliDir);
  addPackagesFromDir(libDir);
  addPackagesFromDir(pluginsDir);

  return packages;
};

const validatePackage = (
  packageJsonPath: string,
): ValidationError | undefined => {
  const errors: string[] = [];
  const packageDir = dirname(packageJsonPath);
  const relativePath = packageJsonPath.replace(ROOT_DIR + "/", "");

  let pkg: PackageJson;
  try {
    pkg = JSON.parse(readFileSync(packageJsonPath, "utf8")) as PackageJson;
  } catch (error) {
    return {
      package: relativePath,
      errors: [`Failed to parse package.json: ${error}`],
    };
  }

  // Skip private packages
  if (pkg.private) {
    return undefined;
  }

  // Check required fields
  for (const field of REQUIRED_FIELDS) {
    if (!(field in pkg) || pkg[field as keyof PackageJson] === undefined) {
      errors.push(`Missing required field: ${field}`);
    }
  }

  // Check repository has required subfields
  if (pkg.repository) {
    if (!pkg.repository.type) errors.push("repository.type is missing");
    if (!pkg.repository.url) errors.push("repository.url is missing");
  }

  // Check publishConfig.access
  if (!pkg.publishConfig?.access) {
    errors.push("publishConfig.access is not set");
  }

  // Check files in files array exist
  if (pkg.files && Array.isArray(pkg.files)) {
    for (const file of pkg.files) {
      const filePath = join(packageDir, file);
      if (!existsSync(filePath)) {
        // Only warn, don't error - files may be created during build
        console.warn(
          `  Warning: ${relativePath}: file "${file}" does not exist yet (will be created during build)`,
        );
      }
    }
  }

  if (errors.length === 0) {
    return undefined;
  }

  return {package: relativePath, errors};
};

const main = (): void => {
  console.log("Checking packages for npm publishing readiness...\n");

  const packageJsonFiles = findPackageJsonFiles(ROOT_DIR);
  const validationErrors: ValidationError[] = [];

  for (const packageJsonPath of packageJsonFiles) {
    const result = validatePackage(packageJsonPath);
    if (result) {
      validationErrors.push(result);
    }
  }

  console.log(`\nChecked ${packageJsonFiles.length} packages.`);

  if (validationErrors.length > 0) {
    console.error("\nValidation errors found:\n");
    for (const {package: pkg, errors} of validationErrors) {
      console.error(`  ${pkg}:`);
      for (const error of errors) {
        console.error(`    - ${error}`);
      }
    }
    process.exit(1);
  }

  console.log("\nAll packages are ready for publishing!");
};

main();
