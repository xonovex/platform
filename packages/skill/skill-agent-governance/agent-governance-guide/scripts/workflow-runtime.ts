import {z} from "zod";

export const workflowInvocationApiVersion =
  "workflow.xonovex.com/v1alpha1" as const;

export const triggerSourceSchema = z.enum([
  "manual",
  "agent-harness-hook",
  "ci-cd-hook",
  "provider-webhook",
  "schedule",
  "sensor",
  "api",
  "agent",
]);

const nativeReferenceSchema = z.string().min(1);
const commandBudgetSchema = z
  .object({
    timeoutSeconds: z.number().int().positive().max(3_600),
  })
  .strict();
const adaptiveBudgetSchema = commandBudgetSchema
  .extend({
    tokenBudget: z.number().int().positive(),
    costBudget: z.number().nonnegative(),
    retryLimit: z.number().int().nonnegative().max(3),
  })
  .strict();

const a1OversightSchema = z
  .object({
    level: z.literal("A1"),
    independentCritiqueReference: nativeReferenceSchema,
  })
  .strict();
const a2OversightSchema = z
  .object({
    level: z.literal("A2"),
    independentCritiqueReference: nativeReferenceSchema,
    journalReference: nativeReferenceSchema,
    approvalReference: nativeReferenceSchema,
    cancellationReference: nativeReferenceSchema,
    killSwitchReference: nativeReferenceSchema,
  })
  .strict();
const a3OversightSchema = z
  .object({
    level: z.literal("A3"),
    independentCritiqueReference: nativeReferenceSchema,
    journalReference: nativeReferenceSchema,
    approvalReference: nativeReferenceSchema,
    cancellationReference: nativeReferenceSchema,
    killSwitchReference: nativeReferenceSchema,
    verdictReference: nativeReferenceSchema,
    protectedTargetReferences: z.array(nativeReferenceSchema).min(1),
    escalationReference: nativeReferenceSchema,
    provenanceReference: nativeReferenceSchema,
  })
  .strict();

export const autonomyOversightSchema = z.discriminatedUnion("level", [
  a1OversightSchema,
  a2OversightSchema,
  a3OversightSchema,
]);

const workflowScriptExecutionSchema = z
  .object({
    family: z.literal("workflow-script"),
    module: z.string().min(1),
    budget: commandBudgetSchema,
  })
  .strict();
const workflowScriptLlmExecutionSchema = z
  .object({
    family: z.literal("workflow-script-llm"),
    module: z.string().min(1),
    evaluator: z.string().min(1),
    budget: adaptiveBudgetSchema,
  })
  .strict();
const agentWorkflowSkillExecutionSchema = z
  .object({
    family: z.literal("agent-workflow-skill"),
    launcher: z.string().min(1),
    workflowSkill: z.string().min(1),
    oversight: autonomyOversightSchema,
    budget: adaptiveBudgetSchema,
    maximumChildDepth: z.number().int().min(0).max(1),
  })
  .strict();

export const workflowExecutionSchema = z.discriminatedUnion("family", [
  workflowScriptExecutionSchema,
  workflowScriptLlmExecutionSchema,
  agentWorkflowSkillExecutionSchema,
]);

export const workflowInvocationSchema = z
  .object({
    apiVersion: z.literal(workflowInvocationApiVersion),
    invocationId: z.string().min(1),
    trigger: z
      .object({
        source: triggerSourceSchema,
        nativeReference: nativeReferenceSchema,
        actor: z.string().min(1),
        parentInvocationReference: nativeReferenceSchema.optional(),
        childDepth: z.number().int().nonnegative().optional(),
      })
      .strict(),
    subject: z
      .object({
        reference: nativeReferenceSchema,
        revision: z.string().min(1),
      })
      .strict(),
    workflow: z
      .object({
        operation: z.string().min(1),
        profileReference: nativeReferenceSchema.optional(),
      })
      .strict(),
    execution: workflowExecutionSchema,
    evidenceProviderReference: nativeReferenceSchema,
  })
  .strict()
  .superRefine((invocation, context) => {
    const childDepth = invocation.trigger.childDepth ?? 0;
    if (
      childDepth > 0 &&
      invocation.trigger.parentInvocationReference === undefined
    ) {
      context.addIssue({
        code: "custom",
        message: "workflow-parent-invocation-reference-missing",
        path: ["trigger", "parentInvocationReference"],
      });
    }
    if (
      invocation.trigger.source === "agent" &&
      invocation.trigger.parentInvocationReference === undefined
    ) {
      context.addIssue({
        code: "custom",
        message: "workflow-agent-trigger-parent-missing",
        path: ["trigger", "parentInvocationReference"],
      });
    }
  });

export type WorkflowInvocation = z.infer<typeof workflowInvocationSchema>;
export type TriggerSource = z.infer<typeof triggerSourceSchema>;
export type AutonomyOversight = z.infer<typeof autonomyOversightSchema>;

export const workflowScriptResultSchema = z
  .object({
    outcome: z.enum(["completed", "blocked"]),
    outputReferences: z.array(nativeReferenceSchema),
    facts: z.record(z.string(), z.unknown()),
  })
  .strict();
export const modelEvaluationResultSchema = z
  .object({
    outcome: z.enum(["advise", "pass", "fail"]),
    reasons: z.array(z.string()),
    evidenceRequests: z.array(z.string()),
    usage: z
      .object({
        tokens: z.number().int().nonnegative(),
        cost: z.number().nonnegative(),
      })
      .strict(),
  })
  .strict();
export const agentWorkflowResultSchema = z
  .object({
    summary: z.string(),
    findings: z.array(z.string()),
    evidenceReferences: z.array(nativeReferenceSchema),
    usage: z
      .object({
        tokens: z.number().int().nonnegative(),
        cost: z.number().nonnegative(),
      })
      .strict(),
  })
  .strict();
const oversightVerificationSchema = z
  .object({
    level: z.enum(["A1", "A2", "A3"]),
    observed: z.literal(true),
    evidenceReferences: z.array(nativeReferenceSchema).min(1),
  })
  .strict();

export interface WorkflowRuntimePorts {
  readonly runScript: (
    request: {
      readonly module: string;
      readonly invocation: WorkflowInvocation;
    },
    signal: AbortSignal,
  ) => Promise<unknown>;
  readonly runModel: (
    request: {
      readonly evaluator: string;
      readonly invocation: WorkflowInvocation;
      readonly facts: z.infer<typeof workflowScriptResultSchema>;
    },
    signal: AbortSignal,
  ) => Promise<unknown>;
  readonly runAgent: (
    request: {
      readonly launcher: string;
      readonly workflowSkill: string;
      readonly invocation: WorkflowInvocation;
    },
    signal: AbortSignal,
  ) => Promise<unknown>;
  readonly verifyOversight: (
    invocation: WorkflowInvocation,
    signal: AbortSignal,
  ) => Promise<unknown>;
  readonly recordEvidence: (
    evidence: WorkflowExecutionEvidence,
  ) => Promise<string>;
}

export interface WorkflowExecutionEvidence {
  readonly apiVersion: typeof workflowInvocationApiVersion;
  readonly invocationId: string;
  readonly triggerSource: TriggerSource;
  readonly triggerReference: string;
  readonly subjectReference: string;
  readonly subjectRevision: string;
  readonly workflowOperation: string;
  readonly executionFamily: WorkflowInvocation["execution"]["family"];
  readonly evidenceProviderReference: string;
  readonly outcome: string;
  readonly outputReferences: readonly string[];
  readonly oversightReferences: readonly string[];
}

export interface WorkflowExecutionResponse extends WorkflowExecutionEvidence {
  readonly evidenceReference: string;
}

const withTimeout = async <Result>(
  timeoutSeconds: number,
  run: (signal: AbortSignal) => Promise<Result>,
): Promise<Result> => {
  const controller = new AbortController();
  let rejectTimeout: ((reason: Error) => void) | undefined;
  const timedOut = new Promise<never>((_resolve, reject) => {
    rejectTimeout = reject;
  });
  const timeout = setTimeout(() => {
    const error = new Error("workflow-execution-timeout");
    controller.abort(error);
    rejectTimeout?.(error);
  }, timeoutSeconds * 1_000);
  try {
    return await Promise.race([run(controller.signal), timedOut]);
  } finally {
    clearTimeout(timeout);
  }
};

const runWithRetries = async <Result>(
  retryLimit: number,
  signal: AbortSignal,
  run: () => Promise<Result>,
): Promise<Result> => {
  let failure: unknown;
  for (let attempt = 0; attempt <= retryLimit; attempt += 1) {
    if (signal.aborted) throw signal.reason;
    try {
      return await run();
    } catch (error) {
      failure = error;
    }
  }
  throw failure;
};

const validateUsage = (
  usage: {readonly tokens: number; readonly cost: number},
  budget: {readonly tokenBudget: number; readonly costBudget: number},
): void => {
  if (usage.tokens > budget.tokenBudget) {
    throw new Error("workflow-token-budget-exceeded");
  }
  if (usage.cost > budget.costBudget) {
    throw new Error("workflow-cost-budget-exceeded");
  }
};

export const executeWorkflowInvocation = async (
  untrustedInvocation: unknown,
  ports: WorkflowRuntimePorts,
): Promise<WorkflowExecutionResponse> => {
  const invocation = workflowInvocationSchema.parse(untrustedInvocation);
  const {execution} = invocation;
  let outcome: string;
  let outputReferences: readonly string[] = [];
  let oversightReferences: readonly string[] = [];

  try {
    if (execution.family === "workflow-script") {
      const result = await withTimeout(
        execution.budget.timeoutSeconds,
        (signal) =>
          ports.runScript({module: execution.module, invocation}, signal),
      );
      const parsed = workflowScriptResultSchema.parse(result);
      outcome = parsed.outcome;
      outputReferences = parsed.outputReferences;
    } else if (execution.family === "workflow-script-llm") {
      const result = await withTimeout(
        execution.budget.timeoutSeconds,
        async (signal) => {
          const facts = workflowScriptResultSchema.parse(
            await ports.runScript(
              {module: execution.module, invocation},
              signal,
            ),
          );
          const evaluation = await runWithRetries(
            execution.budget.retryLimit,
            signal,
            async () =>
              modelEvaluationResultSchema.parse(
                await ports.runModel(
                  {evaluator: execution.evaluator, invocation, facts},
                  signal,
                ),
              ),
          );
          validateUsage(evaluation.usage, execution.budget);
          return {facts, evaluation};
        },
      );
      outcome = result.evaluation.outcome;
      outputReferences = result.facts.outputReferences;
    } else {
      const childDepth = invocation.trigger.childDepth ?? 0;
      if (childDepth > execution.maximumChildDepth) {
        throw new Error("workflow-maximum-child-depth-exceeded");
      }
      const result = await withTimeout(
        execution.budget.timeoutSeconds,
        async (signal) => {
          const oversight = oversightVerificationSchema.parse(
            await ports.verifyOversight(invocation, signal),
          );
          if (oversight.level !== execution.oversight.level) {
            throw new Error("workflow-oversight-level-mismatch");
          }
          const agent = await runWithRetries(
            execution.budget.retryLimit,
            signal,
            async () =>
              agentWorkflowResultSchema.parse(
                await ports.runAgent(
                  {
                    launcher: execution.launcher,
                    workflowSkill: execution.workflowSkill,
                    invocation,
                  },
                  signal,
                ),
              ),
          );
          validateUsage(agent.usage, execution.budget);
          return {agent, oversight};
        },
      );
      outcome = "completed";
      outputReferences = result.agent.evidenceReferences;
      oversightReferences = result.oversight.evidenceReferences;
    }
  } catch (error) {
    const code = error instanceof Error ? error.message : String(error);
    const evidenceReference = await ports.recordEvidence({
      apiVersion: workflowInvocationApiVersion,
      invocationId: invocation.invocationId,
      triggerSource: invocation.trigger.source,
      triggerReference: invocation.trigger.nativeReference,
      subjectReference: invocation.subject.reference,
      subjectRevision: invocation.subject.revision,
      workflowOperation: invocation.workflow.operation,
      executionFamily: execution.family,
      evidenceProviderReference: invocation.evidenceProviderReference,
      outcome: `failed:${code}`,
      outputReferences: [],
      oversightReferences: [],
    });
    if (evidenceReference.length === 0) {
      throw new Error("workflow-evidence-reference-missing", {cause: error});
    }
    throw new Error(
      `workflow-execution-failed:${code}:evidence:${evidenceReference}`,
      {cause: error},
    );
  }

  const evidence: WorkflowExecutionEvidence = {
    apiVersion: workflowInvocationApiVersion,
    invocationId: invocation.invocationId,
    triggerSource: invocation.trigger.source,
    triggerReference: invocation.trigger.nativeReference,
    subjectReference: invocation.subject.reference,
    subjectRevision: invocation.subject.revision,
    workflowOperation: invocation.workflow.operation,
    executionFamily: execution.family,
    evidenceProviderReference: invocation.evidenceProviderReference,
    outcome,
    outputReferences,
    oversightReferences,
  };
  const evidenceReference = await ports.recordEvidence(evidence);
  if (evidenceReference.length === 0) {
    throw new Error("workflow-evidence-reference-missing");
  }
  return {...evidence, evidenceReference};
};
