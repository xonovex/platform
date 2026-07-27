import {execFileSync} from "node:child_process";
import {readFileSync, writeFileSync} from "node:fs";
import {parseCliArgs} from "@xonovex/script-moon-common/cli-args";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
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
  const {values} = parseCliArgs(cliSpec, argv);
  publishPackage(values["dry-run"] === true, dependencies);
  return 0;
};
