// Claude Code loads CLAUDE.md, every other harness loads AGENTS.md. The repository
// keeps AGENTS.md as the single source and pairs it with a CLAUDE.md that points at
// the sibling, so guidance written once reaches both.
export const claudePointer = "See @AGENTS.md for complete documentation.";

export interface InstructionDirectory {
  // path is the repository-relative directory holding the instruction files.
  readonly path: string;
  readonly hasAgentsFile: boolean;
  // claudeText is the CLAUDE.md contents, or undefined when the file is absent.
  readonly claudeText: string | undefined;
}

// instructionDocFailures reports a directory whose instruction files break the
// pairing: a package group missing its AGENTS.md, an AGENTS.md with no CLAUDE.md
// beside it, or a CLAUDE.md that does not reference the sibling AGENTS.md. An
// unpaired AGENTS.md is invisible to Claude Code, which reads only CLAUDE.md.
export const instructionDocFailures = (
  directories: readonly InstructionDirectory[],
  requiredDirectories: readonly string[],
): readonly string[] => {
  const failures: string[] = [];
  const byPath = new Map(
    directories.map((directory) => [directory.path, directory]),
  );

  for (const required of [...requiredDirectories].toSorted()) {
    if (byPath.get(required)?.hasAgentsFile !== true) {
      failures.push(`${required} must document the group in AGENTS.md`);
    }
  }

  for (const directory of [...directories].toSorted((left, right) =>
    left.path.localeCompare(right.path),
  )) {
    if (!directory.hasAgentsFile) continue;
    if (directory.claudeText === undefined) {
      failures.push(
        `${directory.path}/AGENTS.md has no CLAUDE.md beside it: Claude Code reads only CLAUDE.md, so the guidance never loads`,
      );
      continue;
    }
    if (!directory.claudeText.includes("@AGENTS.md")) {
      failures.push(
        `${directory.path}/CLAUDE.md must point at its sibling with "${claudePointer}"`,
      );
    }
  }

  return failures;
};
