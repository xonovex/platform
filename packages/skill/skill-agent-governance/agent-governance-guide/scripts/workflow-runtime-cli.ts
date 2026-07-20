#!/usr/bin/env node
import {readFile} from "node:fs/promises";
import {pathToFileURL} from "node:url";
import {
  createCommandRuntimePorts,
  createJsonlWorkflowEvidenceRecorder,
} from "./workflow-command-runtime.ts";
import {executeWorkflowInvocation} from "./workflow-runtime.ts";

const readStandardInput = async (maximumBytes: number): Promise<string> => {
  const chunks: Buffer[] = [];
  let bytes = 0;
  for await (const chunk of process.stdin) {
    const buffer = Buffer.isBuffer(chunk) ? chunk : Buffer.from(String(chunk));
    bytes += buffer.length;
    if (bytes > maximumBytes) {
      throw new Error("workflow-invocation-input-limit-exceeded");
    }
    chunks.push(buffer);
  }
  return Buffer.concat(chunks).toString("utf8");
};

const main = async (): Promise<void> => {
  const registryPath = process.argv[2];
  const evidencePath = process.argv[3];
  if (registryPath === undefined || evidencePath === undefined) {
    throw new Error(
      "usage: workflow-runtime-cli.ts <trusted-command-registry.json> <evidence.jsonl>",
    );
  }

  const registry = JSON.parse(await readFile(registryPath, "utf8")) as unknown;
  const invocation = JSON.parse(await readStandardInput(1_048_576)) as unknown;
  const recorder = await createJsonlWorkflowEvidenceRecorder(evidencePath);
  const result = await executeWorkflowInvocation(
    invocation,
    createCommandRuntimePorts(registry, recorder),
  );
  process.stdout.write(`${JSON.stringify(result)}\n`);
};

if (import.meta.url === pathToFileURL(process.argv[1] ?? "").href) {
  main().catch((error: unknown) => {
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(`${message}\n`);
    process.exitCode = 1;
  });
}
