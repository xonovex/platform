import {join, relative} from "node:path";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
import {issue, type ValidationIssue} from "./validation.js";

// The skill tables sit under the README's `## Skills` heading, so a table
// elsewhere in the file is not read as a skill listing.
const skillsSection = (source: string): string => {
  const lines = source.split("\n");
  const start = lines.findIndex((line) => /^##\s+Skills\s*$/u.test(line));
  if (start === -1) return "";
  const rest = lines.slice(start + 1);
  const end = rest.findIndex((line) => /^##\s+\S/u.test(line));
  return (end === -1 ? rest : rest.slice(0, end)).join("\n");
};

// A skill row starts with a link whose text is the skill name and whose target
// is that skill's SKILL.md, for example `[using-git](skills/using-git/SKILL.md)`.
// A README without the heading lists no skill.
export const parseReadmeSkills = (source: string): ReadonlySet<string> =>
  new Set(
    [
      ...skillsSection(source).matchAll(
        /^\|\s*\[([a-z0-9-]+)\]\(skills\/\1\/SKILL\.md\)\s*\|/gmu,
      ),
    ].flatMap((match) => (match[1] === undefined ? [] : [match[1]])),
  );

// checkReadmeCatalog keeps the README skill tables and the skills directory
// from drifting apart when a skill is added or removed.
export const checkReadmeCatalog = (
  packageDirectory: string,
  skillNames: readonly string[],
  repositoryRoot: string,
  fs: FileSystem = nodeFileSystem,
): readonly ValidationIssue[] => {
  const readmePath = join(packageDirectory, "README.md");
  if (!fs.isFile(readmePath)) return [];
  const displayPath = relative(repositoryRoot, readmePath);
  const listed = parseReadmeSkills(fs.readText(readmePath));
  const shipped = new Set(skillNames);
  return [
    ...[...listed.difference(shipped)].map((name) =>
      issue(
        "readme.skill-missing",
        displayPath,
        `README lists skill '${name}' but 'skills/${name}/SKILL.md' does not exist`,
      ),
    ),
    ...[...shipped.difference(listed)].map((name) =>
      issue(
        "readme.skill-unlisted",
        displayPath,
        `skill '${name}' is not listed in the README skill tables`,
      ),
    ),
  ];
};
