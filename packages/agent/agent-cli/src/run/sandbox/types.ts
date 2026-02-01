import type {AgentConfig} from "../agents/types.js";
import type {ModelProvider} from "../providers/types.js";
import type {TerminalConfig} from "../wrapper/types.js";

/**
 * Sandbox execution method
 */
export type SandboxMethod = "none" | "bwrap" | "docker" | "compose" | "nix";

/**
 * Sandbox configuration options
 */
export interface SandboxConfig {
  /** Unique agent identifier */
  agentId: string;

  /** Sandbox method to use */
  method: SandboxMethod;

  /** Agent to use */
  agent?: AgentConfig;

  /** Custom sandbox home directory */
  homeDir?: string;

  /** Docker image to use */
  image?: string;

  /** Docker Compose file path */
  composeFile?: string;

  /** Docker Compose service name */
  service?: string;

  /** Working directory */
  workDir: string;

  /** Original repository directory (for worktrees, this is the source repo) */
  repoDir?: string;

  /** Enable network access */
  network: boolean;

  /** Additional read-write bind mounts */
  bindPaths: string[];

  /** Additional read-only bind mounts */
  roBindPaths: string[];

  /** Custom environment variables (KEY=VALUE) */
  customEnv: string[];

  /** Model provider to use */
  provider?: ModelProvider;

  /** Arguments to pass to the agent */
  agentArgs: string[];

  /** Commands to run inside the sandbox before the agent */
  sandboxInitCommands?: string[];

  /** Enable verbose output */
  verbose: boolean;

  /** Enable debug mode */
  debug: boolean;

  /** Dry run - show config without executing */
  dryRun: boolean;

  /** Terminal wrapper configuration */
  terminal?: TerminalConfig;
}

/**
 * Sandbox executor interface
 */
export interface SandboxExecutor {
  /** Check if this sandbox method is available */
  isAvailable(): Promise<boolean>;

  /** Execute agent in the sandbox */
  execute(config: SandboxConfig): Promise<number>;

  /** Get the command that would be executed (for dry-run) */
  getCommand(config: SandboxConfig): string[];
}
