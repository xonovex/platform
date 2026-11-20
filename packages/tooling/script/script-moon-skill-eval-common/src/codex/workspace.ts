import {join} from "node:path";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import type {Environment} from "../environment.js";

/** A skill copied into a Codex workspace, with text appended to its SKILL.md. */
export interface StagedSkill {
  readonly source: string;
  readonly name: string;
  readonly appendix: string | undefined;
}

/** How a Codex process starts in its staged workspace. */
export interface CodexWorkspace {
  readonly cwd: string;
  readonly env: Environment;
}

/**
 * Where Codex finds the staged skills: in the workspace as project skills, or in
 * the isolated home as user skills.
 */
export type CodexSkillScope = "workspace" | "home";

/**
 * Stages a Codex run in workspace: a throwaway HOME and CODEX_HOME, so the run
 * reads none of the developer's configuration, and a copy of each skill it may use.
 */
export const stageCodexWorkspace = (
  workspace: string,
  scope: CodexSkillScope,
  skills: readonly StagedSkill[],
  environment: Environment,
  fs: FileSystem,
): CodexWorkspace => {
  const home = join(workspace, "home");
  const codexHome = join(workspace, ".codex");
  fs.makeDirectory(home);
  fs.makeDirectory(codexHome);
  const skillsRoot = join(
    scope === "workspace" ? workspace : home,
    ".agents",
    "skills",
  );
  for (const skill of skills) {
    const target = join(skillsRoot, skill.name);
    fs.copyDirectory(skill.source, target);
    if (skill.appendix !== undefined) {
      const skillFile = join(target, "SKILL.md");
      fs.writeFile(skillFile, fs.readText(skillFile) + skill.appendix);
    }
  }
  return {
    cwd: workspace,
    env: {...environment, CODEX_HOME: codexHome, HOME: home},
  };
};
