import assert from "node:assert/strict";
import {describe, it} from "node:test";
import {
  executeWorkflowInvocation,
  workflowInvocationApiVersion,
  type WorkflowExecutionEvidence,
  type WorkflowInvocation,
  type WorkflowRuntimePorts,
} from "./workflow-runtime.ts";

const baseInvocation = (): WorkflowInvocation => ({
  apiVersion: workflowInvocationApiVersion,
  invocationId: "invocation-1",
  trigger: {
    source: "manual",
    nativeReference: "terminal:session-1",
    actor: "human:maintainer",
  },
  subject: {reference: "repository:xonovex", revision: "commit:abc123"},
  workflow: {operation: "review-run"},
  execution: {
    family: "workflow-script",
    module: "review-script/1",
    budget: {timeoutSeconds: 30},
  },
  evidenceProviderReference: "evidence:local-test",
});

const runtimePorts = (
  overrides: Partial<WorkflowRuntimePorts> = {},
): WorkflowRuntimePorts => ({
  runScript: async () => ({
    outcome: "completed",
    outputReferences: ["result:script-1"],
    facts: {checks: 3},
  }),
  runModel: async () => ({
    outcome: "advise",
    reasons: ["review one failed check"],
    evidenceRequests: ["ci:check-1"],
    usage: {tokens: 200, cost: 0.1},
  }),
  runAgent: async () => ({
    summary: "reviewed exact revision",
    findings: [],
    evidenceReferences: ["result:agent-1"],
    usage: {tokens: 800, cost: 0.5},
  }),
  verifyOversight: async () => ({
    level: "A1",
    observed: true,
    evidenceReferences: ["oversight:critique-1"],
  }),
  recordEvidence: async (evidence: WorkflowExecutionEvidence) =>
    `evidence:${evidence.invocationId}`,
  ...overrides,
});

describe("executeWorkflowInvocation", () => {
  it("preserves a CI/CD hook as origin while a workflow script executes", async () => {
    const invocation = baseInvocation();
    invocation.trigger = {
      source: "ci-cd-hook",
      nativeReference: "github-actions:run-42",
      actor: "external:github-actions",
    };

    const result = await executeWorkflowInvocation(invocation, runtimePorts());

    assert.equal(result.triggerSource, "ci-cd-hook");
    assert.equal(result.executionFamily, "workflow-script");
    assert.deepEqual(result.outputReferences, ["result:script-1"]);
  });

  it("uses deterministic facts before bounded model evaluation", async () => {
    const invocation = baseInvocation();
    invocation.execution = {
      family: "workflow-script-llm",
      module: "assessment-script/1",
      evaluator: "bounded-assessment/1",
      budget: {
        timeoutSeconds: 30,
        tokenBudget: 500,
        costBudget: 1,
        retryLimit: 0,
      },
    };

    const result = await executeWorkflowInvocation(invocation, runtimePorts());

    assert.equal(result.outcome, "advise");
    assert.deepEqual(result.outputReferences, ["result:script-1"]);
  });

  it("requires observed A1 oversight before launching an agent with a workflow skill", async () => {
    const invocation = baseInvocation();
    invocation.execution = {
      family: "agent-workflow-skill",
      launcher: "bounded-agent/1",
      workflowSkill: "workflow-guide",
      oversight: {
        level: "A1",
        independentCritiqueReference: "review:independent-1",
      },
      budget: {
        timeoutSeconds: 60,
        tokenBudget: 1_000,
        costBudget: 1,
        retryLimit: 0,
      },
      maximumChildDepth: 0,
    };

    const result = await executeWorkflowInvocation(invocation, runtimePorts());

    assert.equal(result.executionFamily, "agent-workflow-skill");
    assert.deepEqual(result.oversightReferences, ["oversight:critique-1"]);
  });

  it("rejects an agent-originated child beyond the declared depth", async () => {
    const invocation = baseInvocation();
    invocation.trigger = {
      source: "agent",
      nativeReference: "agent-run:parent-1",
      actor: "agent:parent-1",
      parentInvocationReference: "workflow:parent-1",
      childDepth: 1,
    };
    invocation.execution = {
      family: "agent-workflow-skill",
      launcher: "bounded-agent/1",
      workflowSkill: "workflow-guide",
      oversight: {
        level: "A1",
        independentCritiqueReference: "review:independent-1",
      },
      budget: {
        timeoutSeconds: 60,
        tokenBudget: 1_000,
        costBudget: 1,
        retryLimit: 0,
      },
      maximumChildDepth: 0,
    };

    await assert.rejects(
      executeWorkflowInvocation(invocation, runtimePorts()),
      /workflow-maximum-child-depth-exceeded/u,
    );
  });

  it("rejects a model result that exceeds its declared token budget", async () => {
    const invocation = baseInvocation();
    invocation.execution = {
      family: "workflow-script-llm",
      module: "assessment-script/1",
      evaluator: "bounded-assessment/1",
      budget: {
        timeoutSeconds: 30,
        tokenBudget: 100,
        costBudget: 1,
        retryLimit: 0,
      },
    };

    await assert.rejects(
      executeWorkflowInvocation(invocation, runtimePorts()),
      /workflow-token-budget-exceeded/u,
    );
  });

  it("times out an uncooperative executor and records failed evidence", async () => {
    const invocation = baseInvocation();
    invocation.execution = {
      family: "workflow-script",
      module: "uncooperative-script/1",
      budget: {timeoutSeconds: 1},
    };
    let recorded: WorkflowExecutionEvidence | undefined;

    await assert.rejects(
      executeWorkflowInvocation(
        invocation,
        runtimePorts({
          runScript: async () => await new Promise<never>(() => undefined),
          recordEvidence: async (evidence) => {
            recorded = evidence;
            return "evidence:timeout-1";
          },
        }),
      ),
      /workflow-execution-timeout/u,
    );
    assert.equal(recorded?.outcome, "failed:workflow-execution-timeout");
    assert.equal(recorded?.triggerReference, "terminal:session-1");
  });
});
