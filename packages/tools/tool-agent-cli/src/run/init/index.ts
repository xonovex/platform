import { execSync, type ExecSyncOptions } from "node:child_process";
import { logError, logInfo } from "@xonovex/tool-lib";

/**
 * Execute an init command in the specified directory
 *
 * @param command - The shell command to execute
 * @param cwd - Working directory for the command
 * @param verbose - Enable verbose output
 * @throws Error if the command fails
 */
export function executeInitCommand(
  command: string,
  cwd: string,
  verbose = false,
): void {
  if (verbose) {
    logInfo(`Running init command: ${command}`);
  }

  const options: ExecSyncOptions = {
    cwd,
    stdio: verbose ? "inherit" : "pipe",
    shell: "/bin/sh",
  };

  try {
    execSync(command, options);
  } catch (error) {
    const message = error instanceof Error ? error.message : String(error);
    logError(`Init command failed: ${message}`);
    throw error;
  }
}

/**
 * Execute multiple init commands in sequence
 *
 * @param commands - Array of shell commands to execute
 * @param cwd - Working directory for the commands
 * @param verbose - Enable verbose output
 * @throws Error if any command fails
 */
export function executeInitCommands(
  commands: string[],
  cwd: string,
  verbose = false,
): void {
  for (const command of commands) {
    executeInitCommand(command, cwd, verbose);
  }
}
