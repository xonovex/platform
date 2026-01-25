/**
 * Supported agent types
 */
export type AgentType = "claude" | "opencode";

/**
 * Agent configuration
 */
export interface AgentConfig {
  /** Agent type identifier */
  type: AgentType;

  /** Display name */
  displayName: string;

  /** Binary name to execute */
  binary: string;

  /** Nix package name that provides this agent (for nix sandbox) */
  nixPackage?: string;

  /**
   * Build CLI arguments for this agent
   * @param baseArgs - User-provided arguments
   * @param options - Execution options
   */
  buildArgs(baseArgs: string[], options: AgentExecOptions): string[];

  /**
   * Build environment variables for this agent
   * @param providerEnv - Environment from provider (if any)
   */
  buildEnv(providerEnv: Record<string, string>): Record<string, string>;
}

/**
 * Options passed to agent for building args/env
 */
export interface AgentExecOptions {
  /** Whether running in sandbox mode */
  sandbox: boolean;

  /** Provider CLI args (e.g., --model for opencode) */
  providerCliArgs: string[];
}
