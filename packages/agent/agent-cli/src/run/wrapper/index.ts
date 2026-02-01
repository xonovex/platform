import {tmuxExecutor} from "./tmux/executor.js";
import type {TerminalConfig, TerminalExecutor} from "./types.js";

export type {TerminalConfig, TerminalExecutor} from "./types.js";

const executors: Record<string, TerminalExecutor> = {
  tmux: tmuxExecutor,
};

/**
 * Get a terminal executor by type
 */
export function getTerminalExecutor(
  type: string,
): TerminalExecutor | undefined {
  return executors[type];
}

/**
 * Execute a command in a terminal wrapper if configured
 * Returns undefined if no wrapper is configured or available
 */
export async function executeInTerminal(
  config: TerminalConfig | undefined,
  command: string[],
  env: Record<string, string | undefined>,
  workDir: string,
  verbose: boolean,
): Promise<number | undefined> {
  if (!config) {
    return undefined;
  }

  const executor = getTerminalExecutor(config.type);
  if (!executor) {
    return undefined;
  }

  // When inside terminal, only proceed if explicitly creating a detached window
  // (detach=true or windowName specified implies intent to create a new window)
  if (executor.isInside() && !config.detach && !config.windowName) {
    return undefined;
  }

  if (!executor.isAvailable()) {
    return undefined;
  }

  return executor.execute(config, command, env, workDir, verbose);
}
