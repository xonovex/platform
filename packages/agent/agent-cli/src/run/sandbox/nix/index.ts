import {spawnSync} from "node:child_process";
import {existsSync, mkdirSync, realpathSync} from "node:fs";
import {homedir} from "node:os";
import {join} from "node:path";
import {logDebug, logError, logInfo} from "@xonovex/core";
import {claudeAgent} from "../../agents/claude/index.js";
import type {AgentConfig} from "../../agents/types.js";
import {buildEnv} from "../../env/build.js";
import {getAgentsDir} from "../../env/resolve.js";
import {
  DEFAULT_BASE_PACKAGES,
  DEFAULT_NIXPKGS_PIN,
  type EnvSpec,
} from "../../env/types.js";
import {getAgentIdEnv} from "../../id.js";
import {buildProviderEnvironment} from "../../providers/types.js";
import {executeInTerminal} from "../../wrapper/index.js";
import {
  bindMountsToBwrapArgs,
  ensureSandboxMountPoint,
  getApplicationBindMounts,
} from "../bindings.js";
import {wrapWithInitCommands} from "../command.js";
import {
  envToBwrapArgs,
  getSandboxEnvironment,
  mergeCustomEnv,
} from "../environment.js";
import {buildAgentCommand, spawnSandboxProcess} from "../shared-utils.js";
import type {SandboxConfig, SandboxExecutor} from "../types.js";

/**
 * Extended sandbox config for Nix
 */
export interface NixSandboxConfig {
  /** Packages to include (added to defaults) */
  packages?: string[];
  /** Nixpkgs pin to use */
  nixpkgs_pin?: string;
  /** Use only specified packages (no defaults) */
  noDefaults?: boolean;
}

/**
 * Package sets - predefined collections of packages for common use cases
 */
const PACKAGE_SETS: Record<string, string[]> = {
  nodejs: [
    "nodejs_24",
    "python312",
    "gnumake",
    "gcc",
    "gnused",
    "gawk",
    "binutils",
  ],
  python: ["python312", "python312Packages.pip"],
  go: ["go"],
  rust: ["rustc", "cargo"],
  kubernetes: ["kubectl", "kubernetes-helm", "k9s"],
  terraform: ["terraform", "terragrunt"],
  docker: ["docker-client"],
  aws: ["awscli2"],
  gcp: ["google-cloud-sdk"],
};

/**
 * Expand set names to package lists
 */
function expandPackageSets(sets: string[]): string[] {
  const packages: string[] = [];
  for (const set of sets) {
    const setPackages = PACKAGE_SETS[set.trim()];
    if (setPackages) {
      packages.push(...setPackages);
    } else {
      logError(`Unknown package set "${set}", skipping`);
    }
  }
  return packages;
}

/**
 * Check if nix-build is available
 */
function checkNixBuildAvailable(): boolean {
  // eslint-disable-next-line sonarjs/no-os-command-from-path
  const result = spawnSync("which", ["nix-build"], {stdio: "pipe"});
  return result.status === 0;
}

/**
 * Check if bubblewrap is available
 */
function checkBwrapAvailable(): boolean {
  // eslint-disable-next-line sonarjs/no-os-command-from-path
  const result = spawnSync("which", ["bwrap"], {stdio: "pipe"});
  return result.status === 0;
}

/**
 * Check if /nix/store exists
 */
function checkNixStoreExists(): boolean {
  return existsSync("/nix/store");
}

/**
 * JSON spec format for sets and packages
 */
interface NixJsonSpec {
  sets?: string;
  packages?: string | string[];
  nixpkgs_pin?: string;
}

/**
 * Parse nix config from sandbox config
 * Supports formats:
 * - nix:preset=claude (use default packages)
 * - nix:sets=nodejs,python (expand sets, add to defaults)
 * - nix:packages=nodejs_24,git,python312 (exact packages, no defaults)
 * - nix:{"sets":"nodejs,python","packages":"extra-pkg"} (sets + extra packages)
 * - nix:{"packages":["nodejs_24"],"nixpkgs_pin":"nixos-24.11"}
 */
function parseNixConfig(config: SandboxConfig): NixSandboxConfig {
  const nixConfig: NixSandboxConfig = {};

  if (config.image?.startsWith("nix:")) {
    const spec = config.image.slice(4);

    // Try JSON first
    if (spec.startsWith("{")) {
      try {
        const json = JSON.parse(spec) as NixJsonSpec;
        const packages: string[] = [];

        // Expand sets if present
        if (json.sets) {
          packages.push(...expandPackageSets(json.sets.split(",")));
        }

        // Add extra packages
        if (json.packages) {
          if (Array.isArray(json.packages)) {
            packages.push(...json.packages);
          } else {
            packages.push(...json.packages.split(","));
          }
        }

        if (packages.length > 0) {
          nixConfig.packages = packages;
        }
        if (json.nixpkgs_pin) {
          nixConfig.nixpkgs_pin = json.nixpkgs_pin;
        }
        return nixConfig;
      } catch {
        // Fall through to simple parsing
      }
    }

    // Simple key=value parsing
    if (spec.startsWith("preset=")) {
      // Presets just use defaults
      const preset = spec.slice(7);
      if (preset !== "claude") {
        logError(`Unknown preset "${preset}", using claude defaults`);
      }
    } else if (spec.startsWith("sets=")) {
      // Expand package sets to actual packages (added to defaults)
      const setNames = spec.slice(5).split(",");
      nixConfig.packages = expandPackageSets(setNames);
    } else if (spec.startsWith("packages=")) {
      nixConfig.packages = spec.slice(9).split(",");
      nixConfig.noDefaults = true;
    }
  }

  return nixConfig;
}

/**
 * Build EnvSpec from NixSandboxConfig and agent
 */
function buildEnvSpec(
  nixConfig: NixSandboxConfig,
  agent: AgentConfig,
): EnvSpec {
  let packages: string[];

  if (nixConfig.noDefaults && nixConfig.packages) {
    packages = nixConfig.packages;
  } else {
    // Start with base packages
    packages = [...DEFAULT_BASE_PACKAGES];

    // Add agent-specific package if available
    if (agent.nixPackage) {
      packages.push(agent.nixPackage);
    }

    // Add any extra packages
    if (nixConfig.packages) {
      packages.push(...nixConfig.packages);
    }
  }

  return {
    nixpkgs_pin: nixConfig.nixpkgs_pin ?? DEFAULT_NIXPKGS_PIN,
    packages,
  };
}

/**
 * Get per-agent runtime directories
 */
function getAgentDirs(agentId: string): {
  root: string;
  work: string;
  tmp: string;
  home: string;
} {
  const root = join(getAgentsDir(), agentId);
  return {
    root,
    work: join(root, "work"),
    tmp: join(root, "tmp"),
    home: join(root, "home"),
  };
}

/**
 * Ensure per-agent directories exist
 */
function ensureAgentDirs(agentId: string): ReturnType<typeof getAgentDirs> {
  const dirs = getAgentDirs(agentId);
  mkdirSync(dirs.work, {recursive: true});
  mkdirSync(dirs.tmp, {recursive: true});
  mkdirSync(dirs.home, {recursive: true});
  return dirs;
}

/**
 * Build bubblewrap arguments for running in the Nix environment
 */
function buildBwrapArgs(
  config: SandboxConfig,
  envOutPath: string,
  agentDirs: ReturnType<typeof getAgentDirs>,
): string[] {
  // Use real home path for max compatibility
  const home = homedir();

  // Ensure the mount point exists in agentDirs.home for workDir
  // (bwrap evaluates mount points before applying namespace mounts, so we must
  // pre-create mount points for paths under home in agentDirs.home)
  ensureSandboxMountPoint(agentDirs.home, config.workDir);

  // Get application-level bind mounts (workdir + user config + custom paths)
  const appMounts = getApplicationBindMounts(
    config.workDir,
    home,
    config.bindPaths,
    config.roBindPaths,
  );
  const appMountArgs = bindMountsToBwrapArgs(appMounts);

  // Get sandbox environment with /env/bin prepended to PATH
  // Don't inherit host PATH - nix sandbox is fully isolated with /env/bin
  const baseEnv = getSandboxEnvironment(
    home,
    ["/env/bin"],
    false,
    "/env/bin/bash",
  );
  baseEnv.NIX_SSL_CERT_FILE = "/etc/ssl/certs/ca-certificates.crt";
  const sandboxEnv = mergeCustomEnv(baseEnv, config.customEnv);
  const envArgs = envToBwrapArgs(sandboxEnv);

  // Build args array using spread to avoid multiple push calls
  const args: string[] = [
    // Mount /nix/store read-only
    "--ro-bind",
    "/nix/store",
    "/nix/store",
    // Mount the environment output to /env
    "--ro-bind",
    envOutPath,
    "/env",
    // Mount per-agent writable directories
    "--bind",
    agentDirs.work,
    "/work",
    "--bind",
    agentDirs.tmp,
    "/tmp",
    // Mount agent home to real home path for compatibility
    "--bind",
    agentDirs.home,
    home,
    // Application-level mounts (workdir + user config + custom paths)
    ...appMountArgs,
    // Minimal /proc and /dev
    "--proc",
    "/proc",
    "--dev",
    "/dev",
    // Unshare namespaces (similar to existing bwrap executor)
    "--unshare-uts",
    "--unshare-ipc",
    "--unshare-pid",
    "--unshare-cgroup",
    // Network
    ...(config.network ? ["--share-net"] : ["--unshare-net"]),
    // Environment variables
    ...envArgs,
    // Create /usr/bin/env symlink for scripts with #!/usr/bin/env shebang
    "--symlink",
    "/env/bin/env",
    "/usr/bin/env",
    // SSL certs if they exist
    ...(existsSync("/etc/ssl/certs")
      ? ["--ro-bind", "/etc/ssl/certs", "/etc/ssl/certs"]
      : []),
    ...(existsSync("/etc/resolv.conf")
      ? ["--ro-bind", "/etc/resolv.conf", "/etc/resolv.conf"]
      : []),
    // Working directory
    "--chdir",
    config.workDir,
    // Die with parent
    "--die-with-parent",
  ];

  return args;
}

/**
 * Nix sandbox executor using host Nix + bubblewrap
 */
export const nixExecutor: SandboxExecutor = {
  isAvailable(): Promise<boolean> {
    const nixAvailable = checkNixBuildAvailable();
    const bwrapAvailable = checkBwrapAvailable();
    const storeExists = checkNixStoreExists();

    if (!nixAvailable) {
      logError("nix-build is not available");
    }
    if (!bwrapAvailable) {
      logError("bubblewrap (bwrap) is not available");
    }
    if (!storeExists) {
      logError("/nix/store does not exist");
    }

    return Promise.resolve(nixAvailable && bwrapAvailable && storeExists);
  },

  async execute(config: SandboxConfig): Promise<number> {
    const agent = config.agent ?? claudeAgent;

    // Check prerequisites
    if (!checkNixBuildAvailable()) {
      logError("nix-build is not available. Install Nix first.");
      return 1;
    }
    if (!checkBwrapAvailable()) {
      logError(
        "bubblewrap (bwrap) is not available. Install bubblewrap first.",
      );
      return 1;
    }
    if (!checkNixStoreExists()) {
      logError(
        "/nix/store does not exist. Nix may not be installed correctly.",
      );
      return 1;
    }

    // Parse config and build EnvSpec
    const nixConfig = parseNixConfig(config);
    const envSpec = buildEnvSpec(nixConfig, agent);

    if (config.verbose) {
      logInfo("Building Nix environment...");
      logInfo(`Packages: ${envSpec.packages.join(", ")}`);
      logInfo(`Nixpkgs pin: ${envSpec.nixpkgs_pin}`);
    }

    // Build the environment
    const {resolved, result} = await buildEnv(envSpec, {
      verbose: config.verbose,
      debug: config.debug,
    });

    if (!result.success) {
      logError(
        `Failed to build Nix environment: ${result.error ?? "unknown error"}`,
      );
      return 1;
    }

    const envOutPath = realpathSync(resolved.outLink);

    if (config.verbose) {
      logInfo(`Environment ready: ${envOutPath}`);
    }

    // Create per-agent directories
    const agentDirs = ensureAgentDirs(config.agentId);

    if (config.debug) {
      logDebug(`Agent ID: ${config.agentId}`);
      logDebug(`Agent dirs: ${JSON.stringify(agentDirs)}`);
    }

    // Build bwrap arguments
    const bwrapArgs = buildBwrapArgs(config, envOutPath, agentDirs);

    // Build agent command, wrapped with init commands if present
    // Use /env/bin prefix for the nix sandbox environment
    const agentCommand = buildAgentCommand(config, "/env/bin");
    const fullCommand = wrapWithInitCommands(
      agentCommand,
      config.sandboxInitCommands,
    );

    // Add command separator and full command
    bwrapArgs.push("--", ...fullCommand);

    if (config.debug) {
      logDebug(`bwrap ${bwrapArgs.join(" ")}`);
    }

    if (config.verbose) {
      logInfo(`Starting sandboxed ${agent.displayName} with Nix + bubblewrap`);
    }

    // Build environment
    const env: Record<string, string | undefined> = {
      ...process.env,
      ...getAgentIdEnv(config.agentId),
    };

    if (config.provider) {
      try {
        const providerEnv = buildProviderEnvironment(config.provider);
        const agentEnv = agent.buildEnv(providerEnv);
        Object.assign(env, agentEnv);
      } catch (error) {
        logError(
          error instanceof Error
            ? error.message
            : "Failed to configure provider",
        );
        return 1;
      }
    }

    const command = ["bwrap", ...bwrapArgs];

    // Try terminal wrapper if configured
    const terminalResult = await executeInTerminal(
      config.terminal,
      command,
      env,
      config.workDir,
      config.verbose,
    );
    if (terminalResult !== undefined) {
      return terminalResult;
    }

    // Execute directly with custom error handler for EPERM
    return spawnSandboxProcess(
      "bwrap",
      bwrapArgs,
      {env: env as NodeJS.ProcessEnv, stdio: "inherit"},
      "bwrap",
      (error) => {
        if (
          error.message.includes("EPERM") ||
          error.message.includes("permission")
        ) {
          logError(
            "Failed to create namespace. Ensure unprivileged user namespaces are enabled:\n" +
              "  sysctl kernel.unprivileged_userns_clone=1\n" +
              "Or run with elevated privileges.",
          );
        } else {
          logError(`Failed to execute bwrap: ${error.message}`);
        }
      },
    );
  },

  getCommand(config: SandboxConfig): string[] {
    const agent = config.agent ?? claudeAgent;
    const nixConfig = parseNixConfig(config);
    const envSpec = buildEnvSpec(nixConfig, agent);

    return [
      `# 1. Build environment:`,
      `#    nix-build ~/.local/share/agent-nix/specs/<envId>.nix -o ~/.local/share/agent-nix/envs/<envId>`,
      `# 2. Run with bubblewrap:`,
      `bwrap --ro-bind /nix/store /nix/store --ro-bind $ENV_OUT /env ` +
        `--bind /work /work --bind /tmp /tmp --bind /home /home ` +
        `--proc /proc --dev /dev --chdir ${config.workDir || "/work"} ` +
        `--setenv PATH "/env/bin:/usr/bin:/bin" ` +
        `-- /env/bin/${config.agent?.binary ?? "claude"} ...`,
      `# Packages: ${envSpec.packages.join(", ")}`,
    ];
  },
};
