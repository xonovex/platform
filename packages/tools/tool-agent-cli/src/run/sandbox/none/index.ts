import { executeAgent } from "../../index.js";
import type { SandboxConfig, SandboxExecutor } from "../types.js";

/**
 * No-sandbox executor - runs agent directly
 */
export const noneExecutor: SandboxExecutor = {
  isAvailable(): Promise<boolean> {
    return Promise.resolve(true);
  },

  async execute(config: SandboxConfig): Promise<number> {
    return executeAgent({
      agent: config.agent,
      agentId: config.agentId,
      provider: config.provider,
      args: config.agentArgs,
      cwd: config.workDir,
      verbose: config.verbose,
      terminal: config.terminal,
      sandbox: false,
      customEnv: config.customEnv,
    });
  },

  getCommand(config: SandboxConfig): string[] {
    const agent = config.agent;
    if (!agent) {
      return ["claude", ...config.agentArgs];
    }

    // Use agent's buildArgs for consistency
    const args = agent.buildArgs(config.agentArgs, {
      sandbox: false,
      providerCliArgs: [],
    });

    return [agent.binary, ...args];
  },
};
