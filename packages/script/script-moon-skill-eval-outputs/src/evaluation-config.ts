import {existsSync, readdirSync, readFileSync} from "node:fs";
import {dirname, join, resolve} from "node:path";
import {parseCliArgs} from "@xonovex/script-moon-common";
import {boundedBatches} from "@xonovex/script-moon-common/batches";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {
  isDirectory,
  isFile,
  resolveClaudePluginDirectories,
  resolveGuideDirectory,
} from "@xonovex/script-moon-common/fs";
import {skillEvalModelDefaults} from "@xonovex/script-moon-common/skill-eval-models";
import {
  buildCodexArgs,
  buildGenerationClaudeArgs,
  evalEntries,
  MAX_OUTPUT_MODEL_CALLS,
  normalizeEval,
  outputModelCallCount,
  parseOutputOptions,
  validateUniqueEvaluationIds,
  type NormalizedEval,
  type OutputTier,
} from "./validation.js";

type EvaluationConfigResult =
  | {
      readonly success: true;
      readonly data: EvaluationConfig;
      readonly warnings: readonly string[];
    }
  | {
      readonly success: false;
      readonly error: string;
      readonly warnings: readonly string[];
    };

type EvaluationLoadResult =
  | {
      readonly success: true;
      readonly data: {
        readonly evaluations: readonly NormalizedEval[];
        readonly skillName: string;
        readonly tier: OutputTier;
      };
      readonly warnings: readonly string[];
    }
  | {
      readonly success: false;
      readonly error: string;
      readonly warnings: readonly string[];
    };

interface EvaluationConfigOptions {
  readonly environment?: Readonly<Record<string, string | undefined>>;
  readonly executableAvailable?: (command: string) => boolean;
  readonly workingDirectory?: string;
}

export interface EvaluationConfig {
  readonly benchmarkPath: string;
  readonly budget: number;
  readonly guideDirectory: string;
  readonly harness: "claude" | "codex";
  readonly concurrency: number;
  readonly cwd: string | undefined;
  readonly evaluations: readonly NormalizedEval[];
  readonly evaluationBatches: readonly (readonly NormalizedEval[])[];
  readonly evaluationsDirectory: string;
  readonly invalidRunPath: string;
  readonly iteration: string;
  readonly iterationDirectory: string;
  readonly judgeBudget: number;
  readonly judgeModel: string;
  readonly maxBatchModelCalls: number;
  readonly model: string;
  readonly runs: number;
  readonly shortName: string;
  readonly skillName: string;
  readonly tier: OutputTier;
  readonly timeout: number;
  readonly withArgs: readonly string[];
  readonly withoutArgs: readonly string[];
}

const cliDefinition = {
  name: "moon-skill-eval-outputs",
  description:
    "Run output-quality evals against a skill: with-skill vs without-skill.",
  options: {
    runs: {
      type: "string",
      description: "runs per arm per eval (env RUNS, default 1, maximum 3)",
    },
    concurrency: {
      type: "string",
      description:
        "parallel model invocations (env CONCURRENCY, default/maximum 2)",
    },
    "batch-size": {
      type: "string",
      description: "evals per sequential batch",
    },
    model: {
      type: "string",
      description:
        "generation model (CLAUDE_MODEL or CODEX_MODEL; pinned harness default when empty)",
    },
    harness: {
      type: "string",
      description:
        "model harness: claude or codex (env SKILL_EVAL_HARNESS, default claude)",
    },
    "judge-model": {
      type: "string",
      description:
        "grading model (JUDGE_MODEL or CODEX_JUDGE_MODEL; pinned harness default when empty)",
    },
    "disallowed-tools": {
      type: "string",
      description:
        "tools blocked in both arms (env DISALLOWED_TOOLS); without-skill also blocks Skill",
    },
    "gen-timeout": {
      type: "string",
      description:
        "per-generation timeout in seconds (env GEN_TIMEOUT, default 600)",
    },
    workspace: {
      type: "string",
      description:
        "workspace base dir (env WORKSPACE, default '<skill>-workspace')",
    },
    "eval-cwd": {
      type: "string",
      description:
        "working dir for generation runs (env EVAL_CWD, default current dir)",
    },
    "plugin-dir": {
      type: "string",
      description:
        "load the with-skill arm from a local plugin directory (env PLUGIN_DIR)",
    },
    "max-budget-usd": {
      type: "string",
      description:
        "hard per-generation spend cap up to 0.10 (env MAX_BUDGET_USD, default 0.10)",
    },
    "judge-max-budget-usd": {
      type: "string",
      description:
        "hard per-judge spend cap up to 0.10 (env JUDGE_MAX_BUDGET_USD, default 0.10)",
    },
  },
} as const;

const defaultExecutableAvailable = (command: string): boolean => {
  try {
    resolveExecutable(command);
    return true;
  } catch {
    return false;
  }
};

const skillNameFromSkillMd = (path: string): string | undefined => {
  if (!isFile(path)) return undefined;
  const content = readFileSync(path, "utf8");
  const frontmatterMatch = /^---\s*\n([\s\S]*?)\n---/.exec(content);
  const block = frontmatterMatch?.[1] ?? content;
  const nameMatch = /^name:\s*(.+?)\s*$/m.exec(block);
  const name = nameMatch?.[1];
  if (!name) return undefined;
  return name.replaceAll(/^["']|["']$/g, "").trim() || undefined;
};

const loadEvaluations = (
  evaluationsFile: string,
  evaluationsArgument: string,
): EvaluationLoadResult => {
  let data: unknown;
  try {
    data = JSON.parse(readFileSync(evaluationsFile, "utf8"));
  } catch (error: unknown) {
    const message = error instanceof Error ? error.message : String(error);
    return {
      success: false,
      error: `invalid JSON in ${evaluationsArgument}: ${message}`,
      warnings: [],
    };
  }

  const entriesResult = evalEntries(data);
  if (!entriesResult.success) {
    return {
      success: false,
      error: `invalid eval structure in ${evaluationsArgument}: ${entriesResult.error}`,
      warnings: [],
    };
  }
  if (entriesResult.data.evals.length === 0) {
    return {
      success: false,
      error: `${evaluationsArgument} has no evals`,
      warnings: [],
    };
  }

  const warnings: string[] = [];
  const evaluations: NormalizedEval[] = [];
  for (const [index, raw] of entriesResult.data.evals.entries()) {
    const result = normalizeEval(raw, index + 1);
    if (result.success) {
      evaluations.push(result.data);
    } else {
      warnings.push(`Skipping eval #${String(index + 1)}: ${result.error}`);
    }
  }
  if (evaluations.length === 0) {
    return {success: false, error: "no gradable evals", warnings};
  }
  const uniqueResult = validateUniqueEvaluationIds(evaluations);
  return uniqueResult.success
    ? {
        success: true,
        data: {
          evaluations: uniqueResult.data,
          skillName: entriesResult.data.skillName,
          tier: entriesResult.data.tier,
        },
        warnings,
      }
    : {success: false, error: uniqueResult.error, warnings};
};

const nextIteration = (base: string): string => {
  if (!existsSync(base)) return "iteration-1";
  const existing = readdirSync(base).flatMap((entry) => {
    const match = /^iteration-(\d+)$/.exec(entry);
    return match !== null && isDirectory(join(base, entry))
      ? [Number(match[1])]
      : [];
  });
  return `iteration-${String(Math.max(0, ...existing) + 1)}`;
};

const failure = (
  error: string,
  warnings: readonly string[] = [],
): EvaluationConfigResult => ({success: false, error, warnings});

export const buildEvaluationPrompt = (
  evaluationsDirectory: string,
  evaluation: NormalizedEval,
): string => {
  if (evaluation.files.length === 0) return evaluation.prompt;
  const paths = evaluation.files.map((file) =>
    resolve(evaluationsDirectory, file),
  );
  return (
    `${evaluation.prompt}\n\nRelevant input files (read them as needed):\n` +
    paths.map((path) => `- ${path}`).join("\n")
  );
};

export const resolveEvaluationConfig = (
  argv: readonly string[],
  options: EvaluationConfigOptions = {},
): EvaluationConfigResult => {
  const environment = options.environment ?? process.env;
  const executableAvailable =
    options.executableAvailable ?? defaultExecutableAvailable;
  const workingDirectory = options.workingDirectory ?? resolve(".");
  const {values, positionals} = parseCliArgs(cliDefinition, argv);
  const guideDirectory = resolveGuideDirectory(workingDirectory);
  const evaluationsArgument =
    positionals[0] ?? join(guideDirectory, "evals.json");
  const evaluationsFile = resolve(workingDirectory, evaluationsArgument);

  if (!isFile(evaluationsFile)) {
    return failure(`evals file not found: ${evaluationsArgument}`);
  }
  const harnessInput =
    (values.harness as string | undefined) ??
    environment.SKILL_EVAL_HARNESS ??
    "claude";
  if (harnessInput !== "claude" && harnessInput !== "codex") {
    return failure(
      `invalid harness '${harnessInput}'; expected claude or codex`,
    );
  }
  const harness = harnessInput;
  if (!executableAvailable(harness)) {
    return failure(`'${harness}' CLI not found in PATH`);
  }

  const skillArgument = positionals[1];
  const skillMd = join(guideDirectory, "SKILL.md");
  const skillName =
    skillArgument && skillArgument.length > 0
      ? skillArgument
      : skillNameFromSkillMd(skillMd);
  if (!skillName) {
    return failure(
      `no skill_name given and no 'name' frontmatter in ${skillMd}`,
    );
  }

  const loaded = loadEvaluations(evaluationsFile, evaluationsArgument);
  if (!loaded.success) return loaded;
  const requestedShortName = skillName.split(":").pop() ?? skillName;
  if (requestedShortName !== loaded.data.skillName) {
    return failure(
      `skill name mismatch: requested '${skillName}' but evals declare '${loaded.data.skillName}'`,
      loaded.warnings,
    );
  }

  const pluginDirectoryArgument =
    (values["plugin-dir"] as string | undefined) ??
    environment.PLUGIN_DIR ??
    "";
  const pluginDirectory = pluginDirectoryArgument
    ? resolve(workingDirectory, pluginDirectoryArgument)
    : undefined;
  if (pluginDirectory !== undefined && !isDirectory(pluginDirectory)) {
    return failure(
      `local plugin directory not found: ${pluginDirectoryArgument}`,
      loaded.warnings,
    );
  }

  const batchSizeRaw = values["batch-size"] as string | undefined;
  const iterationArgument = positionals[2] ?? "";
  const optionsResult = parseOutputOptions({
    runs: (values.runs as string | undefined) ?? environment.RUNS ?? "1",
    concurrency:
      (values.concurrency as string | undefined) ??
      environment.CONCURRENCY ??
      "2",
    timeout:
      (values["gen-timeout"] as string | undefined) ??
      environment.GEN_TIMEOUT ??
      "600",
    budget:
      (values["max-budget-usd"] as string | undefined) ??
      environment.MAX_BUDGET_USD ??
      "0.10",
    judgeBudget:
      (values["judge-max-budget-usd"] as string | undefined) ??
      environment.JUDGE_MAX_BUDGET_USD ??
      "0.10",
    ...(batchSizeRaw === undefined ? {} : {batchSize: batchSizeRaw}),
    ...(iterationArgument ? {iteration: iterationArgument} : {}),
  });
  if (!optionsResult.success) {
    return failure(
      `invalid evaluator options: ${optionsResult.error}`,
      loaded.warnings,
    );
  }

  const {runs, concurrency, timeout, batchSize} = optionsResult.data;
  const evaluationBatches = boundedBatches(
    loaded.data.evaluations,
    batchSize ?? Math.max(loaded.data.evaluations.length, 1),
  );
  const maxBatchModelCalls = Math.max(
    0,
    ...evaluationBatches.map((batch) =>
      outputModelCallCount(batch.length, runs),
    ),
  );
  if (maxBatchModelCalls > MAX_OUTPUT_MODEL_CALLS) {
    return failure(
      `output eval batch would launch ${String(maxBatchModelCalls)} model calls; ` +
        `maximum is ${String(MAX_OUTPUT_MODEL_CALLS)}. Split the eval set into bounded batches.`,
      loaded.warnings,
    );
  }

  const shortName = skillName.split(":").pop() ?? skillName;
  const workspaceArgument =
    (values.workspace as string | undefined) ?? environment.WORKSPACE ?? "";
  const workspace = workspaceArgument
    ? resolve(workingDirectory, workspaceArgument)
    : resolve(workingDirectory, `${shortName}-workspace`);
  const iteration = optionsResult.data.iteration ?? nextIteration(workspace);
  const iterationDirectory = join(workspace, iteration);
  const defaultModels = skillEvalModelDefaults(harness);
  const environmentModel =
    harness === "claude" ? environment.CLAUDE_MODEL : environment.CODEX_MODEL;
  const modelInput =
    (values.model as string | undefined) ??
    environmentModel ??
    defaultModels.generation;
  const model =
    modelInput.trim().length > 0 ? modelInput : defaultModels.generation;
  const judgeModelInput =
    (values["judge-model"] as string | undefined) ??
    (harness === "codex"
      ? environment.CODEX_JUDGE_MODEL
      : environment.JUDGE_MODEL) ??
    defaultModels.judge;
  const judgeModel =
    judgeModelInput.trim().length > 0 ? judgeModelInput : defaultModels.judge;
  const disallowedTools =
    (values["disallowed-tools"] as string | undefined) ??
    environment.DISALLOWED_TOOLS ??
    "Bash,Edit,Write,NotebookEdit,WebFetch";
  const budget = optionsResult.data.budget;
  const evaluationWorkingDirectory =
    (values["eval-cwd"] as string | undefined) ?? environment.EVAL_CWD ?? "";

  return {
    success: true,
    warnings: loaded.warnings,
    data: {
      benchmarkPath: join(iterationDirectory, "benchmark.json"),
      budget,
      guideDirectory,
      harness,
      concurrency,
      cwd: evaluationWorkingDirectory || undefined,
      evaluations: loaded.data.evaluations,
      evaluationBatches,
      evaluationsDirectory: dirname(evaluationsFile),
      invalidRunPath: join(iterationDirectory, "invalid-run.json"),
      iteration,
      iterationDirectory,
      judgeBudget: optionsResult.data.judgeBudget,
      judgeModel,
      maxBatchModelCalls,
      model,
      runs,
      shortName,
      skillName,
      tier: loaded.data.tier,
      timeout,
      withArgs:
        harness === "claude"
          ? buildGenerationClaudeArgs({
              arm: "with_skill",
              model,
              budget,
              disallowedTools,
              ...(pluginDirectory
                ? {
                    pluginDirectories:
                      resolveClaudePluginDirectories(pluginDirectory),
                  }
                : {}),
            })
          : buildCodexArgs({model}),
      withoutArgs:
        harness === "claude"
          ? buildGenerationClaudeArgs({
              arm: "without_skill",
              model,
              budget,
              disallowedTools,
            })
          : buildCodexArgs({model}),
    },
  };
};
