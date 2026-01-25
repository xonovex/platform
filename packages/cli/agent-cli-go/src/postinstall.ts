import { createRequire } from "node:module";
import { arch, platform } from "node:os";
import { dirname, join } from "node:path";
import { existsSync, mkdirSync, symlinkSync, unlinkSync } from "node:fs";

const require = createRequire(import.meta.url);

type Platform = "darwin" | "linux" | "win32";
type Arch = "arm64" | "x64";

interface PlatformConfig {
  readonly packageName: string;
  readonly binaryName: string;
}

const getPlatformConfig = (): PlatformConfig | undefined => {
  const os = platform() as Platform;
  const cpu = arch() as Arch;

  const platformMap: Record<Platform, Record<Arch, string | undefined>> = {
    darwin: {
      arm64: "@xonovex/agent-cli-go-darwin-arm64",
      x64: "@xonovex/agent-cli-go-darwin-x64",
    },
    linux: {
      arm64: "@xonovex/agent-cli-go-linux-arm64",
      x64: "@xonovex/agent-cli-go-linux-x64",
    },
    win32: {
      arm64: undefined,
      x64: "@xonovex/agent-cli-go-win32-x64",
    },
  };

  const packageName = platformMap[os]?.[cpu];
  if (!packageName) {
    return undefined;
  }

  const binaryName = os === "win32" ? "agent-cli-go.exe" : "agent-cli-go";

  return { packageName, binaryName };
};

const resolveBinaryPath = (packageName: string, binaryName: string): string | undefined => {
  try {
    const packageJsonPath = require.resolve(`${packageName}/package.json`);
    const packageDir = dirname(packageJsonPath);
    const binaryPath = join(packageDir, "bin", binaryName);

    if (existsSync(binaryPath)) {
      return binaryPath;
    }

    return undefined;
  } catch {
    return undefined;
  }
};

const createSymlink = (sourcePath: string, targetPath: string): void => {
  const targetDir = dirname(targetPath);

  if (!existsSync(targetDir)) {
    mkdirSync(targetDir, { recursive: true });
  }

  if (existsSync(targetPath)) {
    unlinkSync(targetPath);
  }

  symlinkSync(sourcePath, targetPath);
};

const main = (): void => {
  const config = getPlatformConfig();

  if (!config) {
    console.warn(
      `[agent-cli-go] No prebuilt binary available for ${platform()}-${arch()}. ` +
        "You may need to build from source."
    );
    process.exitCode = 0;
    return;
  }

  const binaryPath = resolveBinaryPath(config.packageName, config.binaryName);

  if (!binaryPath) {
    console.warn(
      `[agent-cli-go] Platform package ${config.packageName} not found. ` +
        "This is expected if the optional dependency was not installed."
    );
    process.exitCode = 0;
    return;
  }

  const packageRoot = dirname(dirname(new URL(import.meta.url).pathname));
  const targetPath = join(packageRoot, "bin", config.binaryName);

  try {
    createSymlink(binaryPath, targetPath);
    console.log(`[agent-cli-go] Linked binary: ${targetPath} -> ${binaryPath}`);
  } catch (error) {
    console.error(`[agent-cli-go] Failed to create symlink: ${error}`);
    process.exitCode = 1;
  }
};

main();
