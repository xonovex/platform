import {constants} from "node:os";

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

export const childExitCode = (
  code: number | null,
  signal: NodeJS.Signals | null,
): number => {
  if (code !== null) return code;
  if (signal === null) return 1;
  return 128 + constants.signals[signal];
};
