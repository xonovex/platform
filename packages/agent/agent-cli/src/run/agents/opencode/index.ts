import type {AgentConfig} from "../types.js";

/**
 * OpenCode agent configuration
 */
export const opencodeAgent: AgentConfig = {
  type: "opencode",
  displayName: "OpenCode",
  binary: "opencode",
  nixPackage: "opencode",

  buildArgs(baseArgs, options) {
    // OpenCode uses CLI args for model selection
    return [...options.providerCliArgs, ...baseArgs];
  },

  buildEnv(_providerEnv) {
    // OpenCode doesn't use provider environment variables
    return {};
  },
};
