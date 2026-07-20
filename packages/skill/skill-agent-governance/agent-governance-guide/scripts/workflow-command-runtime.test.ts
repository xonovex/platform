import assert from "node:assert/strict";
import {readFile, rm, writeFile} from "node:fs/promises";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {describe, it} from "node:test";
import {
  createCommandPluginRegistry,
  type WorkflowCommandRegistry,
} from "./workflow-command-runtime.ts";
import {
  executeWorkflowInvocation,
  workflowInvocationApiVersion,
  type WorkflowInvocation,
} from "./workflow-runtime.ts";

type ExecutorDefinition = WorkflowCommandRegistry["executors"][string];
type ScriptCommand = Extract<
  ExecutorDefinition,
  {adapter: "script"}
>["command"];

const jsonCommand = (result: unknown): ScriptCommand => ({
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
    kind: "provider/github/webhook",
    reference: "github:delivery-42",
  },
  subject: {reference: "repository:xonovex", revision: "commit:abc123"},
  operation: "review",
  executor: {plugin: "review"},
  controls: [],
  evidence: [],
  requiredCapabilities: [],
  metadata: {},
});

describe("createCommandPluginRegistry", () => {
  it("runs a script and model composition through one executor plugin", async () => {
    const evidencePath = join(
      tmpdir(),
      `xonovex-workflow-${String(process.pid)}-${String(Date.now())}.jsonl`,
    );
    const registry: WorkflowCommandRegistry = {
      executors: {
        review: {
          adapter: "script-llm",
          script: jsonCommand({failedChecks: 1}),
          model: jsonCommand({
            outcome: "advise",
            output: {reasons: ["one check failed"]},
            references: ["artifact:facts-1"],
          }),
          capabilities: ["execution:script", "execution:llm"],
        },
      },
      controls: {},
      evidenceSinks: {
        local: {
          adapter: "jsonl",
          path: evidencePath,
          capabilities: ["evidence:local"],
        },
      },
    };
    const invocation = baseInvocation();
    invocation.evidence = [{plugin: "local", failure: "fail"}];

    const result = await executeWorkflowInvocation(
      invocation,
      await createCommandPluginRegistry(registry),
    );

    assert.equal(result.execution?.outcome, "advise");
    assert.equal(result.evidenceReferences.length, 4);
    assert.equal(
      (await readFile(evidencePath, "utf8")).trim().split("\n").length,
      3,
    );
  });

  it("launches an agent with a workflow skill without embedding a maturity level", async () => {
    const registry: WorkflowCommandRegistry = {
      executors: {
        agent: {
          adapter: "agent-skill",
          agent: jsonCommand({
            outcome: "completed",
            references: ["artifact:agent-1"],
          }),
          skill: "workflow/review",
          capabilities: ["execution:agent", "workflow-skill:review"],
        },
      },
      controls: {},
      evidenceSinks: {},
    };
    const invocation = baseInvocation();
    invocation.executor = {plugin: "agent"};

    const result = await executeWorkflowInvocation(
      invocation,
      await createCommandPluginRegistry(registry),
    );

    assert.equal(result.outcome, "executed");
    assert.deepEqual(result.execution?.references, ["artifact:agent-1"]);
  });

  it("keeps a command control independent from the executor adapter", async () => {
    const registry: WorkflowCommandRegistry = {
      executors: {
        review: {
          adapter: "script",
          command: jsonCommand({outcome: "completed", references: []}),
          capabilities: [],
        },
      },
      controls: {
        approval: {
          command: jsonCommand({
            decision: "deny",
            reason: "approval-missing",
            references: ["approval:missing"],
          }),
          phases: ["before"],
          capabilities: ["oversight:approval"],
        },
      },
      evidenceSinks: {},
    };
    const invocation = baseInvocation();
    invocation.controls = [{plugin: "approval", mode: "enforce"}];

    const result = await executeWorkflowInvocation(
      invocation,
      await createCommandPluginRegistry(registry),
    );

    assert.equal(result.outcome, "denied");
    assert.equal(result.execution, undefined);
  });

  it("does not initialize an unselected evidence sink", async (context) => {
    const evidencePath = join(
      tmpdir(),
      `xonovex-unused-evidence-${String(process.pid)}-${String(Date.now())}.jsonl`,
    );
    await writeFile(evidencePath, "not-json\n", "utf8");
    context.after(async () => await rm(evidencePath, {force: true}));
    const registry: WorkflowCommandRegistry = {
      executors: {
        review: {
          adapter: "script",
          command: jsonCommand({outcome: "completed", references: []}),
          capabilities: [],
        },
      },
      controls: {},
      evidenceSinks: {
        unused: {
          adapter: "jsonl",
          path: evidencePath,
          capabilities: [],
        },
      },
    };

    const result = await executeWorkflowInvocation(
      baseInvocation(),
      await createCommandPluginRegistry(registry),
    );

    assert.equal(result.outcome, "executed");
  });
});
