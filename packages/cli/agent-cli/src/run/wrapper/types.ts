/**
 * Terminal wrapper configuration
 */
export interface TerminalConfig {
  /** Terminal wrapper type */
  type: string;

  /** Session name (auto-generated if not provided) */
  sessionName?: string;

  /** Window name (defaults to directory basename) */
  windowName?: string;

  /** Detach after starting (run in background) */
  detach: boolean;

  /** Attach to existing session if it exists */
  attachExisting: boolean;
}

/**
 * Terminal wrapper executor interface
 */
export interface TerminalExecutor {
  /** Check if the terminal wrapper is available */
  isAvailable(): boolean;

  /** Check if already inside this terminal wrapper */
  isInside(): boolean;

  /** Execute a command inside the terminal wrapper */
  execute(
    config: TerminalConfig,
    command: string[],
    env: Record<string, string | undefined>,
    workDir: string,
    verbose: boolean,
  ): Promise<number>;
}
