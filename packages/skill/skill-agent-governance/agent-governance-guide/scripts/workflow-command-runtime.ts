import {spawn} from "node:child_process";
import {createHash} from "node:crypto";
import {appendFile, mkdir, readFile} from "node:fs/promises";
import {dirname} from "node:path";
import {pathToFileURL} from "node:url";
import {z} from "zod";
import {
  type WorkflowControl,
  type WorkflowEvidenceEvent,
  type WorkflowEvidenceSink,
  type WorkflowExecutor,
  type WorkflowInvocation,
  type WorkflowPluginRegistry,
} from "./workflow-runtime.ts";

const commandDefinitionSchema = z
  .object({
    executable: z.string().min(1),
    arguments: z.array(z.string()).default([]),
    workingDirectory: z.string().min(1).optional(),
    environment: z.record(z.string(), z.string()).default({}),
    timeoutSeconds: z.number().positive().optional(),
    maximumOutputBytes: z.number().int().positive().optional(),
  })
  .strict();

const capabilitiesSchema = z.array(z.string().min(1)).default([]);

const scriptExecutorSchema = z
  .object({
    adapter: z.literal("script"),
    command: commandDefinitionSchema,
    capabilities: capabilitiesSchema,
  })
  .strict();

const scriptLlmExecutorSchema = z
  .object({
    adapter: z.literal("script-llm"),
    script: commandDefinitionSchema,
    model: commandDefinitionSchema,
    capabilities: capabilitiesSchema,
  })
  .strict();

const agentSkillExecutorSchema = z
  .object({
    adapter: z.literal("agent-skill"),
    agent: commandDefinitionSchema,
    skill: z.string().min(1),
    capabilities: capabilitiesSchema,
  })
  .strict();

const executorDefinitionSchema = z.discriminatedUnion("adapter", [
  scriptExecutorSchema,
  scriptLlmExecutorSchema,
  agentSkillExecutorSchema,
]);

const controlDefinitionSchema = z
  .object({
    command: commandDefinitionSchema,
    phases: z.array(z.enum(["before", "after"])).min(1),
    capabilities: capabilitiesSchema,
  })
  .strict();

const jsonlEvidenceDefinitionSchema = z
  .object({
    adapter: z.literal("jsonl"),
    path: z.string().min(1),
    baseReference: z.string().min(1).optional(),
    capabilities: capabilitiesSchema,
  })
  .strict();

export const workflowCommandRegistrySchema = z
  .object({
    executors: z.record(z.string(), executorDefinitionSchema),
    controls: z.record(z.string(), controlDefinitionSchema).default({}),
    evidenceSinks: z
      .record(z.string(), jsonlEvidenceDefinitionSchema)
      .default({}),
  })
  .strict();

export type WorkflowCommandRegistry = z.infer<
  typeof workflowCommandRegistrySchema
>;

type CommandDefinition = z.infer<typeof commandDefinitionSchema>;

interface CommandRequest {
  readonly kind:
    | "script.execute"
    | "script.collect"
    | "model.evaluate"
    | "agent.execute"
    | "control.evaluate";
  readonly invocation: WorkflowInvocation;
  readonly input: unknown;
  readonly facts?: unknown;
  readonly skill?: string;
  readonly phase?: "before" | "after";
  readonly execution?: unknown;
}

const terminateProcess = (
  child: ReturnType<typeof spawn>,
  terminationGraceMilliseconds: number,
): void => {
  if (child.exitCode !== null || child.signalCode !== null) return;
  const terminate = (signal: NodeJS.Signals): void => {
    if (process.platform !== "win32" && child.pid !== undefined) {
      try {
        process.kill(-child.pid, signal);
        return;
      } catch {
        // The process may have exited between the state check and the signal.
      }
    }
    child.kill(signal);
  };
  terminate("SIGTERM");
  const force = setTimeout(
    () => terminate("SIGKILL"),
    terminationGraceMilliseconds,
  );
  force.unref();
};

const runProcess = async (
  definition: CommandDefinition,
  request: CommandRequest,
  signal: AbortSignal,
): Promise<unknown> => {
  if (signal.aborted) throw signal.reason;
  const maximumOutputBytes = definition.maximumOutputBytes ?? 1_048_576;
  const terminationGraceMilliseconds = 1_000;

  return await new Promise((resolve, reject) => {
    const child = spawn(definition.executable, definition.arguments, {
      cwd: definition.workingDirectory,
      env: {PATH: process.env.PATH ?? "", ...definition.environment},
      shell: false,
      detached: process.platform !== "win32",
      stdio: ["pipe", "pipe", "pipe"],
      windowsHide: true,
    });
    const stdout: Buffer[] = [];
    const stderr: Buffer[] = [];
    let stdoutBytes = 0;
    let stderrBytes = 0;
    let settled = false;

    const finish = (error?: Error, value?: unknown): void => {
      if (settled) return;
      settled = true;
      signal.removeEventListener("abort", abort);
      if (error === undefined) resolve(value);
      else reject(error);
    };
    const abort = (): void => {
      terminateProcess(child, terminationGraceMilliseconds);
      finish(
        signal.reason instanceof Error
          ? signal.reason
          : new Error("workflow-command-aborted"),
      );
    };
    signal.addEventListener("abort", abort, {once: true});

    child.once("error", (error) =>
      finish(new Error(`workflow-command-start-failed:${error.message}`)),
    );
    child.stdout.on("data", (chunk: Buffer) => {
      stdoutBytes += chunk.length;
      if (stdoutBytes > maximumOutputBytes) {
        terminateProcess(child, terminationGraceMilliseconds);
        finish(new Error("workflow-command-output-limit-exceeded"));
        return;
      }
      stdout.push(chunk);
    });
    child.stderr.on("data", (chunk: Buffer) => {
      stderrBytes += chunk.length;
      if (stderrBytes <= maximumOutputBytes) stderr.push(chunk);
    });
    child.once("close", (code, processSignal) => {
      if (settled) return;
      const errorOutput = Buffer.concat(stderr).toString("utf8").trim();
      if (code !== 0) {
        finish(
          new Error(
            `workflow-command-failed:${String(code ?? processSignal)}${errorOutput === "" ? "" : `:${errorOutput}`}`,
          ),
        );
        return;
      }
      try {
        finish(undefined, JSON.parse(Buffer.concat(stdout).toString("utf8")));
      } catch (error) {
        finish(
          new Error(
            `workflow-command-invalid-json:${error instanceof Error ? error.message : String(error)}`,
          ),
        );
      }
    });
    child.stdin.once("error", (error) =>
      finish(new Error(`workflow-command-input-failed:${error.message}`)),
    );
    child.stdin.end(`${JSON.stringify(request)}\n`);
  });
};

const executeCommand = async (
  definition: CommandDefinition,
  request: CommandRequest,
  parentSignal: AbortSignal,
): Promise<unknown> => {
  const controller = new AbortController();
  const abort = (): void => controller.abort(parentSignal.reason);
  parentSignal.addEventListener("abort", abort, {once: true});
  const timeout =
    definition.timeoutSeconds === undefined
      ? undefined
      : setTimeout(
          () => controller.abort(new Error("workflow-command-timeout")),
          definition.timeoutSeconds * 1_000,
        );
  try {
    return await runProcess(definition, request, controller.signal);
  } finally {
    parentSignal.removeEventListener("abort", abort);
    if (timeout !== undefined) clearTimeout(timeout);
  }
};

const createExecutor = (
  id: string,
  definition: z.infer<typeof executorDefinitionSchema>,
): WorkflowExecutor => {
  if (definition.adapter === "script") {
    return {
      id,
      capabilities: definition.capabilities,
      execute: async (invocation, input, signal) =>
        await executeCommand(
          definition.command,
          {kind: "script.execute", invocation, input},
          signal,
        ),
    };
  }
  if (definition.adapter === "script-llm") {
    return {
      id,
      capabilities: definition.capabilities,
      execute: async (invocation, input, signal) => {
        const facts = await executeCommand(
          definition.script,
          {kind: "script.collect", invocation, input},
          signal,
        );
        return await executeCommand(
          definition.model,
          {kind: "model.evaluate", invocation, input, facts},
          signal,
        );
      },
    };
  }
  return {
    id,
    capabilities: definition.capabilities,
    execute: async (invocation, input, signal) =>
      await executeCommand(
        definition.agent,
        {
          kind: "agent.execute",
          invocation,
          input,
          skill: definition.skill,
        },
        signal,
      ),
  };
};

const createControl = (
  id: string,
  definition: z.infer<typeof controlDefinitionSchema>,
): WorkflowControl => ({
  id,
  capabilities: definition.capabilities,
  phases: definition.phases,
  evaluate: async ({phase, invocation, execution, input}, signal) =>
    await executeCommand(
      definition.command,
      {
        kind: "control.evaluate",
        invocation,
        input,
        phase,
        execution,
      },
      signal,
    ),
});

const eventDigest = (event: WorkflowEvidenceEvent): string =>
  createHash("sha256").update(JSON.stringify(event)).digest("hex");

const createJsonlEvidenceSink = async (
  id: string,
  definition: z.infer<typeof jsonlEvidenceDefinitionSchema>,
): Promise<WorkflowEvidenceSink> => {
  const baseReference =
    definition.baseReference ?? pathToFileURL(definition.path).href;
  const recorded = new Set<string>();
  let initialized = false;
  const initialize = async (): Promise<void> => {
    if (initialized) return;
    try {
      const content = await readFile(definition.path, "utf8");
      for (const line of content.split("\n").filter((value) => value !== "")) {
        recorded.add(eventDigest(JSON.parse(line) as WorkflowEvidenceEvent));
      }
    } catch (error) {
      if (
        error === null ||
        typeof error !== "object" ||
        Reflect.get(error, "code") !== "ENOENT"
      ) {
        throw error;
      }
    }
    initialized = true;
  };

  let pending = Promise.resolve();
  return {
    id,
    capabilities: definition.capabilities,
    record: (event) => {
      const digest = eventDigest(event);
      const reference = `${baseReference}#sha256:${digest}`;
      const write = pending.then(async () => {
        await initialize();
        if (!recorded.has(digest)) {
          await mkdir(dirname(definition.path), {recursive: true});
          await appendFile(
            definition.path,
            `${JSON.stringify(event)}\n`,
            "utf8",
          );
          recorded.add(digest);
        }
        return reference;
      });
      pending = write.then(
        () => undefined,
        () => undefined,
      );
      return write;
    },
  };
};

export const createCommandPluginRegistry = async (
  untrustedRegistry: unknown,
): Promise<WorkflowPluginRegistry> => {
  const registry = workflowCommandRegistrySchema.parse(untrustedRegistry);
  const executors = Object.fromEntries(
    Object.entries(registry.executors).map(([id, definition]) => [
      id,
      createExecutor(id, definition),
    ]),
  );
  const controls = Object.fromEntries(
    Object.entries(registry.controls).map(([id, definition]) => [
      id,
      createControl(id, definition),
    ]),
  );
  const evidenceSinks = Object.fromEntries(
    await Promise.all(
      Object.entries(registry.evidenceSinks).map(
        async ([id, definition]) =>
          [id, await createJsonlEvidenceSink(id, definition)] as const,
      ),
    ),
  );
  return {executors, controls, evidenceSinks};
};
