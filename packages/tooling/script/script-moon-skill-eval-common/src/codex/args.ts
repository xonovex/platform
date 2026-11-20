interface IsolatedCodexOptions {
  readonly model: string;
}

// buildIsolatedCodexArgs is the Codex counterpart of buildIsolatedClaudeArgs:
// an ephemeral read-only exec that reads neither user config nor rules.
export const buildIsolatedCodexArgs = (
  options: IsolatedCodexOptions,
): readonly string[] => {
  const args = [
    "exec",
    "--json",
    "--ephemeral",
    "--sandbox",
    "read-only",
    "--ignore-user-config",
    "--ignore-rules",
    "--skip-git-repo-check",
  ];
  if (options.model.length > 0) args.push("--model", options.model);
  return args;
};
