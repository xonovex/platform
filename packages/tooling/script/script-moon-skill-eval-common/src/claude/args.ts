interface TriggerClaudeOptions {
  readonly model: string;
  readonly budget: number;
  readonly pluginDirectories: readonly string[];
}

// One turn to choose a skill, and a second only for the run that needs it: a model
// that writes a skill name the harness cannot resolve launches nothing, and the turn
// it spent says nothing about routing. The second turn lets it answer the harness's
// error with a name that exists. A run that picks a skill on its first turn ends
// there, so the extra turn costs nothing on the runs that get it right.
export const TRIGGER_MAX_TURNS = 2;

// buildIsolatedClaudeArgs carries the flags that keep an eval run hermetic: no
// settings from the developer's machine, no MCP servers, no persisted session, no
// browser. A run that inherits any of them scores the local environment instead of
// the skill, so every Claude eval invocation starts from this set.
export const buildIsolatedClaudeArgs = (
  outputFormat: "json" | "stream-json",
): readonly string[] => [
  "-p",
  "--output-format",
  outputFormat,
  "--setting-sources",
  "",
  "--strict-mcp-config",
  "--mcp-config",
  '{"mcpServers":{}}',
  "--no-session-persistence",
  "--no-chrome",
];

export const buildTriggerClaudeArgs = (
  options: TriggerClaudeOptions,
): readonly string[] => {
  const args = [
    ...buildIsolatedClaudeArgs("stream-json"),
    "--verbose",
    "--include-partial-messages",
    "--model",
    options.model,
    "--max-budget-usd",
    String(options.budget),
    "--max-turns",
    String(TRIGGER_MAX_TURNS),
    "--system-prompt",
    "Decide which available skill best matches the user request. " +
      "If one applies, invoke only that Skill immediately. Otherwise reply with one short sentence. " +
      "Do not perform the requested task.",
    "--tools",
    "Skill",
  ];
  for (const pluginDirectory of options.pluginDirectories) {
    args.push("--plugin-dir", pluginDirectory);
  }
  return args;
};
