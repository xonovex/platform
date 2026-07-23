#!/usr/bin/env node
import {existsSync, readFileSync} from "node:fs";
import {resolve} from "node:path";
import {fileURLToPath} from "node:url";
import {parseArgs} from "node:util";
import {composeWorkflowRequest} from "./workflow-composition-runtime.js";

const HELP = `usage: xonovex-workflow-compose --request <workflow-request.json> [options]

Resolve a normalized WorkflowRequest against an installed skill snapshot.

options:
  --catalog <path>          composition catalog (bundle defaults to its packaged snapshot)
  --installed-root <path>   root containing installed skill plugins; repeatable
  --inventory <path>        explicit installed-skill inventory JSON
  --request <path>          normalized WorkflowRequest JSON
  -h, --help                show this help`;

const readJson = (
  path: string,
): {readonly input: unknown; readonly sourceText: string} => {
  const sourceText = readFileSync(path, "utf8");
  return {input: JSON.parse(sourceText) as unknown, sourceText};
};

const main = (argv: readonly string[]): number => {
  if (argv.includes("-h") || argv.includes("--help")) {
    process.stdout.write(`${HELP}\n`);
    return 0;
  }
  let parsed: ReturnType<typeof parseArgs>;
  try {
    parsed = parseArgs({
      args: [...argv],
      allowPositionals: false,
      strict: true,
      options: {
        catalog: {type: "string"},
        "installed-root": {type: "string", multiple: true},
        inventory: {type: "string"},
        request: {type: "string"},
      },
    });
  } catch (error) {
    process.stderr.write(
      `workflow composition input error: ${error instanceof Error ? error.message : String(error)}\n`,
    );
    return 2;
  }
  const requestPath = parsed.values.request;
  if (typeof requestPath !== "string") {
    process.stderr.write(
      "workflow composition input error: --request is required\n",
    );
    return 2;
  }
  const catalogPath =
    typeof parsed.values.catalog === "string"
      ? resolve(parsed.values.catalog)
      : fileURLToPath(
          new URL("../assets/composition-catalog.json", import.meta.url),
        );
  if (!existsSync(catalogPath)) {
    process.stderr.write(
      "workflow composition input error: --catalog is required outside the packaged workflow bundle\n",
    );
    return 2;
  }
  try {
    const catalog = readJson(catalogPath);
    const request = readJson(resolve(requestPath));
    const inventoryPath = parsed.values.inventory;
    const inventory =
      typeof inventoryPath === "string"
        ? readJson(resolve(inventoryPath)).input
        : undefined;
    const rawInstalledRoots = parsed.values["installed-root"];
    let installedRoots: readonly string[];
    if (Array.isArray(rawInstalledRoots)) {
      installedRoots = rawInstalledRoots.flatMap((root) =>
        typeof root === "string" ? [resolve(root)] : [],
      );
    } else {
      installedRoots = inventory === undefined ? [process.cwd()] : [];
    }
    const result = composeWorkflowRequest({
      catalogInput: catalog.input,
      catalogSourceText: catalog.sourceText,
      installedRoots,
      ...(inventory === undefined ? {} : {installedSkills: inventory}),
      workflowRequest: request.input,
    });
    if (!result.success) {
      process.stderr.write(`${result.errors.join("\n")}\n`);
      return 2;
    }
    process.stdout.write(`${JSON.stringify(result.data, undefined, 2)}\n`);
    return result.data.status === "blocked" ? 1 : 0;
  } catch (error) {
    process.stderr.write(
      `workflow composition failed: ${error instanceof Error ? error.message : String(error)}\n`,
    );
    return 2;
  }
};

process.exitCode = main(process.argv.slice(2));
