import {
  isIncludedType,
  parseConventionalCommit,
  REPO_URL,
  type Commit,
} from "./git-log.js";

interface DepUpdate {
  readonly name: string;
  readonly version: string;
}

type BumpLevel = "patch" | "minor" | "major";

const sectionTitle = (level: BumpLevel): string => {
  switch (level) {
    case "major": {
      return "Major Changes";
    }
    case "minor": {
      return "Minor Changes";
    }
    case "patch": {
      return "Patch Changes";
    }
  }
};

const determineBumpLevel = (
  oldVersion: string,
  newVersion: string,
): BumpLevel => {
  const oldParts = oldVersion.split(".").map(Number);
  const newParts = newVersion.split(".").map(Number);
  const oldMajor = oldParts[0] ?? 0;
  const oldMinor = oldParts[1] ?? 0;
  const newMajor = newParts[0] ?? 0;
  const newMinor = newParts[1] ?? 0;
  if (newMajor > oldMajor) return "major";
  if (newMinor > oldMinor) return "minor";
  return "patch";
};

const formatCommitEntry = (
  hash: string,
  author: string,
  description: string,
): string => {
  const shortHash = hash.slice(0, 7);
  return `- [\`${shortHash}\`](${REPO_URL}/commit/${hash}) [@${author}](https://github.com/${author})! - ${description}`;
};

const generateChangelogEntry = (
  version: string,
  commits: readonly Commit[],
  bumpLevel: BumpLevel,
  depUpdates?: readonly DepUpdate[],
  includedTypes?: ReadonlySet<string>,
): string => {
  const lines: string[] = [
    `## ${version}`,
    "",
    `### ${sectionTitle(bumpLevel)}`,
    "",
  ];

  let hasEntries = false;

  for (const commit of commits) {
    for (const msg of commit.messages) {
      const parsed = parseConventionalCommit(msg);
      if (!parsed || !isIncludedType(parsed.type, includedTypes)) continue;
      lines.push(
        formatCommitEntry(commit.hash, commit.author, parsed.description),
      );
      hasEntries = true;
    }
  }

  if (depUpdates && depUpdates.length > 0) {
    for (const dep of depUpdates) {
      lines.push(`- Updated dependency \`${dep.name}\` to \`${dep.version}\``);
    }
    hasEntries = true;
  }

  if (!hasEntries) {
    lines.push("- Version bump");
  }

  lines.push("");
  return lines.join("\n");
};

const renderUpdatedChangelog = (
  existing: string | undefined,
  packageName: string,
  newEntry: string,
): string => {
  const title = `# ${packageName}`;
  if (existing === undefined) return `${title}\n\n${newEntry}`;
  const titleIndex = existing.indexOf(title);

  if (titleIndex === -1) {
    return `${title}\n\n${newEntry}\n${existing}`;
  }

  const insertPos = titleIndex + title.length + 1;
  return (
    existing.slice(0, insertPos) + "\n" + newEntry + existing.slice(insertPos)
  );
};

export {
  generateChangelogEntry,
  renderUpdatedChangelog,
  determineBumpLevel,
  formatCommitEntry,
};
export type {DepUpdate};
