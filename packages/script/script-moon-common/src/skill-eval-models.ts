export type SkillEvalHarness = "claude" | "codex";

export interface SkillEvalModelDefaults {
  readonly generation: string;
  readonly judge: string;
}

const MODEL_DEFAULTS = {
  claude: {
    generation: "claude-haiku-4-5-20251001",
    judge: "claude-sonnet-4-6",
  },
  codex: {
    generation: "gpt-5.3-codex",
    judge: "gpt-5.3-codex",
  },
} as const satisfies Readonly<Record<SkillEvalHarness, SkillEvalModelDefaults>>;

export const skillEvalModelDefaults = (
  harness: SkillEvalHarness,
): SkillEvalModelDefaults => MODEL_DEFAULTS[harness];
