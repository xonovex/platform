import {spawnSync} from "node:child_process";
import {mkdirSync} from "node:fs";
import {homedir} from "node:os";
import {logDebug, logError, logInfo} from "@xonovex/core";
import {claudeAgent} from "../../agents/claude/index.js";
import {getAgentIdEnv} from "../../id.js";
import {buildProviderEnvironment} from "../../providers/types.js";
import {executeInTerminal} from "../../wrapper/index.js";
import {
  bindMountsToBwrapArgs,
  ensureSandboxMountPoint,
  getApplicationBindMounts,
  getBaseMounts,
  getDefaultSandboxHome,
} from "../bindings.js";
import {wrapWithInitCommands} from "../command.js";
import {
  envToBwrapArgs,
  getSandboxEnvironment,
  mergeCustomEnv,
} from "../environment.js";
import {buildAgentCommand, spawnSandboxProcess} from "../shared-utils.js";
import type {SandboxConfig, SandboxExecutor} from "../types.js";

/**
 * Check if bubblewrap is installed
 */
function checkBwrapInstalled(): boolean {
  // eslint-disable-next-line sonarjs/no-os-command-from-path
  const result = spawnSync("which", ["bwrap"], {stdio: "pipe"});
  return result.status === 0;
}

/**
 * Build the full bubblewrap command arguments
 */
function buildBwrapArgs(config: SandboxConfig): string[] {
  const sandboxHome = config.homeDir ?? getDefaultSandboxHome();
  const home = homedir();

  // Ensure sandbox home exists
  mkdirSync(sandboxHome, {recursive: true});

  // Ensure the mount point exists in sandboxHome for workDir
  // (bwrap evaluates mount points before applying namespace mounts, so we must
  // pre-create mount points for paths under home in sandboxHome)
  ensureSandboxMountPoint(sandboxHome, config.workDir, config.verbose);

  // Get base system mounts (/, /dev, /proc, /tmp, sandboxHome -> home)
  const baseMounts = getBaseMounts(sandboxHome);

  // Get application mounts (workDir, user config, custom paths)
  const appMounts = getApplicationBindMounts(
    config.workDir,
    home,
    config.bindPaths,
    config.roBindPaths,
  );

  // Get sandbox environment with custom overrides
  const baseEnv = getSandboxEnvironment(home);
  const sandboxEnv = mergeCustomEnv(baseEnv, config.customEnv);
  const envArgs = envToBwrapArgs(sandboxEnv);

  // Unshare namespaces individually instead of --unshare-all
  // to avoid user namespace issues that can cause segfaults in Bun
  const args: string[] = [
    "--unshare-uts",
    "--unshare-ipc",
    "--unshare-pid",
    "--unshare-cgroup",
  ];

  // Share network if enabled
  if (config.network) {
    args.push("--share-net");
  }

  // Build command for the agent, wrapped with init commands if present
  const agentCommand = buildAgentCommand(config);
  const fullCommand = wrapWithInitCommands(
    agentCommand,
    config.sandboxInitCommands,
  );

  // Add environment, base mounts, app mounts, working directory, and command
  // Mount points for paths under home are pre-created in sandboxHome
  args.push(
    ...envArgs,
    ...bindMountsToBwrapArgs(baseMounts),
    ...bindMountsToBwrapArgs(appMounts),
    "--chdir",
    config.workDir,
    "--die-with-parent",
    "--",
    ...fullCommand,
  );

  return args;
}

/**
 * Bubblewrap sandbox executor
 */
export const bwrapExecutor: SandboxExecutor = {
  isAvailable(): Promise<boolean> {
    return Promise.resolve(checkBwrapInstalled());
  },

  async execute(config: SandboxConfig): Promise<number> {
    if (!checkBwrapInstalled()) {
      logError("bubblewrap (bwrap) is not installed");
      return 1;
    }

    const agent = config.agent ?? claudeAgent;
    const bwrapArgs = buildBwrapArgs(config);

    if (config.debug) {
      logDebug("bwrap " + bwrapArgs.join(" "));
    }

    if (config.verbose) {
      logInfo("Starting sandboxed " + agent.displayName + " with bubblewrap");
    }

    // Build environment - let the agent decide what to include
    const env: Record<string, string | undefined> = {
      ...process.env,
      ...getAgentIdEnv(config.agentId),
    };

    if (config.provider) {
      try {
        const providerEnv = buildProviderEnvironment(config.provider);
        const agentEnv = agent.buildEnv(providerEnv);
        Object.assign(env, agentEnv);
      } catch (error) {
        logError(
          error instanceof Error
            ? error.message
            : "Failed to configure provider",
        );
        return 1;
      }
    }

    const command = ["bwrap", ...bwrapArgs];

    // Try to execute in terminal wrapper if configured
    const terminalResult = await executeInTerminal(
      config.terminal,
      command,
      env,
      config.workDir,
      config.verbose,
    );
    if (terminalResult !== undefined) {
      return terminalResult;
    }

    return spawnSandboxProcess(
      "bwrap",
      bwrapArgs,
      {env: env as NodeJS.ProcessEnv, stdio: "inherit"},
      "bwrap",
    );
  },

  getCommand(config: SandboxConfig): string[] {
    return ["bwrap", ...buildBwrapArgs(config)];
  },
};
