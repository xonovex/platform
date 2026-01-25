import type {AgentConfig} from "../types.js";

/**
 * Claude Code agent configuration
 */
export const claudeAgent: AgentConfig = {
  type: "claude",
  displayName: "Claude Code",
  binary: "claude",
  nixPackage: "claude-code",

  buildArgs(baseArgs, options) {
    const args: string[] = [];

    // Add permission bypass for sandbox mode
    if (options.sandbox) {
      args.push("--permission-mode", "bypassPermissions");
    }

    // Claude uses environment variables, not CLI args for provider config
    // So we ignore providerCliArgs

    args.push(...baseArgs);
    return args;
  },

  buildEnv(providerEnv) {
    // Claude uses the provider environment directly
    return {...providerEnv};
  },
};
