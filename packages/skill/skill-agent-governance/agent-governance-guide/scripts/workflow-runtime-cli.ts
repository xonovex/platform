#!/usr/bin/env node
import {readFile} from "node:fs/promises";
import {pathToFileURL} from "node:url";
import {createCommandPluginRegistry} from "./workflow-command-runtime.ts";
import {
  executeWorkflowInvocation,
  explainWorkflowComposition,
} from "./workflow-runtime.ts";
import {adaptTriggerToWorkflowInvocation} from "./workflow-trigger-adapters.ts";

const readStandardInput = async (maximumBytes: number): Promise<string> => {
  const chunks: Buffer[] = [];
  let bytes = 0;
  for await (const chunk of process.stdin) {
    const buffer = Buffer.isBuffer(chunk) ? chunk : Buffer.from(String(chunk));
    bytes += buffer.length;
    if (bytes > maximumBytes) {
      throw new Error("workflow-input-limit-exceeded");
    }
    chunks.push(buffer);
  }
  return Buffer.concat(chunks).toString("utf8");
};

const usage =
  "usage: workflow-runtime-cli.ts <run|explain> <registry.json> | trigger <registry.json> <template.json>";

const main = async (): Promise<void> => {
  const command = process.argv[2];
  const registryPath = process.argv[3];
  if (registryPath === undefined) throw new Error(usage);

  const registry = await createCommandPluginRegistry(
    JSON.parse(await readFile(registryPath, "utf8")) as unknown,
  );
  const input = JSON.parse(await readStandardInput(1_048_576)) as unknown;

  if (command === "run") {
    const result = await executeWorkflowInvocation(input, registry);
    process.stdout.write(`${JSON.stringify(result)}\n`);
    return;
  }
  if (command === "explain") {
    const result = explainWorkflowComposition(input, registry);
    process.stdout.write(`${JSON.stringify(result)}\n`);
    return;
  }
  if (command === "trigger") {
    const templatePath = process.argv[4];
    if (templatePath === undefined) throw new Error(usage);
    const invocation = adaptTriggerToWorkflowInvocation(
      input,
      JSON.parse(await readFile(templatePath, "utf8")) as unknown,
    );
    const result = await executeWorkflowInvocation(invocation, registry);
    process.stdout.write(`${JSON.stringify(result)}\n`);
    return;
  }
  throw new Error(usage);
};

if (import.meta.url === pathToFileURL(process.argv[1] ?? "").href) {
  main().catch((error: unknown) => {
    const message = error instanceof Error ? error.message : String(error);
    process.stderr.write(`${message}\n`);
    process.exitCode = 1;
  });
}
