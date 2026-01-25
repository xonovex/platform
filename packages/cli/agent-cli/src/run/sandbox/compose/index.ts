import {spawnSync} from "node:child_process";
import {existsSync} from "node:fs";
import {homedir} from "node:os";
import {resolve} from "node:path";
import {logDebug, logError, logInfo, logWarning} from "@xonovex/core";
import {
  BASH_RESERVED_ENV_VARS,
  DEFAULT_FALLBACK_GID,
  DEFAULT_FALLBACK_UID,
} from "../../../constants.js";
import {claudeAgent} from "../../agents/claude/index.js";
import {getAgentIdEnv} from "../../id.js";
import {buildProviderEnvironment} from "../../providers/types.js";
import {executeInTerminal} from "../../wrapper/index.js";
import {bindMountsToDockerArgs, getApplicationBindMounts} from "../bindings.js";
import {wrapWithInitCommands} from "../command.js";
import {
  envToDockerArgs,
  getSandboxEnvironment,
  mergeCustomEnv,
} from "../environment.js";
import {buildAgentCommand, spawnSandboxProcess} from "../shared-utils.js";
import type {SandboxConfig, SandboxExecutor} from "../types.js";

// Default compose file location
const DEFAULT_COMPOSE_FILE = "stacks/ai-agent.yaml";

// Default service name
const DEFAULT_SERVICE = "ai-agent";

/**
 * Check if docker compose is available
 */
function checkComposeAvailable(): boolean {
  // Try "docker compose" (v2)
  // eslint-disable-next-line sonarjs/no-os-command-from-path
  const result = spawnSync("docker", ["compose", "version"], {stdio: "pipe"});
  return result.status === 0;
}

/**
 * Find the compose file, searching from workDir up to git root
 */
function findComposeFile(
  workDir: string,
  configFile?: string,
): string | undefined {
  if (configFile) {
    const absolutePath = resolve(workDir, configFile);
    if (existsSync(absolutePath)) {
      return absolutePath;
    }
    return undefined;
  }

  // Search from workDir up to find the default compose file
  let currentDir = workDir;
  const root = "/";

  while (currentDir !== root) {
    const candidatePath = resolve(currentDir, DEFAULT_COMPOSE_FILE);
    if (existsSync(candidatePath)) {
      return candidatePath;
    }
    currentDir = resolve(currentDir, "..");
  }

  return undefined;
}

/**
 * Build docker compose run arguments
 */
function buildComposeArgs(
  config: SandboxConfig,
  composeFile: string,
): string[] {
  const service = config.service ?? DEFAULT_SERVICE;
  const agentCommand = buildAgentCommand(config);
  const fullCommand = wrapWithInitCommands(
    agentCommand,
    config.sandboxInitCommands,
  );
  const home = homedir();

  // Get application-level bind mounts (workdir + user config + custom paths)
  // These are added on top of compose file volumes
  const appMounts = getApplicationBindMounts(
    config.workDir,
    home,
    config.bindPaths,
    config.roBindPaths,
  );
  const appMountArgs = bindMountsToDockerArgs(appMounts);

  // Get sandbox environment with custom overrides
  // Don't inherit host PATH - container has its own filesystem
  const baseEnv = getSandboxEnvironment(home, [], false);
  const sandboxEnv = mergeCustomEnv(baseEnv, config.customEnv);
  const envArgs = envToDockerArgs(sandboxEnv);

  const args: string[] = [
    "compose",
    "-f",
    composeFile,
    "run",
    "--rm",
    // Environment variables
    ...envArgs,
    // Application-level volumes on top of compose file volumes
    ...appMountArgs,
    service,
    ...fullCommand,
  ];

  return args;
}

/**
 * Build environment variables for docker compose
 */
function buildComposeEnv(config: SandboxConfig): Record<string, string> {
  const env: Record<string, string> = {};

  // Copy process.env but skip read-only bash variables that cause issues with tmux/bash
  for (const [key, value] of Object.entries(process.env)) {
    if (value === undefined) continue;
    if (
      BASH_RESERVED_ENV_VARS.includes(
        key as (typeof BASH_RESERVED_ENV_VARS)[number],
      )
    )
      continue;
    env[key] = value;
  }

  return {
    ...env,
    // Set AGENT_WORK_DIR for docker-compose volume mounts (current working directory)
    AGENT_WORK_DIR: config.workDir,
    // Set AGENT_REPO_DIR for data volumes (original repo, defaults to workDir)
    AGENT_REPO_DIR: config.repoDir ?? config.workDir,
    // Set AGENT_UID/AGENT_GID for user mapping
    AGENT_UID: String(process.getuid?.() ?? DEFAULT_FALLBACK_UID),
    AGENT_GID: String(process.getgid?.() ?? DEFAULT_FALLBACK_GID),
    // Add agent ID environment variables
    ...getAgentIdEnv(config.agentId),
  };
}

/**
 * Docker Compose sandbox executor
 */
export const composeExecutor: SandboxExecutor = {
  isAvailable(): Promise<boolean> {
    return Promise.resolve(checkComposeAvailable());
  },

  async execute(config: SandboxConfig): Promise<number> {
    if (!checkComposeAvailable()) {
      logError("Docker Compose is not available (docker compose v2 required)");
      return 1;
    }

    // Look for compose file in repoDir (original repo) when using worktree
    const composeSearchDir = config.repoDir ?? config.workDir;
    const composeFile = findComposeFile(composeSearchDir, config.composeFile);
    if (!composeFile) {
      logError(
        "Docker Compose file not found: " +
          (config.composeFile ?? DEFAULT_COMPOSE_FILE),
      );
      return 1;
    }

    const agent = config.agent ?? claudeAgent;
    const dockerArgs = buildComposeArgs(config, composeFile);
    const env = buildComposeEnv(config);

    // If provider is specified, add environment variables
    if (config.provider) {
      try {
        const providerEnv = buildProviderEnvironment(config.provider);
        const agentEnv = agent.buildEnv(providerEnv);

        // Pass agent environment variables to docker compose via the environment
        // The compose file should use ${VAR:-} syntax to pass these through
        for (const [key, value] of Object.entries(agentEnv)) {
          env[key] = value;
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

    if (config.debug) {
      logDebug(
        "AGENT_WORK_DIR=" + config.workDir + " docker " + dockerArgs.join(" "),
      );
    }

    if (config.verbose) {
      logInfo(
        "Starting " +
          agent.displayName +
          " via Docker Compose service: " +
          (config.service ?? DEFAULT_SERVICE),
      );
      logInfo("Compose file: " + composeFile);
    }

    // Warn about provider override when using compose with pre-configured services
    if (config.provider && config.service?.includes("-glm")) {
      logWarning(
        "Provider override may conflict with pre-configured service environment",
      );
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
      {env, stdio: "inherit", cwd: config.workDir},
      "docker compose",
    );
  },

  getCommand(config: SandboxConfig): string[] {
    const composeSearchDir = config.repoDir ?? config.workDir;
    const composeFile = findComposeFile(composeSearchDir, config.composeFile);
    if (!composeFile) {
      return ["# Error: Compose file not found"];
    }
    return [
      "AGENT_WORK_DIR=" + config.workDir,
      "AGENT_REPO_DIR=" + (config.repoDir ?? config.workDir),
      "docker",
      ...buildComposeArgs(config, composeFile),
    ];
  },
};
