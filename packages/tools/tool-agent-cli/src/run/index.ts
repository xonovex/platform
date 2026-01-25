import { spawn } from "node:child_process";
import { logError, logInfo } from "@xonovex/tool-lib";
import { claudeAgent } from "./agents/claude/index.js";
import type { AgentConfig, AgentExecOptions } from "./agents/types.js";
import { getAgentIdEnv } from "./id.js";
import {
  buildProviderEnvironment,
  getProviderCliArgs,
  type ModelProvider,
} from "./providers/types.js";
import { parseCustomEnv } from "./sandbox/environment.js";
import { executeInTerminal, type TerminalConfig } from "./wrapper/index.js";

export interface ExecuteOptions {
  /** Agent to use (defaults to claude) */
  agent?: AgentConfig;

  /** Unique agent identifier */
  agentId?: string;

  /** Model provider to use */
  provider?: ModelProvider;

  /** Additional arguments to pass to the agent */
  args: string[];

  /** Working directory */
  cwd?: string;

  /** Enable verbose output */
  verbose?: boolean;

  /** Terminal wrapper configuration */
  terminal?: TerminalConfig;

  /** Whether running in sandbox mode */
  sandbox?: boolean;

  /** Custom environment variables (KEY=VALUE format) */
  customEnv?: string[];
}

/**
 * Execute an agent with the specified options
 */
export async function executeAgent(options: ExecuteOptions): Promise<number> {
  const {
    agent = claudeAgent,
    agentId,
    provider,
    args,
    cwd = process.cwd(),
    verbose = false,
    terminal,
    sandbox = false,
    customEnv = [],
  } = options;

  if (verbose) {
    logInfo("Using agent: " + agent.displayName);
    if (provider) {
      logInfo("Using provider: " + provider.displayName);
    }
  }

  // Get provider configuration
  let providerEnv: Record<string, string> = {};
  let providerCliArgs: string[] = [];

  if (provider) {
    try {
      providerEnv = buildProviderEnvironment(provider);
      providerCliArgs = getProviderCliArgs(provider);
    } catch (error) {
      logError(
        error instanceof Error ? error.message : "Failed to configure provider",
      );
      return 1;
    }
  }

  // Let the agent build its own args and env
  const execOptions: AgentExecOptions = {
    sandbox,
    providerCliArgs,
  };

  const agentArgs = agent.buildArgs(args, execOptions);
  const agentEnv = agent.buildEnv(providerEnv);

  // Parse custom environment variables
  const customEnvVars = parseCustomEnv(customEnv);

  // Merge with process env, agent env, custom env, and agent ID
  const env: Record<string, string | undefined> = {
    ...process.env,
    ...agentEnv,
    ...customEnvVars,
    ...(agentId ? getAgentIdEnv(agentId) : {}),
  };

  // Build command
  const command = [agent.binary, ...agentArgs];

  if (verbose) {
    logInfo("Executing: " + command.join(" "));
  }

  // Try to execute in terminal wrapper if configured
  const terminalResult = await executeInTerminal(
    terminal,
    command,
    env,
    cwd,
    verbose,
  );
  if (terminalResult !== undefined) {
    return terminalResult;
  }

  // Execute directly
  const binary = command[0];
  if (!binary) {
    logError("No command to execute");
    return 1;
  }

  return new Promise((resolve) => {
    const child = spawn(binary, command.slice(1), {
      cwd,
      env: env as NodeJS.ProcessEnv,
      stdio: "inherit",
    });

    child.on("error", (error: Error) => {
      logError("Failed to execute " + agent.displayName + ": " + error.message);
      resolve(1);
    });

    child.on("close", (code: number | null) => {
      resolve(code ?? 0);
    });
  });
}
