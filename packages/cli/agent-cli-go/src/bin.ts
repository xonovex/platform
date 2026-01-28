#!/usr/bin/env node
import {spawn} from "node:child_process";
import {existsSync} from "node:fs";
import {createRequire} from "node:module";
import {arch, platform} from "node:os";
import {dirname, join} from "node:path";

const require = createRequire(import.meta.url);

type Platform = "darwin" | "linux" | "win32";
type Arch = "arm64" | "x64";

const getPlatformPackage = (): string | undefined => {
  const os = platform() as Platform;
  const cpu = arch() as Arch;

  const map: Record<string, string> = {
    "darwin-arm64": "@xonovex/agent-cli-go-darwin-arm64",
    "darwin-x64": "@xonovex/agent-cli-go-darwin-x64",
    "linux-arm64": "@xonovex/agent-cli-go-linux-arm64",
    "linux-x64": "@xonovex/agent-cli-go-linux-x64",
    "win32-x64": "@xonovex/agent-cli-go-win32-x64",
  };

  return map[`${os}-${cpu}`];
};

const findBinary = (): string => {
  const packageName = getPlatformPackage();
  if (!packageName) {
    console.error(`No binary available for ${platform()}-${arch()}`);
    process.exit(1);
  }

  const binaryName =
    platform() === "win32" ? "agent-cli-go.exe" : "agent-cli-go";

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

  child.on("close", (code) => {
    process.exit(code ?? 0);
  });
};

main();
