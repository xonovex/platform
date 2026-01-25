import { existsSync, readFileSync } from "node:fs";
import { homedir } from "node:os";
import { join } from "node:path";
import { parse as parseYaml } from "yaml";
import type { SandboxConfig, SandboxMethod } from "./types.js";

/**
 * Default config file location
 */
const DEFAULT_CONFIG_PATH = join(
  homedir(),
  ".config",
  "sandboxed-claude",
  "config",
);

/**
 * Config file schema
 */
interface ConfigFile {
  method?: SandboxMethod;
  homeDir?: string;
  network?: boolean;
  bindPaths?: string[];
  roBindPaths?: string[];
  provider?: string;
}

/**
 * Load sandbox configuration from a file
 */
export function loadConfigFile(configPath?: string): Partial<ConfigFile> {
  const path = configPath ?? DEFAULT_CONFIG_PATH;

  if (!existsSync(path)) {
    return {};
  }

  const content = readFileSync(path, "utf8");

  // Try YAML first, then fall back to simple key=value format
  try {
    return parseYaml(content) as ConfigFile;
  } catch {
    // Parse as simple key=value format
    const config: Partial<ConfigFile> = {};
    for (const line of content.split("\n")) {
      const trimmed = line.trim();
      if (!trimmed || trimmed.startsWith("#")) continue;

      const [keyPart, ...valueParts] = trimmed.split("=");
      if (!keyPart) continue;
      const key = keyPart.trim();
      const value = valueParts.join("=").trim();

      switch (key) {
        case "method": {
          config.method = value as SandboxMethod;
          break;
        }
        case "homeDir":
        case "SANDBOXHOMEDIR": {
          config.homeDir = value;
          break;
        }
        case "network":
        case "ENABLE_NETWORK": {
          config.network = value === "true";
          break;
        }
        case "provider": {
          config.provider = value;
          break;
        }
      }
    }
    return config;
  }
}

/**
 * Merge config file with CLI options (CLI takes precedence)
 */
export function mergeConfig(
  fileConfig: Partial<ConfigFile>,
  cliConfig: Partial<SandboxConfig>,
): Partial<SandboxConfig> {
  return {
    method: cliConfig.method ?? fileConfig.method,
    homeDir: cliConfig.homeDir ?? fileConfig.homeDir,
    network: cliConfig.network ?? fileConfig.network,
    bindPaths: [
      ...(fileConfig.bindPaths ?? []),
      ...(cliConfig.bindPaths ?? []),
    ],
    roBindPaths: [
      ...(fileConfig.roBindPaths ?? []),
      ...(cliConfig.roBindPaths ?? []),
    ],
  };
}
