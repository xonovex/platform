import {z} from "zod";

export const workflowInvocationApiVersion = "workflow.xonovex.com/v1" as const;

const referenceSchema = z.string().min(1);
const pluginSelectionSchema = z
  .object({
    plugin: z.string().min(1),
    input: z.unknown().optional(),
  })
  .strict();

export const controlModeSchema = z.enum(["observe", "enforce"]);
export const controlPhaseSchema = z.enum(["before", "after"]);

export const controlSelectionSchema = pluginSelectionSchema
  .extend({mode: controlModeSchema})
  .strict();

export const evidenceSelectionSchema = pluginSelectionSchema
  .extend({failure: z.enum(["ignore", "fail"])})
  .strict();

export const workflowInvocationSchema = z
  .object({
    apiVersion: z.literal(workflowInvocationApiVersion),
    invocationId: referenceSchema,
    trigger: z
      .object({
        kind: z.string().min(1),
        reference: referenceSchema,
        actor: z.string().min(1).optional(),
        data: z.unknown().optional(),
      })
      .strict(),
    subject: z
      .object({
        reference: referenceSchema,
        revision: z.string().min(1).optional(),
      })
      .strict(),
    operation: z.string().min(1),
    executor: pluginSelectionSchema,
    controls: z.array(controlSelectionSchema).default([]),
    evidence: z.array(evidenceSelectionSchema).default([]),
    requiredCapabilities: z.array(z.string().min(1)).default([]),
    metadata: z.record(z.string(), z.unknown()).default({}),
  })
  .strict();

export const executorResultSchema = z
  .object({
    outcome: z.string().min(1),
    output: z.unknown().optional(),
    references: z.array(referenceSchema).default([]),
  })
  .strict();

export const controlResultSchema = z
  .object({
    decision: z.enum(["allow", "deny", "abstain"]),
    reason: z.string().min(1).optional(),
    references: z.array(referenceSchema).default([]),
    data: z.unknown().optional(),
  })
  .strict();

export type WorkflowInvocation = z.infer<typeof workflowInvocationSchema>;
export type ExecutorResult = z.infer<typeof executorResultSchema>;
export type ControlResult = z.infer<typeof controlResultSchema>;
export type ControlMode = z.infer<typeof controlModeSchema>;
export type ControlPhase = z.infer<typeof controlPhaseSchema>;

export interface WorkflowExecutor {
  readonly id: string;
  readonly capabilities: readonly string[];
  readonly execute: (
    invocation: WorkflowInvocation,
    input: unknown,
    signal: AbortSignal,
  ) => Promise<unknown>;
}

export interface WorkflowControl {
  readonly id: string;
  readonly capabilities: readonly string[];
  readonly phases: readonly ControlPhase[];
  readonly evaluate: (
    request: {
      readonly phase: ControlPhase;
      readonly invocation: WorkflowInvocation;
      readonly execution?: ExecutorResult;
      readonly input: unknown;
    },
    signal: AbortSignal,
  ) => Promise<unknown>;
}

export interface WorkflowEvidenceEvent {
  readonly apiVersion: typeof workflowInvocationApiVersion;
  readonly invocationId: string;
  readonly kind:
    | "composition.started"
    | "control.evaluated"
    | "execution.completed"
    | "composition.completed"
    | "composition.failed";
  readonly payload: Readonly<Record<string, unknown>>;
}

export interface WorkflowEvidenceSink {
  readonly id: string;
  readonly capabilities: readonly string[];
  readonly record: (
    event: WorkflowEvidenceEvent,
    input: unknown,
  ) => Promise<string | undefined>;
}

export interface WorkflowPluginRegistry {
  readonly executors: Readonly<Record<string, WorkflowExecutor>>;
  readonly controls: Readonly<Record<string, WorkflowControl>>;
  readonly evidenceSinks: Readonly<Record<string, WorkflowEvidenceSink>>;
}

interface ResolvedControl {
  readonly plugin: WorkflowControl;
  readonly mode: ControlMode;
  readonly input: unknown;
}

interface ResolvedEvidenceSink {
  readonly plugin: WorkflowEvidenceSink;
  readonly failure: "ignore" | "fail";
  readonly input: unknown;
}

interface ResolvedComposition {
  readonly invocation: WorkflowInvocation;
  readonly executor: WorkflowExecutor;
  readonly controls: readonly ResolvedControl[];
  readonly evidenceSinks: readonly ResolvedEvidenceSink[];
  readonly capabilities: readonly string[];
  readonly missingCapabilities: readonly string[];
}

export interface WorkflowControlObservation {
  readonly plugin: string;
  readonly phase: ControlPhase;
  readonly mode: ControlMode;
  readonly result: ControlResult;
}

export interface WorkflowExecutionResponse {
  readonly invocationId: string;
  readonly outcome: "executed" | "denied";
  readonly execution?: ExecutorResult;
  readonly controls: readonly WorkflowControlObservation[];
  readonly evidenceReferences: readonly string[];
}

export interface WorkflowCompositionExplanation {
  readonly invocationId: string;
  readonly trigger: string;
  readonly executor: {
    readonly plugin: string;
    readonly capabilities: readonly string[];
  };
  readonly controls: readonly {
    readonly plugin: string;
    readonly mode: ControlMode;
    readonly phases: readonly ControlPhase[];
    readonly capabilities: readonly string[];
  }[];
  readonly evidenceSinks: readonly {
    readonly plugin: string;
    readonly failure: "ignore" | "fail";
    readonly capabilities: readonly string[];
  }[];
  readonly requiredCapabilities: readonly string[];
  readonly availableCapabilities: readonly string[];
  readonly missingCapabilities: readonly string[];
  readonly enforcementPoints: readonly string[];
}

const lookupPlugin = <Plugin>(
  plugins: Readonly<Record<string, Plugin>>,
  id: string,
  kind: string,
): Plugin => {
  const plugin = plugins[id];
  if (plugin === undefined) {
    throw new Error(`workflow-${kind}-not-registered:${id}`);
  }
  return plugin;
};

const uniqueSorted = (values: readonly string[]): readonly string[] =>
  [...new Set(values)].toSorted();

const resolveComposition = (
  untrustedInvocation: unknown,
  registry: WorkflowPluginRegistry,
): ResolvedComposition => {
  const invocation = workflowInvocationSchema.parse(untrustedInvocation);
  const executor = lookupPlugin(
    registry.executors,
    invocation.executor.plugin,
    "executor",
  );
  const controls = invocation.controls.map((selection) => ({
    plugin: lookupPlugin(registry.controls, selection.plugin, "control"),
    mode: selection.mode,
    input: selection.input,
  }));
  const evidenceSinks = invocation.evidence.map((selection) => ({
    plugin: lookupPlugin(
      registry.evidenceSinks,
      selection.plugin,
      "evidence-sink",
    ),
    failure: selection.failure,
    input: selection.input,
  }));
  const capabilities = uniqueSorted([
    ...executor.capabilities,
    ...controls.flatMap(({plugin}) => plugin.capabilities),
    ...evidenceSinks.flatMap(({plugin}) => plugin.capabilities),
  ]);
  const available = new Set(capabilities);
  const missingCapabilities = uniqueSorted(
    invocation.requiredCapabilities.filter(
      (capability) => !available.has(capability),
    ),
  );
  return {
    invocation,
    executor,
    controls,
    evidenceSinks,
    capabilities,
    missingCapabilities,
  };
};

export const explainWorkflowComposition = (
  untrustedInvocation: unknown,
  registry: WorkflowPluginRegistry,
): WorkflowCompositionExplanation => {
  const composition = resolveComposition(untrustedInvocation, registry);
  return {
    invocationId: composition.invocation.invocationId,
    trigger: composition.invocation.trigger.kind,
    executor: {
      plugin: composition.executor.id,
      capabilities: uniqueSorted(composition.executor.capabilities),
    },
    controls: composition.controls.map(({plugin, mode}) => ({
      plugin: plugin.id,
      mode,
      phases: uniqueSorted(plugin.phases) as readonly ControlPhase[],
      capabilities: uniqueSorted(plugin.capabilities),
    })),
    evidenceSinks: composition.evidenceSinks.map(({plugin, failure}) => ({
      plugin: plugin.id,
      failure,
      capabilities: uniqueSorted(plugin.capabilities),
    })),
    requiredCapabilities: uniqueSorted(
      composition.invocation.requiredCapabilities,
    ),
    availableCapabilities: composition.capabilities,
    missingCapabilities: composition.missingCapabilities,
    enforcementPoints: composition.controls.flatMap(({plugin, mode}) =>
      mode === "enforce"
        ? plugin.phases.map((phase) => `${phase}:${plugin.id}`)
        : [],
    ),
  };
};

const recordEvidence = async (
  composition: ResolvedComposition,
  event: WorkflowEvidenceEvent,
): Promise<readonly string[]> => {
  const references: string[] = [];
  for (const sink of composition.evidenceSinks) {
    try {
      const reference = await sink.plugin.record(event, sink.input);
      if (reference !== undefined && reference !== "") {
        references.push(reference);
      }
    } catch (error) {
      if (sink.failure === "fail") throw error;
    }
  }
  return references;
};

const runControls = async (
  composition: ResolvedComposition,
  phase: ControlPhase,
  execution: ExecutorResult | undefined,
  signal: AbortSignal,
): Promise<{
  readonly observations: readonly WorkflowControlObservation[];
  readonly evidenceReferences: readonly string[];
  readonly denied: boolean;
}> => {
  const observations: WorkflowControlObservation[] = [];
  const evidenceReferences: string[] = [];
  let denied = false;
  for (const control of composition.controls) {
    if (!control.plugin.phases.includes(phase)) continue;
    const result = controlResultSchema.parse(
      await control.plugin.evaluate(
        {
          phase,
          invocation: composition.invocation,
          execution,
          input: control.input,
        },
        signal,
      ),
    );
    observations.push({
      plugin: control.plugin.id,
      phase,
      mode: control.mode,
      result,
    });
    evidenceReferences.push(
      ...result.references,
      ...(await recordEvidence(composition, {
        apiVersion: workflowInvocationApiVersion,
        invocationId: composition.invocation.invocationId,
        kind: "control.evaluated",
        payload: {
          plugin: control.plugin.id,
          phase,
          mode: control.mode,
          decision: result.decision,
          references: result.references,
          ...(result.reason === undefined ? {} : {reason: result.reason}),
        },
      })),
    );
    if (control.mode === "enforce" && result.decision === "deny") {
      denied = true;
      break;
    }
  }
  return {observations, evidenceReferences, denied};
};

const inactiveSignal = new AbortController().signal;

export const executeWorkflowInvocation = async (
  untrustedInvocation: unknown,
  registry: WorkflowPluginRegistry,
  signal: AbortSignal = inactiveSignal,
): Promise<WorkflowExecutionResponse> => {
  const composition = resolveComposition(untrustedInvocation, registry);
  if (composition.missingCapabilities.length > 0) {
    throw new Error(
      `workflow-required-capabilities-missing:${composition.missingCapabilities.join(",")}`,
    );
  }
  const evidenceReferences: string[] = [];
  const observations: WorkflowControlObservation[] = [];

  try {
    evidenceReferences.push(
      ...(await recordEvidence(composition, {
        apiVersion: workflowInvocationApiVersion,
        invocationId: composition.invocation.invocationId,
        kind: "composition.started",
        payload: {
          trigger: composition.invocation.trigger.kind,
          executor: composition.executor.id,
        },
      })),
    );

    const before = await runControls(composition, "before", undefined, signal);
    observations.push(...before.observations);
    evidenceReferences.push(...before.evidenceReferences);
    if (before.denied) {
      evidenceReferences.push(
        ...(await recordEvidence(composition, {
          apiVersion: workflowInvocationApiVersion,
          invocationId: composition.invocation.invocationId,
          kind: "composition.completed",
          payload: {outcome: "denied", phase: "before"},
        })),
      );
      return {
        invocationId: composition.invocation.invocationId,
        outcome: "denied",
        controls: observations,
        evidenceReferences: uniqueSorted(evidenceReferences),
      };
    }

    const execution = executorResultSchema.parse(
      await composition.executor.execute(
        composition.invocation,
        composition.invocation.executor.input,
        signal,
      ),
    );
    evidenceReferences.push(
      ...execution.references,
      ...(await recordEvidence(composition, {
        apiVersion: workflowInvocationApiVersion,
        invocationId: composition.invocation.invocationId,
        kind: "execution.completed",
        payload: {
          executor: composition.executor.id,
          outcome: execution.outcome,
          references: execution.references,
        },
      })),
    );

    const after = await runControls(composition, "after", execution, signal);
    observations.push(...after.observations);
    evidenceReferences.push(...after.evidenceReferences);
    const outcome = after.denied ? "denied" : "executed";
    evidenceReferences.push(
      ...(await recordEvidence(composition, {
        apiVersion: workflowInvocationApiVersion,
        invocationId: composition.invocation.invocationId,
        kind: "composition.completed",
        payload: {outcome},
      })),
    );
    return {
      invocationId: composition.invocation.invocationId,
      outcome,
      execution,
      controls: observations,
      evidenceReferences: uniqueSorted(evidenceReferences),
    };
  } catch (error) {
    const failure = error instanceof Error ? error.message : String(error);
    await recordEvidence(composition, {
      apiVersion: workflowInvocationApiVersion,
      invocationId: composition.invocation.invocationId,
      kind: "composition.failed",
      payload: {failure},
    });
    throw new Error(`workflow-execution-failed:${failure}`, {cause: error});
  }
};
