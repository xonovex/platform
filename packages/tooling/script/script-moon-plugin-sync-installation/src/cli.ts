import {parseCliArgs} from "@xonovex/script-moon-common/cli-args";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {
  logError,
  logInfo,
  logSuccess,
} from "@xonovex/script-moon-common/logging";
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

export const runCli = (
  argv: readonly string[],
  root: string,
  fs: FileSystem,
): number => {
  const {values} = parseCliArgs(cliSpec, argv);
  if (values.check === true) {
    const drifted = checkInstallations(root, fs);
    if (drifted.length === 0) {
      logSuccess("Plugin installation sections are in sync.");
      return 0;
    }
    logError(
      `Plugin installation sections have drifted in ${String(drifted.length)} README(s):`,
    );
    for (const path of drifted) logError(`  - ${path}`);
    logError(`Run ${syncTask} to update them.`);
    return 1;
  }
  const updated = syncInstallations(root, fs);
  for (const path of updated) logInfo(`Updated ${path}`);
  logSuccess(
    `Updated ${String(updated.length)} plugin installation section(s).`,
  );
  return 0;
};
