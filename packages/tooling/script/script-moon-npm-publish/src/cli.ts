import {execFileSync} from "node:child_process";
import {readFileSync, writeFileSync} from "node:fs";
import {resolveExecutable} from "@xonovex/script-moon-common/adapters/executable-node";
import {parseCliArgs} from "@xonovex/script-moon-common/cli-args";
import {logInfo} from "@xonovex/script-moon-common/logging";
import {readPlatformMeta} from "@xonovex/script-moon-common/platform-meta";
import {
  isPublished,
  publishArgs,
  publishPackage,
  type PublishDependencies,
} from "./publish.js";

export const defaultDependencies: PublishDependencies = {
  readPackageJson: () => readFileSync("package.json", "utf8"),
  writePackageJson: (contents) => {
    writeFileSync("package.json", contents);
  },
  currentDirectory: () => process.cwd(),
  readPlatformMeta,
  isPublished,
  publish: (dryRun) => {
    execFileSync(resolveExecutable("npm"), [...publishArgs(dryRun)], {
      stdio: "inherit",
    });
  },
  log: logInfo,
};

const cliSpec = {
  name: "moon-npm-publish",
  description: "Publish a package to npm if not already published",
  options: {
    "dry-run": {
      type: "boolean" as const,
      short: "d",
      description: "Run npm publish in dry-run mode",
    },
  },
};

export const main = (
  argv: readonly string[] = process.argv.slice(2),
  dependencies: PublishDependencies = defaultDependencies,
): number => {
  const args = parseCliArgs(cliSpec, argv);
  if (args.kind === "help") {
    console.log(args.text);
    return 0;
  }
  const {values} = args;
  publishPackage(values["dry-run"] === true, dependencies);
  return 0;
};
