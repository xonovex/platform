import {join, relative} from "node:path";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import {issue, type ValidationIssue} from "./validation.js";

// The command tables sit under the README's `## Commands` heading, so a table
// elsewhere in the file is not read as a command listing.
const commandsSection = (source: string): string | undefined => {
  const lines = source.split("\n");
  const start = lines.findIndex((line) => /^##\s+Commands\s*$/u.test(line));
  if (start === -1) return undefined;
  const rest = lines.slice(start + 1);
  const end = rest.findIndex((line) => /^##\s+\S/u.test(line));
  return (end === -1 ? rest : rest.slice(0, end)).join("\n");
};

// undefined means the README declares no command catalog, which is a different
// state from a catalog that lists nothing.
export const parseReadmeCommands = (
  source: string,
): ReadonlySet<string> | undefined => {
  const section = commandsSection(source);
  if (section === undefined) return undefined;
  return new Set(
    [...section.matchAll(/^\|\s*`([a-z0-9-]+)`\s*\|/gmu)].flatMap((match) =>
      match[1] === undefined ? [] : [match[1]],
    ),
  );
};

// checkReadmeCatalog keeps the README command tables and the commands directory
// from drifting apart when a command is added or removed.
export const checkReadmeCatalog = (
  packageDirectory: string,
  commandNames: readonly string[],
  repositoryRoot: string,
  fs: FileSystem = nodeFileSystem,
): readonly ValidationIssue[] => {
  const readmePath = join(packageDirectory, "README.md");
  if (!fs.isFile(readmePath)) return [];
  const displayPath = relative(repositoryRoot, readmePath);
  const listed = parseReadmeCommands(fs.readText(readmePath));
  if (listed === undefined) return [];
  const shipped = new Set(commandNames);
  return [
    ...[...listed.difference(shipped)].map((name) =>
      issue(
        "readme.command-missing",
        displayPath,
        `README lists command '${name}' but 'commands/${name}.md' does not exist`,
      ),
    ),
    ...[...shipped.difference(listed)].map((name) =>
      issue(
        "readme.command-unlisted",
        displayPath,
        `command '${name}' is not listed in the README command tables`,
      ),
    ),
  ];
};
