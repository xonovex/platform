#!/usr/bin/env node
import {main} from "./audit.js";

try {
  process.exitCode = await main(process.argv.slice(2));
} catch (error: unknown) {
  const message = error instanceof Error ? error.message : String(error);
  process.stderr.write(`error: ${message}\n`);
  process.exitCode = 2;
}
