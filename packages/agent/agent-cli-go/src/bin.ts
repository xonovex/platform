#!/usr/bin/env node
import {spawn} from "node:child_process";
import {existsSync} from "node:fs";
import {createRequire} from "node:module";
import {arch, platform} from "node:os";
import {launchAgentCli} from "./launcher.js";

const require = createRequire(import.meta.url);

launchAgentCli({
  currentPlatform: platform(),
  architecture: arch(),
  arguments: process.argv.slice(2),
  environment: process.env,
  resolvePackageJson: (packageName) =>
    require.resolve(`${packageName}/package.json`),
  binaryExists: existsSync,
  startBinary: (binaryPath, arguments_, environment, onError, onClose) => {
    const child = spawn(binaryPath, arguments_, {
      stdio: "inherit",
      env: environment,
    });
    child.on("error", onError);
    child.on("close", onClose);
  },
  reportError: (message) => {
    console.error(message);
  },
  exit: (code) => process.exit(code),
});
