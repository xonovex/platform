#!/usr/bin/env node
import {resolve} from "node:path";
import {syncCompositionCatalog} from "./sync-composition-catalog.js";

const main = (args: readonly string[]): number => {
  const repoRoot = resolve(args[0] ?? ".");
  try {
    const result = syncCompositionCatalog(repoRoot);
    console.log(
      result.changed
        ? `Updated ${result.snapshotPath} from ${result.sourcePath}.`
        : `Composition catalog snapshot is current at ${result.snapshotPath}.`,
    );
    return 0;
  } catch (error) {
    console.error(
      `Could not synchronize the composition catalog: ${error instanceof Error ? error.message : String(error)}`,
    );
    return 1;
  }
};

process.exitCode = main(process.argv.slice(2));
