import assert from "node:assert/strict";
import {readFile, writeFile} from "node:fs/promises";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {describe, it} from "node:test";
import {
  createCommandRuntimePorts,
  createJsonlWorkflowEvidenceRecorder,
  type WorkflowCommandRegistry,
} from "./workflow-command-runtime.ts";
import {
  executeWorkflowInvocation,
  workflowInvocationApiVersion,
  type WorkflowExecutionEvidence,
  type WorkflowInvocation,
} from "./workflow-runtime.ts";

const jsonCommand = (
  result: unknown,
): WorkflowCommandRegistry["scripts"][string] => ({
  executable: process.execPath,
  arguments: [
    "-e",
    `process.stdin.resume();process.stdin.on("end",()=>console.log(${JSON.stringify(JSON.stringify(result))}))`,
  ],
  environment: {},
});

const baseInvocation = (): WorkflowInvocation => ({
  apiVersion: workflowInvocationApiVersion,
  invocationId: `invocation-${String(process.pid)}-${String(Date.now())}`,
  trigger: {
    source: "provider-webhook",
    nativeReference: "github:delivery-42",
    actor: "external:github",
  },
  subject: {reference: "repository:xonovex", revision: "commit:abc123"},
  workflow: {operation: "review-run"},
  execution: {
    family: "workflow-script-llm",
    module: "facts/1",
    evaluator: "review/1",
    budget: {
      timeoutSeconds: 5,
      tokenBudget: 500,
      costBudget: 1,
      retryLimit: 0,
    },
  },
  evidenceProviderReference: "evidence:local-test",
});

describe("createCommandRuntimePorts", () => {
  it("rejects conflicting invocation evidence already present on disk", async () => {
    const evidencePath = join(
      tmpdir(),
      `xonovex-workflow-collision-${String(process.pid)}-${String(Date.now())}.jsonl`,
    );
    const first = {
      apiVersion: workflowInvocationApiVersion,
      invocationId: "invocation-collision",
      triggerSource: "manual",
      triggerReference: "terminal:session-1",
      subjectReference: "repository:xonovex",
      subjectRevision: "commit:abc123",
      workflowOperation: "review-run",
      executionFamily: "workflow-script",
      evidenceProviderReference: "evidence:local-test",
      outcome: "completed",
      outputReferences: [],
      oversightReferences: [],
    } satisfies WorkflowExecutionEvidence;
    await writeFile(
      evidencePath,
      `${JSON.stringify(first)}\n${JSON.stringify({...first, outcome: "blocked"})}\n`,
      "utf8",
    );

    await assert.rejects(
      createJsonlWorkflowEvidenceRecorder(evidencePath),
      /workflow-evidence-log-invocation-collision/u,
    );
  });

  it("runs a registered deterministic script before a registered bounded model", async () => {
    const evidencePath = join(
      tmpdir(),
      `xonovex-workflow-${String(process.pid)}-${String(Date.now())}.jsonl`,
    );
    const recorder = await createJsonlWorkflowEvidenceRecorder(evidencePath);
    const registry: WorkflowCommandRegistry = {
      scripts: {
        "facts/1": jsonCommand({
          outcome: "completed",
          outputReferences: ["artifact:facts-1"],
          facts: {failedChecks: 1},
        }),
      },
      models: {
        "review/1": jsonCommand({
          outcome: "advise",
          reasons: ["one check failed"],
          evidenceRequests: ["artifact:facts-1"],
          usage: {tokens: 100, cost: 0.1},
        }),
      },
      agents: {},
      oversightVerifiers: {},
    };

    const result = await executeWorkflowInvocation(
      baseInvocation(),
      createCommandRuntimePorts(registry, recorder),
    );

    assert.equal(result.outcome, "advise");
    assert.equal(result.triggerSource, "provider-webhook");
    assert.match(result.evidenceReference, /#sha256:[0-9a-f]{64}$/u);
    assert.equal(
      (await readFile(evidencePath, "utf8")).trim().split("\n").length,
      1,
    );
  });

  it("launches an agent only after an A2 verifier observes the workflow skill controls", async () => {
    const evidencePath = join(
      tmpdir(),
      `xonovex-agent-workflow-${String(process.pid)}-${String(Date.now())}.jsonl`,
    );
    const recorder = await createJsonlWorkflowEvidenceRecorder(evidencePath);
    const registry: WorkflowCommandRegistry = {
      scripts: {},
      models: {},
      agents: {
        "agent/1": jsonCommand({
          summary: "skill completed",
          findings: [],
          evidenceReferences: ["artifact:agent-1"],
          usage: {tokens: 200, cost: 0.2},
        }),
      },
      oversightVerifiers: {
        A2: jsonCommand({
          level: "A2",
          observed: true,
          evidenceReferences: ["oversight:observed-a2"],
        }),
      },
    };
    const invocation = baseInvocation();
    invocation.execution = {
      family: "agent-workflow-skill",
      launcher: "agent/1",
      workflowSkill: "workflow-skill:review/1",
      oversight: {
        level: "A2",
        independentCritiqueReference: "critique:1",
        journalReference: "journal:1",
        approvalReference: "approval:1",
        cancellationReference: "cancel:1",
        killSwitchReference: "kill:1",
      },
      budget: {
        timeoutSeconds: 5,
        tokenBudget: 500,
        costBudget: 1,
        retryLimit: 0,
      },
      maximumChildDepth: 0,
    };

    const result = await executeWorkflowInvocation(
      invocation,
      createCommandRuntimePorts(registry, recorder),
    );

    assert.deepEqual(result.oversightReferences, ["oversight:observed-a2"]);
    assert.deepEqual(result.outputReferences, ["artifact:agent-1"]);
  });
});
