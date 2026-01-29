import {execSync} from "node:child_process";
import {existsSync, readFileSync, writeFileSync} from "node:fs";
import {dirname, join, resolve} from "node:path";
import {fileURLToPath} from "node:url";

const __dirname = dirname(fileURLToPath(import.meta.url));
const ROOT_DIR = resolve(__dirname, "..");

interface PackageJson {
  name?: string;
  version?: string;
  private?: boolean;
  dependencies?: Record<string, string>;
  devDependencies?: Record<string, string>;
  peerDependencies?: Record<string, string>;
}

type BumpType = "patch" | "minor" | "major";

const SCOPE = "@xonovex/";

const bumpVersion = (version: string, type: BumpType): string => {
  const [major, minor, patch] = version.split(".").map(Number);
  switch (type) {
    case "major": {
      return `${major + 1}.0.0`;
    }
    case "minor": {
      return `${major}.${minor + 1}.0`;
    }
    case "patch": {
      return `${major}.${minor}.${patch + 1}`;
    }
  }
};

const findAllPackageJsonPaths = (): readonly string[] => {
  // eslint-disable-next-line sonarjs/no-os-command-from-path
  const output = execSync("npx moon query projects", {
    encoding: "utf8",
    maxBuffer: 10 * 1024 * 1024,
    cwd: ROOT_DIR,
  });
  const moonOutput = JSON.parse(output) as {
    projects: readonly {source: string}[];
  };
  return moonOutput.projects
    .map((p) => join(ROOT_DIR, p.source, "package.json"))
    .filter((p) => existsSync(p));
};

const readPkg = (path: string): PackageJson =>
  JSON.parse(readFileSync(path, "utf8")) as PackageJson;

const writePkg = (path: string, pkg: PackageJson): void =>
  writeFileSync(path, JSON.stringify(pkg, null, 2) + "\n");

const updateDependencyVersions = (
  deps: Record<string, string> | undefined,
  name: string,
  newVersion: string,
): boolean => {
  if (!deps || !(name in deps)) return false;
  deps[name] = newVersion;
  return true;
};

const main = (): void => {
  const args = process.argv.slice(2);

  if (args.length < 2 || args.includes("--help") || args.includes("-h")) {
    console.log(`Usage: npx tsx scripts/version-bump.ts <bump-type> <package> [<package>...]

Arguments:
  bump-type   patch | minor | major
  package     Package name (with or without @xonovex/ prefix), or "all"

Examples:
  npx tsx scripts/version-bump.ts patch skills ts-config-base
  npx tsx scripts/version-bump.ts minor @xonovex/core
  npx tsx scripts/version-bump.ts patch all`);
    process.exit(args.includes("--help") || args.includes("-h") ? 0 : 1);
  }

  const bumpType = args[0] as BumpType;
  if (!["patch", "minor", "major"].includes(bumpType)) {
    console.error(
      `Invalid bump type: ${bumpType}. Use patch, minor, or major.`,
    );
    process.exit(1);
  }

  const allPaths = findAllPackageJsonPaths();
  const allPackages = allPaths.map((p) => ({path: p, pkg: readPkg(p)}));
  const nonPrivate = allPackages.filter(
    (
      p,
    ): p is {
      path: string;
      pkg: PackageJson & {name: string; version: string};
    } =>
      typeof p.pkg.name === "string" &&
      !p.pkg.private &&
      typeof p.pkg.version === "string",
  );

  // Resolve target packages
  const requestedNames = args.slice(1);
  const isAll = requestedNames.length === 1 && requestedNames[0] === "all";

  const targets = isAll
    ? nonPrivate
    : nonPrivate.filter((p) => {
        const shortName = p.pkg.name.replace(SCOPE, "");
        return requestedNames.some((r) => r === p.pkg.name || r === shortName);
      });

  if (targets.length === 0) {
    console.error("No matching packages found.");
    console.error("Available:", nonPrivate.map((p) => p.pkg.name).join(", "));
    process.exit(1);
  }

  // Bump versions
  const bumped = new Map<string, string>();
  for (const {path, pkg} of targets) {
    const oldVersion = pkg.version;
    const newVersion = bumpVersion(oldVersion, bumpType);
    pkg.version = newVersion;
    writePkg(path, pkg);
    bumped.set(pkg.name, newVersion);
    console.log(`${pkg.name}: ${oldVersion} -> ${newVersion}`);
  }

  // Update internal dependency references across all packages
  let depsUpdated = 0;
  for (const {path, pkg} of allPackages) {
    let changed = false;
    for (const [name, newVersion] of bumped) {
      if (updateDependencyVersions(pkg.dependencies, name, newVersion))
        changed = true;
      if (updateDependencyVersions(pkg.devDependencies, name, newVersion))
        changed = true;
      if (updateDependencyVersions(pkg.peerDependencies, name, newVersion))
        changed = true;
    }
    if (changed) {
      writePkg(path, pkg);
      depsUpdated++;
    }
  }

  // Update plugin.json if skills was bumped
  const skillsVersion = bumped.get("@xonovex/skills");
  if (skillsVersion) {
    const pluginJsonPath = join(
      ROOT_DIR,
      "packages/plugins/skills/.claude-plugin/plugin.json",
    );
    if (existsSync(pluginJsonPath)) {
      const pluginJson = JSON.parse(
        readFileSync(pluginJsonPath, "utf8"),
      ) as Record<string, unknown>;
      pluginJson.version = skillsVersion;
      writeFileSync(pluginJsonPath, JSON.stringify(pluginJson, null, 2) + "\n");
      console.log(`plugin.json -> ${skillsVersion}`);
    }
  }

  console.log(
    `\nBumped ${bumped.size} package(s), updated deps in ${depsUpdated} file(s).`,
  );
};

main();
