import {join} from "node:path";
import {findExecutable} from "@xonovex/script-moon-common/executable";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import {parseFrontmatterName} from "@xonovex/script-moon-common/frontmatter";
import {
  skillEvalModelDefaults,
  type SkillEvalHarness,
} from "@xonovex/script-moon-common/skill-eval-models";
import type {Environment} from "./environment.js";

/** The harness a flag or SKILL_EVAL_HARNESS names, before it is checked. */
export const harnessInput = (
  flag: string | undefined,
  environment: Environment,
): string => flag ?? environment.SKILL_EVAL_HARNESS ?? "claude";

export const parseHarness = (input: string): SkillEvalHarness | undefined =>
  input === "claude" || input === "codex" ? input : undefined;

/** The harness CLI on the PATH of the environment, or undefined when it is absent. */
export const locateHarness = (
  harness: SkillEvalHarness,
  environment: Environment,
  platform: NodeJS.Platform,
  fs: FileSystem,
): string | undefined => {
  try {
    return findExecutable(
      harness,
      {
        searchPath: environment.PATH,
        extensions: environment.PATHEXT,
        platform,
      },
      fs,
    );
  } catch {
    return undefined;
  }
};

const MODEL_VARIABLES = {
  claude: {generation: "CLAUDE_MODEL", judge: "JUDGE_MODEL"},
  codex: {generation: "CODEX_MODEL", judge: "CODEX_JUDGE_MODEL"},
} as const;

/**
 * The model of a harness role: the flag, then the harness variable, then the
 * pinned default. An empty value also selects the default.
 */
export const resolveHarnessModel = (
  harness: SkillEvalHarness,
  role: "generation" | "judge",
  flag: string | undefined,
  environment: Environment,
): string => {
  const fallback = skillEvalModelDefaults(harness)[role];
  const input = flag ?? environment[MODEL_VARIABLES[harness][role]] ?? fallback;
  return input.trim().length > 0 ? input : fallback;
};

/** The bare name of a skill whose name may carry a plugin prefix. */
export const shortSkillName = (skillName: string): string =>
  skillName.split(":").pop() ?? skillName;

/** The name the SKILL.md of a skill directory declares, if it has one. */
export const declaredSkillName = (
  skillDirectory: string,
  fs: FileSystem,
): string | undefined => {
  const skillFile = join(skillDirectory, "SKILL.md");
  return fs.isFile(skillFile) ? parseFrontmatterName(skillFile, fs) : undefined;
};
