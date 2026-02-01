/**
 * Combine agent arguments from positional args and unknown options (after --)
 *
 * @param positionalArgs - Positional agent args from CLI
 * @param unknownArgs - Unknown options passed through from Commander (after --)
 * @returns Combined array of all agent arguments, deduplicated
 */
export function combineAgentArgs(
  positionalArgs: string[],
  unknownArgs: string[],
): string[] {
  return [
    ...positionalArgs,
    ...unknownArgs.filter((a) => !positionalArgs.includes(a)),
  ];
}
