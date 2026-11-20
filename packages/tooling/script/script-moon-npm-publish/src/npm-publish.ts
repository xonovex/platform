import {join} from "node:path";
import {parseCliArgs} from "@xonovex/script-moon-common/cli-args";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {infoLine} from "@xonovex/script-moon-common/logging";
import type {Output} from "@xonovex/script-moon-common/output";
import {readPlatformMeta} from "@xonovex/script-moon-common/platform-meta";
import {
  publishPackage,
  type PackageIdentity,
  type PublishDependencies,
} from "./publish.js";

export interface NpmPublishDependencies {
  readonly fs: FileSystem;
  readonly output: Output;
  readonly isPublished: (identity: PackageIdentity) => boolean;
  readonly publish: (dryRun: boolean) => void;
}

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

/** Publishes the package in packageRoot unless npm already has its version. */
export const main = (
  argv: readonly string[],
  packageRoot: string,
  dependencies: NpmPublishDependencies,
): number => {
  const args = parseCliArgs(cliSpec, argv);
  if (args.kind === "help") {
    dependencies.output.stdout(args.text);
    return 0;
  }
  const {fs, output} = dependencies;
  const packageJsonPath = join(packageRoot, "package.json");
  const publishDependencies: PublishDependencies = {
    readPackageJson: () => fs.readText(packageJsonPath),
    writePackageJson: (contents) => {
      fs.writeFile(packageJsonPath, contents);
    },
    currentDirectory: () => packageRoot,
    readPlatformMeta: (directory) => readPlatformMeta(directory, fs),
    isPublished: dependencies.isPublished,
    publish: dependencies.publish,
    log: (message) => {
      output.stdout(infoLine(message));
    },
  };
  publishPackage(args.values["dry-run"] === true, publishDependencies);
  return 0;
};
