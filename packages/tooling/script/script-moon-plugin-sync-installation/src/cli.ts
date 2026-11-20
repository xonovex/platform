import {parseCliArgs} from "@xonovex/script-moon-common/cli-args";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {
  errorLine,
  infoLine,
  successLine,
} from "@xonovex/script-moon-common/logging";
import type {Output} from "@xonovex/script-moon-common/output";
import {checkInstallations, syncInstallations} from "./readme-sync.js";

const syncTask =
  "npx moon run script-moon-plugin-sync-installation:plugin-installation-sync";

const cliSpec = {
  name: "moon-plugin-sync-installation",
  description:
    "Write the installation section of each plugin README, or report drift with --check",
  options: {
    check: {
      type: "boolean",
      description: "Report stale READMEs and exit 1 without writing",
    },
  },
} as const;

export interface SyncInstallationDependencies {
  readonly fs: FileSystem;
  readonly output: Output;
}

/** Writes or checks the installation section of each plugin README and returns the exit code. */
export const main = (
  argv: readonly string[],
  root: string,
  dependencies: SyncInstallationDependencies,
): number => {
  const {fs, output} = dependencies;
  const args = parseCliArgs(cliSpec, argv);
  if (args.kind === "help") {
    output.stdout(args.text);
    return 0;
  }
  if (args.values.check === true) {
    const drifted = checkInstallations(root, fs);
    if (drifted.length === 0) {
      output.stdout(successLine("Plugin installation sections are in sync."));
      return 0;
    }
    for (const line of [
      `Plugin installation sections have drifted in ${String(drifted.length)} README(s):`,
      ...drifted.map((path) => `  - ${path}`),
      `Run ${syncTask} to update them.`,
    ]) {
      output.stderr(errorLine(line));
    }
    return 1;
  }
  const updated = syncInstallations(root, fs);
  for (const path of updated) output.stdout(infoLine(`Updated ${path}`));
  output.stdout(
    successLine(
      `Updated ${String(updated.length)} plugin installation section(s).`,
    ),
  );
  return 0;
};
