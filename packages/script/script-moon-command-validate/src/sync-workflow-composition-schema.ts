#!/usr/bin/env node
import {readFileSync, writeFileSync} from "node:fs";
import {resolve} from "node:path";
import {workflowRequestCompositionSchemaDefinitions} from "@xonovex/core/workflow-composition";
import {format, resolveConfig} from "prettier";

const path = resolve(process.argv[2] ?? "");
if (process.argv.length !== 3) {
  throw new Error(
    "usage: moon-command-workflow-schema-sync <workflow-request.schema.json>",
  );
}
const input = JSON.parse(readFileSync(path, "utf8")) as unknown;
if (
  typeof input !== "object" ||
  input === null ||
  Array.isArray(input) ||
  !("$defs" in input) ||
  typeof input.$defs !== "object" ||
  input.$defs === null ||
  Array.isArray(input.$defs)
) {
  throw new TypeError("workflow request schema needs an object $defs");
}
const definitions = input.$defs as Record<string, unknown>;
Object.assign(definitions, workflowRequestCompositionSchemaDefinitions());
const prettierConfig = await resolveConfig(path);
writeFileSync(
  path,
  await format(JSON.stringify(input), {
    ...prettierConfig,
    filepath: path,
  }),
);
