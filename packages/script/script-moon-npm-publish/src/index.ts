#!/usr/bin/env node
import {execSync} from "node:child_process";
import {readFileSync, writeFileSync} from "node:fs";
import {
  logInfo,
  parseCliArgs,
  readPlatformMeta,
  type PackageJson,
} from "@xonovex/script-moon-common";

const {values} = parseCliArgs({
  name: "moon-npm-publish",
  description: "Publish a package to npm if not already published",
  options: {
    "dry-run": {
      type: "boolean",
      short: "d",
      description: "Run npm publish in dry-run mode",
    },
  },
});
const dryRun = values["dry-run"] === true;

const pkgPath = "package.json";
const original = readFileSync(pkgPath, "utf8");
const {name, version} = JSON.parse(original) as PackageJson & {
  name: string;
  version: string;
};

try {
  execSync(`npm view ${name}@${version} version`, {stdio: "ignore"});
  logInfo(`Skipping ${name}@${version} — already published`);
  process.exit(0);
} catch {
  // Version not found, proceed with publish
}

const platformMeta = readPlatformMeta(process.cwd());
if (platformMeta) {
  const pkg = JSON.parse(original) as Record<string, unknown>;
  pkg.os = platformMeta.os;
  pkg.cpu = platformMeta.cpu;
  if (platformMeta.libc) pkg.libc = platformMeta.libc;
  writeFileSync(pkgPath, JSON.stringify(pkg, null, 2) + "\n");
  logInfo(`Injected platform fields for ${name}`);
}

try {
  const publishCmd = dryRun
    ? "npm publish --dry-run --access public"
    : "npm publish --provenance --access public";

  execSync(publishCmd, {stdio: "inherit"});
} finally {
  if (platformMeta) {
    writeFileSync(pkgPath, original);
    logInfo(`Restored original package.json for ${name}`);
  }
}
