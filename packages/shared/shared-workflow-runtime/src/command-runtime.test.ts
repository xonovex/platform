import {readFile, rm, writeFile} from "node:fs/promises";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {describe, expect, it, onTestFinished} from "vitest";
import {
  createCommandPluginRegistry,
  type WorkflowCommandRegistry,
} from "./command-runtime.js";
import {
  executeWorkflowInvocation,
  workflowInvocationApiVersion,
  type WorkflowInvocation,
} from "./runtime.js";

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

const nodeCommand = (source: string): ScriptCommand => ({
  executable: process.execPath,
  arguments: ["-e", source],
  environment: {},
});

const registryWithScript = (
  command: ScriptCommand,
): WorkflowCommandRegistry => ({
  executors: {
    review: {adapter: "script", command, capabilities: []},
  },
  controls: {},
  evidenceSinks: {},
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
    onTestFinished(async () => {
      await rm(evidencePath, {force: true});
    });
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
      createCommandPluginRegistry(registry),
    );

    expect(result.execution?.outcome).toBe("advise");
    expect(result.evidenceReferences).toHaveLength(4);
    const evidence = await readFile(evidencePath, "utf8");
    expect(evidence.trim().split("\n").length).toBe(3);
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
      createCommandPluginRegistry(registry),
    );

    expect(result.outcome).toBe("executed");
    expect(result.execution?.references).toEqual(["artifact:agent-1"]);
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
      createCommandPluginRegistry(registry),
    );

    expect(result.outcome).toBe("denied");
    expect(result.execution).toBeUndefined();
  });

  it("does not initialize an unselected evidence sink", async () => {
    const evidencePath = join(
      tmpdir(),
      `xonovex-unused-evidence-${String(process.pid)}-${String(Date.now())}.jsonl`,
    );
    await writeFile(evidencePath, "not-json\n", "utf8");
    onTestFinished(async () => {
      await rm(evidencePath, {force: true});
    });
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
      createCommandPluginRegistry(registry),
    );

    expect(result.outcome).toBe("executed");
  });

  it("reports a command's non-zero exit and stderr", async () => {
    const registry = registryWithScript(
      nodeCommand('process.stderr.write("broken");process.exit(2)'),
    );

    const execution = executeWorkflowInvocation(
      baseInvocation(),
      createCommandPluginRegistry(registry),
    );

    await expect(execution).rejects.toThrow("workflow-command-failed:2:broken");
  });

  it("rejects command output that is not JSON", async () => {
    const registry = registryWithScript(
      nodeCommand('process.stdout.write("not-json")'),
    );

    const execution = executeWorkflowInvocation(
      baseInvocation(),
      createCommandPluginRegistry(registry),
    );

    await expect(execution).rejects.toThrow("workflow-command-invalid-json");
  });

  it("stops a command whose stdout exceeds its configured limit", async () => {
    const command = nodeCommand('process.stdout.write("12345")');
    command.maximumOutputBytes = 4;
    const registry = registryWithScript(command);

    const execution = executeWorkflowInvocation(
      baseInvocation(),
      createCommandPluginRegistry(registry),
    );

    await expect(execution).rejects.toThrow(
      "workflow-command-output-limit-exceeded",
    );
  });
});
