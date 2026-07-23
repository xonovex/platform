#!/usr/bin/env node
import {readFileSync} from "node:fs";
import {
  parseCompositionCatalog,
  parseCompositionRequest,
  parseInstalledSkillInventory,
  resolveComposition,
} from "./skill-composition.js";

const readJson = (
  path: string,
): {readonly input: unknown; readonly text: string} => {
  const text = readFileSync(path, "utf8");
  return {input: JSON.parse(text) as unknown, text};
};

const fail = (message: string): never => {
  throw new Error(message);
};

const main = (): void => {
  const [catalogPath, inventoryPath, requestPath, ...unexpected] =
    process.argv.slice(2);
  if (
    catalogPath === undefined ||
    inventoryPath === undefined ||
    requestPath === undefined ||
    unexpected.length > 0
  ) {
    return fail(
      "usage: xonovex-skill-compose <catalog.json> <installed-skills.json> <request.json>",
    );
  }
  const catalogInput = readJson(catalogPath);
  const catalog = parseCompositionCatalog(
    catalogInput.input,
    catalogInput.text,
  );
  if (!catalog.success) return fail(catalog.errors.join("\n"));
  const inventoryInput = readJson(inventoryPath);
  const inventory = parseInstalledSkillInventory(inventoryInput.input);
  if (!inventory.success) return fail(inventory.errors.join("\n"));
  const requestInput = readJson(requestPath);
  const request = parseCompositionRequest(requestInput.input);
  if (!request.success) return fail(request.errors.join("\n"));

  process.stdout.write(
    `${JSON.stringify(
      resolveComposition(catalog.data, inventory.data, request.data),
      undefined,
      2,
    )}\n`,
  );
};

try {
  main();
} catch (error) {
  process.stderr.write(
    `${error instanceof Error ? error.message : String(error)}\n`,
  );
  process.exitCode = 1;
}
