import { spawn, type SpawnOptions } from "node:child_process";
import { logError } from "@xonovex/tool-lib";
import { claudeAgent } from "../agents/claude/index.js";
import type { AgentConfig } from "../agents/types.js";
import {
  buildProviderEnvironment,
  getProviderCliArgs,
} from "../providers/types.js";
import type { SandboxConfig } from "./types.js";

/**
 * Build the agent command array from sandbox configuration.
 * Combines agent binary with arguments, including provider CLI args if configured.
 *
 * @param config - Sandbox configuration
 * @param binaryPrefix - Optional prefix for the binary path (e.g., "/env/bin" for nix sandbox)
 */
export function buildAgentCommand(
  config: SandboxConfig,
  binaryPrefix?: string,
): string[] {
  const agent = config.agent ?? claudeAgent;

  // Get provider CLI args if available
  const providerCliArgs = config.provider
    ? getProviderCliArgs(config.provider)
    : [];

  // Let the agent build its own args
  const args = agent.buildArgs(config.agentArgs, {
    sandbox: true,
    providerCliArgs,
  });

  const binary = binaryPrefix
    ? binaryPrefix + "/" + agent.binary
    : agent.binary;

  return [binary, ...args];
}

/**
 * Build provider environment safely, handling errors.
 * Returns the combined agent environment or null if provider config fails.
 *
 * @param config - Sandbox configuration
 * @param agent - Agent configuration
 * @returns Combined environment or null on error
 */
export function buildProviderEnvironmentSafe(
  config: SandboxConfig,
  agent: AgentConfig,
): Record<string, string | undefined> | null {
  if (!config.provider) {
    return agent.buildEnv({});
  }

  try {
    const providerEnv = buildProviderEnvironment(config.provider);
    return agent.buildEnv(providerEnv);
  } catch (error) {
    logError(
      error instanceof Error ? error.message : "Failed to configure provider",
    );
    return null;
  }
}

/**
 * Error handler for spawn errors
 */
export type SpawnErrorHandler = (error: Error) => void;

/**
 * Default error handler that logs the error message
 */
function defaultErrorHandler(errorPrefix: string): SpawnErrorHandler {
  return (error: Error) => {
    logError(`Failed to execute ${errorPrefix}: ${error.message}`);
  };
}

/**
 * Spawn a sandbox process with error handling.
 * Returns a promise that resolves to the exit code.
 *
 * @param command - Command to execute
 * @param args - Command arguments
 * @param options - Spawn options
 * @param errorPrefix - Prefix for error messages (e.g., "Docker sandbox")
 * @param onError - Optional custom error handler
 * @returns Promise resolving to exit code (0 for success)
 */
export function spawnSandboxProcess(
  command: string,
  args: string[],
  options: SpawnOptions,
  errorPrefix: string,
  onError?: SpawnErrorHandler,
): Promise<number> {
  return new Promise((resolve) => {
    const child = spawn(command, args, options);

    child.on("error", (error) => {
      const handler = onError ?? defaultErrorHandler(errorPrefix);
      handler(error);
      resolve(1);
    });

    child.on("close", (code) => {
      resolve(code ?? 0);
    });
  });
}
