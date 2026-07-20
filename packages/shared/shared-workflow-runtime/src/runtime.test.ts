import {describe, expect, it} from "vitest";
import {
  executeWorkflowInvocation,
  explainWorkflowComposition,
  workflowInvocationApiVersion,
  type WorkflowControl,
  type WorkflowEvidenceEvent,
  type WorkflowInvocation,
  type WorkflowPluginRegistry,
} from "./runtime.js";

const baseInvocation = (): WorkflowInvocation => ({
  apiVersion: workflowInvocationApiVersion,
  invocationId: "invocation-1",
  trigger: {
    kind: "manual",
    reference: "terminal:session-1",
    actor: "human:maintainer",
  },
  subject: {reference: "repository:xonovex", revision: "commit:abc123"},
  operation: "review",
  executor: {plugin: "review"},
  controls: [],
  evidence: [],
  requiredCapabilities: [],
  metadata: {},
});

const control = (
  id: string,
  decision: "allow" | "deny" | "abstain",
  capabilities: readonly string[] = [],
): WorkflowControl => ({
  id,
  capabilities,
  phases: ["before"],
  evaluate: () => Promise.resolve({decision, references: [`control:${id}`]}),
});

const registry = (
  overrides: Partial<WorkflowPluginRegistry> = {},
): WorkflowPluginRegistry => ({
  executors: {
    review: {
      id: "review",
      capabilities: ["execution:script"],
      execute: () =>
        Promise.resolve({
          outcome: "completed",
          references: ["result:review-1"],
        }),
    },
  },
  controls: {},
  evidenceSinks: {},
  ...overrides,
});

describe("executeWorkflowInvocation", () => {
  it("keeps an arbitrary trigger independent from executor selection", async () => {
    const invocation = baseInvocation();
    invocation.trigger = {
      kind: "ci/github/pull-request",
      reference: "github:delivery-42",
    };

    const result = await executeWorkflowInvocation(invocation, registry());

    expect(result.outcome).toBe("executed");
    expect(result.execution?.outcome).toBe("completed");
  });

  it("records an observing denial without blocking execution", async () => {
    const invocation = baseInvocation();
    invocation.controls = [{plugin: "review-policy", mode: "observe"}];

    const result = await executeWorkflowInvocation(
      invocation,
      registry({controls: {"review-policy": control("review-policy", "deny")}}),
    );

    expect(result.outcome).toBe("executed");
    expect(result.controls[0]?.result.decision).toBe("deny");
  });

  it("stops before execution when an enforcing control denies", async () => {
    let executions = 0;
    const invocation = baseInvocation();
    invocation.controls = [{plugin: "review-policy", mode: "enforce"}];

    const result = await executeWorkflowInvocation(
      invocation,
      registry({
        executors: {
          review: {
            id: "review",
            capabilities: [],
            execute: () => {
              executions += 1;
              return Promise.resolve({outcome: "completed"});
            },
          },
        },
        controls: {"review-policy": control("review-policy", "deny")},
      }),
    );

    expect(result.outcome).toBe("denied");
    expect(result.execution).toBeUndefined();
    expect(executions).toBe(0);
  });

  it("fails before execution when a required capability is absent", async () => {
    const invocation = baseInvocation();
    invocation.requiredCapabilities = ["oversight:independent-review"];

    await expect(
      executeWorkflowInvocation(invocation, registry()),
    ).rejects.toThrow(
      /workflow-required-capabilities-missing:oversight:independent-review/u,
    );
  });

  it("routes lifecycle events only to explicitly selected evidence sinks", async () => {
    const events: WorkflowEvidenceEvent[] = [];
    const invocation = baseInvocation();
    invocation.evidence = [{plugin: "memory", failure: "fail"}];

    const result = await executeWorkflowInvocation(
      invocation,
      registry({
        evidenceSinks: {
          memory: {
            id: "memory",
            capabilities: ["evidence:memory"],
            record: (event) => {
              events.push(event);
              return Promise.resolve(`memory:${event.kind}`);
            },
          },
        },
      }),
    );

    expect(events.map(({kind}) => kind)).toEqual([
      "composition.started",
      "execution.completed",
      "composition.completed",
    ]);
    expect(result.evidenceReferences).toEqual([
      "memory:composition.completed",
      "memory:composition.started",
      "memory:execution.completed",
      "result:review-1",
    ]);
  });
});

describe("explainWorkflowComposition", () => {
  it("shows configured enforcement points and missing capabilities without running", () => {
    const invocation = baseInvocation();
    invocation.controls = [{plugin: "approval", mode: "enforce"}];
    invocation.requiredCapabilities = ["control:approval", "evidence:durable"];

    const explanation = explainWorkflowComposition(
      invocation,
      registry({
        controls: {
          approval: control("approval", "allow", ["control:approval"]),
        },
      }),
    );

    expect(explanation.enforcementPoints).toEqual(["before:approval"]);
    expect(explanation.missingCapabilities).toEqual(["evidence:durable"]);
  });
});
