import {dirname, join} from "node:path";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {resolveSkillDirectory} from "@xonovex/script-moon-common/skill-directory";

export interface SkillLocation {
  readonly kind: "found";
  /** The SKILL.md to validate. */
  readonly skillPath: string;
  readonly skillDir: string;
}

export interface MissingSkill {
  readonly kind: "missing";
  readonly message: string;
}

// A directory names a skill or the package that holds one; a file names the
// SKILL.md itself.
export const locateSkill = (
  target: string,
  fs: FileSystem,
): SkillLocation | MissingSkill => {
  let skillPath: string;
  if (fs.isDirectory(target)) {
    skillPath = join(resolveSkillDirectory(target, fs), "SKILL.md");
  } else if (fs.isFile(target)) {
    skillPath = target;
  } else {
    return {kind: "missing", message: `target not found: ${target}`};
  }
  if (!fs.isFile(skillPath)) {
    return {kind: "missing", message: `SKILL.md not found at ${skillPath}`};
  }
  return {kind: "found", skillPath, skillDir: dirname(skillPath)};
};
