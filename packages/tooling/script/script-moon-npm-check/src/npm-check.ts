import {join} from "node:path";
import {parseCliArgs} from "@xonovex/script-moon-common/cli-args";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {
  errorLine,
  infoLine,
  successLine,
} from "@xonovex/script-moon-common/logging";
import type {Output} from "@xonovex/script-moon-common/output";
import {readPkg} from "@xonovex/script-moon-common/package-json";
import {validatePackedPackage, type PackLister} from "./packed-package.js";
import {validateDeclaredFiles, validatePackage} from "./validate.js";

export interface NpmCheckDependencies {
  readonly fs: FileSystem;
  readonly listPackedFiles: PackLister;
  readonly output: Output;
}

type Attempt<Value> = {readonly value: Value} | {readonly error: unknown};

const attempt = <Value>(run: () => Value): Attempt<Value> => {
  try {
    return {value: run()};
  } catch (error: unknown) {
    return {error};
  }
};

const cliSpec = {
  name: "moon-npm-check",
  description: "Validate a package is ready for npm publishing",
} as const;

/** Checks that the package at packageRoot is ready to publish. */
export const main = (
  argv: readonly string[],
  packageRoot: string,
  dependencies: NpmCheckDependencies,
): number => {
  const args = parseCliArgs(cliSpec, argv);
  if (args.kind === "help") {
    dependencies.output.stdout(args.text);
    return 0;
  }
  const {fs, listPackedFiles, output} = dependencies;
  const packageJsonPath = join(packageRoot, "package.json");

  const read = attempt(() => readPkg(packageJsonPath, fs));
  if ("error" in read) {
    output.stderr(
      errorLine(`Failed to parse package.json: ${String(read.error)}`),
    );
    return 1;
  }
  const pkg = read.value;
  if (pkg.private === true) {
    output.stdout(
      infoLine(`Skipping private package ${pkg.name ?? "(unnamed)"}`),
    );
    return 0;
  }

  const packed = attempt(() => listPackedFiles(packageRoot));
  if ("error" in packed) {
    output.stderr(
      errorLine(`Failed to inspect packed package: ${String(packed.error)}`),
    );
    return 1;
  }

  const errors = [
    ...validatePackage(pkg),
    ...validateDeclaredFiles(pkg, (file) => {
      const path = join(packageRoot, file);
      return fs.isFile(path) || fs.isDirectory(path);
    }),
    ...validatePackedPackage(pkg, packed.value, (file) =>
      fs.readText(join(packageRoot, file)),
    ),
  ];
  if (errors.length > 0) {
    output.stderr(
      errorLine(
        `\n${pkg.name ?? packageJsonPath} is not ready for publishing:\n`,
      ),
    );
    for (const error of errors) output.stderr(errorLine(`  - ${error}`));
    return 1;
  }

  output.stdout(
    successLine(
      `${pkg.name ?? "unknown"}@${pkg.version ?? "unknown"} is ready for publishing`,
    ),
  );
  return 0;
};
