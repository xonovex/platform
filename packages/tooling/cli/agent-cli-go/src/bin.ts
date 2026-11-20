#!/usr/bin/env node
import {spawn} from "node:child_process";
import {existsSync} from "node:fs";
import {createRequire} from "node:module";
import {arch, constants, platform} from "node:os";
import {dirname, join} from "node:path";
import {launchAgentCli} from "./launcher.js";

const require = createRequire(import.meta.url);

process.exitCode = await launchAgentCli({
  currentPlatform: platform(),
  architecture: arch(),
  arguments: process.argv.slice(2),
  environment: process.env,
  signalNumber: (signal) => constants.signals[signal],
  locateBinary: (packageName, binaryName) => {
    try {
      const binaryPath = join(
        dirname(require.resolve(`${packageName}/package.json`)),
        "bin",
        binaryName,
      );
      return existsSync(binaryPath) ? binaryPath : undefined;
    } catch {
      return;
    }
  },
  runBinary: (binaryPath, arguments_, environment) =>
    new Promise((resolve) => {
      const child = spawn(binaryPath, arguments_, {
        stdio: "inherit",
        env: environment,
      });
      child.on("error", (error) => {
        resolve({kind: "failed-to-start", error});
      });
      child.on("close", (code, signal) => {
        resolve({kind: "closed", code, signal});
      });
    }),
  reportError: (message) => {
    console.error(message);
  },
});
