import {existsSync, readFileSync, writeFileSync} from "node:fs";
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
): string => {
  const lines: string[] = [
    `## ${version}`,
    "",
    `### ${sectionTitle(bumpLevel)}`,
    "",
  ];

  for (const commit of commits) {
    const parsed = parseConventionalCommit(commit.message);
    if (!parsed || !isIncludedType(parsed.type)) continue;
    lines.push(
      formatCommitEntry(commit.hash, commit.author, parsed.description),
    );
  }

  if (depUpdates && depUpdates.length > 0) {
    for (const dep of depUpdates) {
      lines.push(`- Updated dependency \`${dep.name}\` to \`${dep.version}\``);
    }
  }

  lines.push("");
  return lines.join("\n");
};

const updateChangelog = (
  changelogPath: string,
  packageName: string,
  newEntry: string,
): void => {
  const title = `# ${packageName}`;

  if (!existsSync(changelogPath)) {
    writeFileSync(changelogPath, `${title}\n\n${newEntry}`, "utf8");
    return;
  }

  const existing = readFileSync(changelogPath, "utf8");
  const titleIndex = existing.indexOf(title);

  if (titleIndex === -1) {
    writeFileSync(changelogPath, `${title}\n\n${newEntry}`, "utf8");
    return;
  }

  const insertPos = titleIndex + title.length + 1;
  const updated =
    existing.slice(0, insertPos) + "\n" + newEntry + existing.slice(insertPos);
  writeFileSync(changelogPath, updated, "utf8");
};

export {
  generateChangelogEntry,
  updateChangelog,
  determineBumpLevel,
  sectionTitle,
  formatCommitEntry,
};
export type {DepUpdate, BumpLevel};
