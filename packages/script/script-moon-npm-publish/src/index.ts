#!/usr/bin/env node
import {execFileSync} from "node:child_process";
import {readFileSync, writeFileSync} from "node:fs";
import {
  logInfo,
  parseCliArgs,
  readPlatformMeta,
} from "@xonovex/script-moon-common";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {isPublished, parsePackageIdentity, publishArgs} from "./publish.js";

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
const identity = parsePackageIdentity(JSON.parse(original) as unknown);
const {name, version} = identity;

if (isPublished(identity)) {
  logInfo(`Skipping ${name}@${version} — already published`);
} else {
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
    execFileSync(resolveExecutable("npm"), [...publishArgs(dryRun)], {
      stdio: "inherit",
    });
  } finally {
    if (platformMeta) {
      writeFileSync(pkgPath, original);
      logInfo(`Restored original package.json for ${name}`);
    }
  }
}
