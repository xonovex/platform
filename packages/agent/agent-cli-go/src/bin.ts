#!/usr/bin/env node
import {spawn} from "node:child_process";
import {existsSync} from "node:fs";
import {createRequire} from "node:module";
import {arch, platform} from "node:os";
import {dirname, join} from "node:path";
import {childExitCode, getBinaryName, getPlatformPackage} from "./launcher.js";

const require = createRequire(import.meta.url);

const findBinary = (): string => {
  const packageName = getPlatformPackage(platform(), arch());
  if (!packageName) {
    console.error(`No binary available for ${platform()}-${arch()}`);
    process.exit(1);
  }

  const binaryName = getBinaryName(platform());

  try {
    const packageJsonPath = require.resolve(`${packageName}/package.json`);
    const binaryPath = join(dirname(packageJsonPath), "bin", binaryName);

    if (existsSync(binaryPath)) {
      return binaryPath;
    }
  } catch {
    // Package not found
  }

  console.error(
    `Binary not found. Install with: npm install -g @xonovex/agent-cli-go`,
  );
  process.exit(1);
};

const main = (): void => {
  const binaryPath = findBinary();
  const child = spawn(binaryPath, process.argv.slice(2), {
    stdio: "inherit",
    env: process.env,
  });

  child.on("error", (err) => {
    console.error(`Failed to start: ${err.message}`);
    process.exit(1);
  });

  child.on("close", (code, signal) => {
    process.exit(childExitCode(code, signal));
  });
};

main();
