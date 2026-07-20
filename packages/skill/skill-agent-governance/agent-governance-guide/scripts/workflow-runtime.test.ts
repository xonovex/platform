import assert from "node:assert/strict";
import {describe, it} from "node:test";
import {
  executeWorkflowInvocation,
  explainWorkflowComposition,
  workflowInvocationApiVersion,
  type WorkflowControl,
  type WorkflowEvidenceEvent,
  type WorkflowInvocation,
  type WorkflowPluginRegistry,
} from "./workflow-runtime.ts";

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
  evaluate: async () => ({decision, references: [`control:${id}`]}),
});

const registry = (
  overrides: Partial<WorkflowPluginRegistry> = {},
): WorkflowPluginRegistry => ({
  executors: {
    review: {
      id: "review",
      capabilities: ["execution:script"],
      execute: async () => ({
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

    assert.equal(result.outcome, "executed");
    assert.equal(result.execution?.outcome, "completed");
  });

  it("records an observing denial without blocking execution", async () => {
    const invocation = baseInvocation();
    invocation.controls = [{plugin: "review-policy", mode: "observe"}];

    const result = await executeWorkflowInvocation(
      invocation,
      registry({controls: {"review-policy": control("review-policy", "deny")}}),
    );

    assert.equal(result.outcome, "executed");
    assert.equal(result.controls[0]?.result.decision, "deny");
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
            execute: async () => {
              executions += 1;
              return {outcome: "completed"};
            },
          },
        },
        controls: {"review-policy": control("review-policy", "deny")},
      }),
    );

    assert.equal(result.outcome, "denied");
    assert.equal(result.execution, undefined);
    assert.equal(executions, 0);
  });

  it("fails before execution when a required capability is absent", async () => {
    const invocation = baseInvocation();
    invocation.requiredCapabilities = ["oversight:independent-review"];

    await assert.rejects(
      executeWorkflowInvocation(invocation, registry()),
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
            record: async (event) => {
              events.push(event);
              return `memory:${event.kind}`;
            },
          },
        },
      }),
    );

    assert.deepEqual(
      events.map(({kind}) => kind),
      ["composition.started", "execution.completed", "composition.completed"],
    );
    assert.deepEqual(result.evidenceReferences, [
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

    assert.deepEqual(explanation.enforcementPoints, ["before:approval"]);
    assert.deepEqual(explanation.missingCapabilities, ["evidence:durable"]);
  });
});
