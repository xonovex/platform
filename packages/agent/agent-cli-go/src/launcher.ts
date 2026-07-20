import {constants} from "node:os";
import {dirname, join} from "node:path";

const PLATFORM_PACKAGES: Readonly<Record<string, string>> = {
  "darwin-arm64": "@xonovex/agent-cli-go-darwin-arm64",
  "darwin-x64": "@xonovex/agent-cli-go-darwin-x64",
  "linux-arm64": "@xonovex/agent-cli-go-linux-arm64",
  "linux-x64": "@xonovex/agent-cli-go-linux-x64",
  "win32-x64": "@xonovex/agent-cli-go-win32-x64",
};

export const getPlatformPackage = (
  platform: NodeJS.Platform,
  architecture: string,
): string | undefined => PLATFORM_PACKAGES[`${platform}-${architecture}`];

export const getBinaryName = (platform: NodeJS.Platform): string =>
  platform === "win32" ? "agent-cli-go.exe" : "agent-cli-go";

interface BinaryLookup {
  readonly resolvePackageJson: (packageName: string) => string;
  readonly binaryExists: (binaryPath: string) => boolean;
}

type BinaryResolution =
  | {readonly ok: true; readonly binaryPath: string}
  | {readonly ok: false; readonly error: string};

interface LauncherRuntime extends BinaryLookup {
  readonly currentPlatform: NodeJS.Platform;
  readonly architecture: string;
  readonly arguments: readonly string[];
  readonly environment: NodeJS.ProcessEnv;
  readonly startBinary: (
    binaryPath: string,
    arguments_: readonly string[],
    environment: NodeJS.ProcessEnv,
    onError: (error: Error) => void,
    onClose: (code: number | null, signal: NodeJS.Signals | null) => void,
  ) => void;
  readonly reportError: (message: string) => void;
  readonly exit: (code: number) => void;
}

export const findBinary = (
  currentPlatform: NodeJS.Platform,
  architecture: string,
  lookup: BinaryLookup,
): BinaryResolution => {
  const packageName = getPlatformPackage(currentPlatform, architecture);
  if (packageName === undefined) {
    return {
      ok: false,
      error: `No binary available for ${currentPlatform}-${architecture}`,
    };
  }

  try {
    const packageJsonPath = lookup.resolvePackageJson(packageName);
    const binaryPath = join(
      dirname(packageJsonPath),
      "bin",
      getBinaryName(currentPlatform),
    );
    if (lookup.binaryExists(binaryPath)) return {ok: true, binaryPath};
  } catch {
    return {
      ok: false,
      error:
        "Binary not found. Install with: npm install -g @xonovex/agent-cli-go",
    };
  }

  return {
    ok: false,
    error:
      "Binary not found. Install with: npm install -g @xonovex/agent-cli-go",
  };
};

export const childExitCode = (
  code: number | null,
  signal: NodeJS.Signals | null,
): number => {
  if (code !== null) return code;
  if (signal === null) return 1;
  return 128 + constants.signals[signal];
};

export const launchAgentCli = (runtime: LauncherRuntime): void => {
  const resolution = findBinary(
    runtime.currentPlatform,
    runtime.architecture,
    runtime,
  );
  if (!resolution.ok) {
    runtime.reportError(resolution.error);
    runtime.exit(1);
    return;
  }

  runtime.startBinary(
    resolution.binaryPath,
    runtime.arguments,
    runtime.environment,
    (error) => {
      runtime.reportError(`Failed to start: ${error.message}`);
      runtime.exit(1);
    },
    (code, signal) => {
      runtime.exit(childExitCode(code, signal));
    },
  );
};
