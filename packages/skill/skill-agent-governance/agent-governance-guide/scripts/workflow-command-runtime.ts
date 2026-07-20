import {spawn} from "node:child_process";
import {createHash} from "node:crypto";
import {appendFile, mkdir, readFile} from "node:fs/promises";
import {dirname} from "node:path";
import {pathToFileURL} from "node:url";
import {z} from "zod";
import {
  type WorkflowExecutionEvidence,
  type WorkflowInvocation,
  type WorkflowRuntimePorts,
} from "./workflow-runtime.ts";

const commandDefinitionSchema = z
  .object({
    executable: z.string().min(1),
    arguments: z.array(z.string()).default([]),
    workingDirectory: z.string().min(1).optional(),
    environment: z.record(z.string(), z.string()).default({}),
  })
  .strict();

export const workflowCommandRegistrySchema = z
  .object({
    scripts: z.record(z.string(), commandDefinitionSchema),
    models: z.record(z.string(), commandDefinitionSchema),
    agents: z.record(z.string(), commandDefinitionSchema),
    oversightVerifiers: z.record(z.string(), commandDefinitionSchema),
  })
  .strict();

export type WorkflowCommandRegistry = z.infer<
  typeof workflowCommandRegistrySchema
>;

interface CommandRequest {
  readonly kind:
    | "workflow-script"
    | "bounded-model"
    | "bounded-agent"
    | "oversight-verification";
  readonly invocation: WorkflowInvocation;
  readonly facts?: unknown;
  readonly workflowSkill?: string;
  readonly oversight?: unknown;
}

interface CommandRunnerOptions {
  readonly maximumOutputBytes?: number;
  readonly terminationGraceMilliseconds?: number;
}

const lookupCommand = (
  commands: Readonly<Record<string, z.infer<typeof commandDefinitionSchema>>>,
  identifier: string,
): z.infer<typeof commandDefinitionSchema> => {
  const command = commands[identifier];
  if (command === undefined) {
    throw new Error(`workflow-command-not-registered:${identifier}`);
  }
  return command;
};

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

const executeCommand = async (
  definition: z.infer<typeof commandDefinitionSchema>,
  request: CommandRequest,
  signal: AbortSignal,
  options: CommandRunnerOptions,
): Promise<unknown> => {
  if (signal.aborted) throw signal.reason;
  const maximumOutputBytes = options.maximumOutputBytes ?? 1_048_576;
  const terminationGraceMilliseconds =
    options.terminationGraceMilliseconds ?? 1_000;

  return await new Promise((resolve, reject) => {
    const child = spawn(definition.executable, definition.arguments, {
      cwd: definition.workingDirectory,
      env: {
        PATH: process.env.PATH ?? "",
        ...definition.environment,
      },
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
      finish(new Error(`workflow-command-start-failed: ${error.message}`)),
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
            `workflow-command-failed:${String(code ?? processSignal)}${errorOutput === "" ? "" : `: ${errorOutput}`}`,
          ),
        );
        return;
      }
      try {
        finish(undefined, JSON.parse(Buffer.concat(stdout).toString("utf8")));
      } catch (error) {
        finish(
          new Error(
            `workflow-command-invalid-json: ${error instanceof Error ? error.message : String(error)}`,
          ),
        );
      }
    });

    child.stdin.once("error", (error) =>
      finish(new Error(`workflow-command-input-failed: ${error.message}`)),
    );
    child.stdin.end(`${JSON.stringify(request)}\n`);
  });
};

export interface WorkflowEvidenceRecorder {
  readonly record: (evidence: WorkflowExecutionEvidence) => Promise<string>;
}

interface RecordedWorkflowEvidence {
  readonly digest: string;
  readonly reference: string;
}

const digestWorkflowEvidence = (evidence: WorkflowExecutionEvidence): string =>
  createHash("sha256").update(JSON.stringify(evidence)).digest("hex");

export const createJsonlWorkflowEvidenceRecorder = async (
  path: string,
  baseReference: string = pathToFileURL(path).href,
): Promise<WorkflowEvidenceRecorder> => {
  await mkdir(dirname(path), {recursive: true});
  const recorded = new Map<string, RecordedWorkflowEvidence>();
  try {
    const content = await readFile(path, "utf8");
    for (const line of content.split("\n").filter((value) => value !== "")) {
      const evidence = JSON.parse(line) as WorkflowExecutionEvidence;
      const digest = digestWorkflowEvidence(evidence);
      const existing = recorded.get(evidence.invocationId);
      if (existing !== undefined && existing.digest !== digest) {
        throw new Error("workflow-evidence-log-invocation-collision");
      }
      recorded.set(evidence.invocationId, {
        digest,
        reference: `${baseReference}#sha256:${digest}`,
      });
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

  let pending = Promise.resolve();
  const record = (evidence: WorkflowExecutionEvidence): Promise<string> => {
    const digest = digestWorkflowEvidence(evidence);
    const reference = `${baseReference}#sha256:${digest}`;
    const write = pending.then(async () => {
      const existing = recorded.get(evidence.invocationId);
      if (existing?.digest === digest) return existing.reference;
      if (existing !== undefined) {
        throw new Error("workflow-evidence-invocation-collision");
      }
      await appendFile(path, `${JSON.stringify(evidence)}\n`, "utf8");
      recorded.set(evidence.invocationId, {digest, reference});
      return reference;
    });
    pending = write.then(
      () => undefined,
      () => undefined,
    );
    return write;
  };
  return {record};
};

export const createCommandRuntimePorts = (
  untrustedRegistry: unknown,
  evidenceRecorder: WorkflowEvidenceRecorder,
  options: CommandRunnerOptions = {},
): WorkflowRuntimePorts => {
  const registry = workflowCommandRegistrySchema.parse(untrustedRegistry);
  return {
    runScript: async ({module, invocation}, signal) =>
      await executeCommand(
        lookupCommand(registry.scripts, module),
        {kind: "workflow-script", invocation},
        signal,
        options,
      ),
    runModel: async ({evaluator, invocation, facts}, signal) =>
      await executeCommand(
        lookupCommand(registry.models, evaluator),
        {kind: "bounded-model", invocation, facts},
        signal,
        options,
      ),
    runAgent: async ({launcher, workflowSkill, invocation}, signal) =>
      await executeCommand(
        lookupCommand(registry.agents, launcher),
        {kind: "bounded-agent", invocation, workflowSkill},
        signal,
        options,
      ),
    verifyOversight: async (invocation, signal) => {
      if (invocation.execution.family !== "agent-workflow-skill") {
        throw new Error("workflow-oversight-not-applicable");
      }
      return await executeCommand(
        lookupCommand(
          registry.oversightVerifiers,
          invocation.execution.oversight.level,
        ),
        {
          kind: "oversight-verification",
          invocation,
          oversight: invocation.execution.oversight,
        },
        signal,
        options,
      );
    },
    recordEvidence: evidenceRecorder.record,
  };
};
