const PLATFORM_PACKAGES: Readonly<Record<string, string>> = {
  "darwin-arm64": "@xonovex/agent-cli-go-darwin-arm64",
  "darwin-x64": "@xonovex/agent-cli-go-darwin-x64",
  "linux-arm64": "@xonovex/agent-cli-go-linux-arm64",
  "linux-x64": "@xonovex/agent-cli-go-linux-x64",
  "win32-x64": "@xonovex/agent-cli-go-win32-x64",
};

const NOT_INSTALLED =
  "Binary not found. Install with: npm install -g @xonovex/agent-cli-go";

export const getPlatformPackage = (
  platform: NodeJS.Platform,
  architecture: string,
): string | undefined => PLATFORM_PACKAGES[`${platform}-${architecture}`];

export const getBinaryName = (platform: NodeJS.Platform): string =>
  platform === "win32" ? "agent-cli-go.exe" : "agent-cli-go";

/**
 * Finds an installed binary. The adapter owns the package layout: it returns
 * the path of the named binary inside the named package, or undefined when
 * either is missing.
 */
interface BinaryLookup {
  readonly locateBinary: (
    packageName: string,
    binaryName: string,
  ) => string | undefined;
}

type BinaryResolution =
  | {readonly ok: true; readonly binaryPath: string}
  | {readonly ok: false; readonly error: string};

/** How a child process ended: it failed to start, or it closed. */
type ChildOutcome =
  | {readonly kind: "failed-to-start"; readonly error: Error}
  | {
      readonly kind: "closed";
      readonly code: number | null;
      readonly signal: NodeJS.Signals | null;
    };

interface LauncherRuntime extends BinaryLookup {
  readonly currentPlatform: NodeJS.Platform;
  readonly architecture: string;
  readonly arguments: readonly string[];
  readonly environment: NodeJS.ProcessEnv;
  readonly signalNumber: (signal: NodeJS.Signals) => number;
  readonly runBinary: (
    binaryPath: string,
    arguments_: readonly string[],
    environment: NodeJS.ProcessEnv,
  ) => Promise<ChildOutcome>;
  readonly reportError: (message: string) => void;
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

  const binaryPath = lookup.locateBinary(
    packageName,
    getBinaryName(currentPlatform),
  );
  return binaryPath === undefined
    ? {ok: false, error: NOT_INSTALLED}
    : {ok: true, binaryPath};
};

export const childExitCode = (
  code: number | null,
  signal: NodeJS.Signals | null,
  signalNumber: (signal: NodeJS.Signals) => number,
): number => {
  if (code !== null) return code;
  if (signal === null) return 1;
  return 128 + signalNumber(signal);
};

/**
 * Runs the platform binary and returns the exit code the launcher process
 * must end with.
 */
export const launchAgentCli = async (
  runtime: LauncherRuntime,
): Promise<number> => {
  const resolution = findBinary(
    runtime.currentPlatform,
    runtime.architecture,
    runtime,
  );
  if (!resolution.ok) {
    runtime.reportError(resolution.error);
    return 1;
  }

  const outcome = await runtime.runBinary(
    resolution.binaryPath,
    runtime.arguments,
    runtime.environment,
  );
  if (outcome.kind === "failed-to-start") {
    runtime.reportError(`Failed to start: ${outcome.error.message}`);
    return 1;
  }
  return childExitCode(outcome.code, outcome.signal, runtime.signalNumber);
};
