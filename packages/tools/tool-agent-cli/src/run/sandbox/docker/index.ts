import { spawnSync } from "node:child_process";
import { homedir } from "node:os";
import { logDebug, logError, logInfo } from "@xonovex/tool-lib";
import {
  DEFAULT_FALLBACK_GID,
  DEFAULT_FALLBACK_UID,
} from "../../../constants.js";
import { claudeAgent } from "../../agents/claude/index.js";
import { getAgentIdEnv } from "../../id.js";
import { buildProviderEnvironment } from "../../providers/types.js";
import { executeInTerminal } from "../../wrapper/index.js";
import {
  bindMountsToDockerArgs,
  getApplicationBindMounts,
  getDefaultSandboxHome,
} from "../bindings.js";
import { wrapWithInitCommands } from "../command.js";
import {
  envToDockerArgs,
  getSandboxEnvironment,
  mergeCustomEnv,
} from "../environment.js";
import { buildAgentCommand, spawnSandboxProcess } from "../shared-utils.js";
import type { SandboxConfig, SandboxExecutor } from "../types.js";

// Default Docker image (fallback, but users should provide their own with agent installed)
const DEFAULT_DOCKER_IMAGE = "node:trixie-slim";

/**
 * Check if Docker is installed and running
 */
function checkDockerAvailable(): boolean {
  // eslint-disable-next-line sonarjs/no-os-command-from-path
  const result = spawnSync("docker", ["info"], { stdio: "pipe" });
  return result.status === 0;
}

/**
 * Build Docker run arguments
 */
function buildDockerArgs(config: SandboxConfig): string[] {
  const sandboxHome = config.homeDir ?? getDefaultSandboxHome();
  const home = homedir();

  // Get application-level bind mounts (workdir + user config + custom paths)
  const appMounts = getApplicationBindMounts(
    config.workDir,
    home,
    config.bindPaths,
    config.roBindPaths,
  );

  // Add sandbox home → real home mapping
  const mounts = [
    { source: sandboxHome, target: home, readOnly: false },
    ...appMounts,
  ];

  // Get sandbox environment with custom overrides
  // Don't inherit host PATH - container has its own filesystem
  const baseEnv = getSandboxEnvironment(home, [], false);
  const sandboxEnv = {
    ...mergeCustomEnv(baseEnv, config.customEnv),
    ...getAgentIdEnv(config.agentId),
  };
  const envArgs = envToDockerArgs(sandboxEnv);

  // User mapping (run as current user)
  const uid = String(process.getuid?.() ?? DEFAULT_FALLBACK_UID);
  const gid = String(process.getgid?.() ?? DEFAULT_FALLBACK_GID);

  // Build agent command, wrapped with init commands if present
  const agentCommand = buildAgentCommand(config);
  const fullCommand = wrapWithInitCommands(
    agentCommand,
    config.sandboxInitCommands,
  );

  // Use provided image or default
  const image = config.image ?? DEFAULT_DOCKER_IMAGE;

  // Build args array
  const args: string[] = [
    "run",
    "--rm",
    "-it",
    // Network mode
    ...(config.network ? [] : ["--network", "none"]),
    // Working directory
    "-w",
    config.workDir,
    // User mapping
    "-u",
    uid + ":" + gid,
    // Environment variables
    ...envArgs,
    // Volume mounts
    ...bindMountsToDockerArgs(mounts),
    // Image and command
    image,
    ...fullCommand,
  ];

  return args;
}

/**
 * Docker sandbox executor
 */
export const dockerExecutor: SandboxExecutor = {
  isAvailable(): Promise<boolean> {
    return Promise.resolve(checkDockerAvailable());
  },

  async execute(config: SandboxConfig): Promise<number> {
    if (!checkDockerAvailable()) {
      logError("Docker is not available (not installed or daemon not running)");
      return 1;
    }

    const agent = config.agent ?? claudeAgent;
    const dockerArgs = buildDockerArgs(config);

    if (config.debug) {
      logDebug("docker " + dockerArgs.join(" "));
    }

    if (config.verbose) {
      logInfo(
        "Starting sandboxed " + agent.displayName + " in Docker container",
      );
    }

    // Build environment - let the agent decide what to include
    const env = { ...process.env };

    if (config.provider) {
      try {
        const providerEnv = buildProviderEnvironment(config.provider);
        const agentEnv = agent.buildEnv(providerEnv);

        // Pass agent environment variables to Docker via -e flags
        const wIndex = dockerArgs.indexOf("-w");
        for (const [key, value] of Object.entries(agentEnv)) {
          dockerArgs.splice(wIndex, 0, "-e", key + "=" + value);
        }
      } catch (error) {
        logError(
          error instanceof Error
            ? error.message
            : "Failed to configure provider",
        );
        return 1;
      }
    }

    const command = ["docker", ...dockerArgs];

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
      "docker",
      dockerArgs,
      { env, stdio: "inherit" },
      "docker",
    );
  },

  getCommand(config: SandboxConfig): string[] {
    return ["docker", ...buildDockerArgs(config)];
  },
};
