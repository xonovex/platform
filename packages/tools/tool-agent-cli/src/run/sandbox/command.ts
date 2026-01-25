/**
 * Shell-quote a string for safe use in a shell command
 */
function shellQuote(s: string): string {
  // If the string contains no special characters, return as-is
  if (/^[\w./:=-]+$/.test(s)) {
    return s;
  }
  // Otherwise, wrap in single quotes and escape any single quotes
  return "'" + s.replaceAll("'", "'\"'\"'") + "'";
}

/**
 * Build a shell command string from an array of arguments
 */
function buildShellCommand(args: string[]): string {
  return args.map(shellQuote).join(" ");
}

/**
 * Wrap a command with init commands that run before it
 *
 * If no init commands are provided, returns the original command unchanged.
 * Otherwise, returns a shell command that runs all init commands in sequence,
 * stopping on first failure, then runs the main command.
 *
 * @param command - The main command to run (e.g., ["claude", "--verbose"])
 * @param initCommands - Commands to run before the main command
 * @returns The wrapped command array
 */
export function wrapWithInitCommands(
  command: string[],
  initCommands?: string[],
): string[] {
  if (!initCommands || initCommands.length === 0) {
    return command;
  }

  // Build a shell command that chains init commands with &&, then runs the agent
  const initChain = initCommands.join(" && ");
  const mainCommand = buildShellCommand(command);
  const fullCommand = `${initChain} && exec ${mainCommand}`;

  return ["sh", "-c", fullCommand];
}
