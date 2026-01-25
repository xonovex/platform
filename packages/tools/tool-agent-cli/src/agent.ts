#!/usr/bin/env node
import { basename, resolve } from "node:path";
import { logError, logInfo } from "@xonovex/tool-lib";
import { program, type Command } from "commander";
import { generateBashCompletion } from "./completion/bash/generator.js";
import { getAgent, getAgentTypes } from "./run/agents/index.js";
import type { AgentType } from "./run/agents/types.js";
import { combineAgentArgs } from "./run/args.js";
import { generateAgentId } from "./run/id.js";
import { executeAgent } from "./run/index.js";
import { executeInitCommands } from "./run/init/index.js";
import {
  getProvider,
  getProviderNamesForAgent,
} from "./run/providers/index.js";
import { getDefaultSandboxHome } from "./run/sandbox/bindings.js";
import { loadConfigFile, mergeConfig } from "./run/sandbox/config.js";
import { executeSandboxed } from "./run/sandbox/index.js";
import type { SandboxConfig, SandboxMethod } from "./run/sandbox/types.js";
import {
  buildBindPathsWithWorktree,
  getCurrentBranchSync,
  getDefaultWorktreeDir,
  getGitRootSync,
  setupWorktree,
  type WorktreeConfig,
} from "./run/worktree/index.js";
import type { TerminalConfig } from "./run/wrapper/types.js";

/**
 * Options for the run command
 */
interface RunOptions {
  agentId?: string;
  agent: string;
  provider?: string;
  sandbox: string;
  homeDir?: string;
  image?: string;
  composeFile?: string;
  service?: string;
  nixPreset?: string;
  nixSets?: string;
  nixPackages?: string;
  workDir?: string;
  worktreeSourceBranch?: string;
  worktreeBranch?: string;
  worktreeDir?: string;
  initCommand: string[];
  sandboxInitCommand: string[];
  config?: string;
  network: boolean;
  bind: string[];
  roBind: string[];
  env: string[];
  terminal?: string;
  terminalSession?: string;
  terminalWindow?: string;
  terminalDetach?: boolean;
  verbose?: boolean;
  debug?: boolean;
  dryRun?: boolean;
}

const packageJson = {
  name: "@xonovex/tool-agent-cli",
  version: "0.1.0",
  description: "Unified CLI for running AI coding agents",
};

/**
 * Build terminal config from options
 *
 * @param options - Run options
 * @param agentId - Agent identifier (for window name)
 * @param repoDir - Repository directory (for session name, use original repo not worktree)
 * @param sessionBranch - Branch for session name (source branch for worktree)
 * @param windowBranch - Branch for window name (worktree branch)
 */
function buildTerminalConfig(
  options: RunOptions,
  agentId: string,
  repoDir: string,
  sessionBranch?: string,
  windowBranch?: string,
): TerminalConfig | undefined {
  if (!options.terminal) {
    return undefined;
  }

  // Compute defaults:
  // - Session in git repo: <parent>-<repodir>/<sourceBranch>
  // - Session not in git repo: <parent>-<currentdir>
  // - Window with worktree: <worktreeBranch>/<id>
  // - Window without worktree: <id>
  let sessionName = options.terminalSession;
  let windowName = options.terminalWindow;

  if (!sessionName) {
    const gitRoot = getGitRootSync(repoDir);
    if (gitRoot) {
      // In a git repo: <parent>-<repodir>/<branch>
      const parentDir = basename(resolve(gitRoot, ".."));
      const repoName = basename(gitRoot);
      const branch = sessionBranch ?? getCurrentBranchSync(repoDir) ?? "HEAD";
      sessionName = `${parentDir}-${repoName}/${branch}`;
    } else {
      // Not in git repo: <parent>-<currentdir>
      const parentDir = basename(resolve(repoDir, ".."));
      const dirName = basename(repoDir);
      sessionName = `${parentDir}-${dirName}`;
    }
  }

  if (!windowName) {
    // Window name includes agent ID for identification
    // In git repo: <branch>/<id> (e.g., "master/8fa680e" or "feature/test-new/8fa680e")
    // Not in git repo: <id> (e.g., "8fa680e")
    const branch = windowBranch ?? getCurrentBranchSync(repoDir);
    windowName = branch ? `${branch}/${agentId}` : agentId;
  }

  return {
    type: options.terminal,
    sessionName,
    windowName,
    detach: options.terminalDetach ?? false,
    attachExisting: true,
  };
}

/**
 * Helper for collecting multiple values
 */
function collect(value: string, previous: string[]): string[] {
  return [...previous, value];
}

program
  .name("agent-cli")
  .description(packageJson.description)
  .version(packageJson.version);

// Main run command (default)
program
  .command("run", { isDefault: true })
  .description("Run an AI coding agent")
  .option(
    "-a, --agent <type>",
    "Agent to run: " + getAgentTypes().join(", "),
    "claude",
  )
  .option("-p, --provider <name>", "Model provider for the agent")
  .option(
    "-s, --sandbox <method>",
    "Sandbox method: none, bwrap, docker, compose, nix",
    "none",
  )
  .option("-H, --home-dir <dir>", "Custom sandbox home directory")
  .option(
    "-I, --image <image>",
    "Docker image to use (requires agent pre-installed)",
  )
  .option(
    "-F, --compose-file <file>",
    "Docker Compose file (default: stacks/ai-agent.yaml)",
  )
  .option(
    "-S, --service <name>",
    "Docker Compose service name (default: ai-agent)",
  )
  .option(
    "--nix-preset <name>",
    "Nix preset: minimal, nodejs, python, fullstack, infra, complete",
  )
  .option(
    "--nix-sets <sets>",
    "Nix package sets (comma-separated): nodejs,python,kubernetes,etc",
  )
  .option(
    "--nix-packages <packages>",
    "Additional Nix packages (comma-separated): python312,awscli2,etc",
  )
  .option("-w, --work-dir <dir>", "Working directory")
  .option(
    "--worktree-source-branch <branch>",
    "Source branch for worktree (defaults to current branch)",
  )
  .option("--worktree-branch <branch>", "New branch name for the worktree")
  .option("--worktree-dir <dir>", "Directory path for the worktree")
  .option(
    "--init-command <command>",
    "Run command after worktree setup (can be repeated)",
    collect,
    [],
  )
  .option(
    "--sandbox-init-command <command>",
    "Run command inside sandbox before agent (can be repeated)",
    collect,
    [],
  )
  .option("-c, --config <file>", "Load configuration from file")
  .option("-N, --no-network", "Disable network access")
  .option("--network", "Enable network access (default)")
  .option("--bind <path>", "Add read-write bind mount", collect, [])
  .option("--ro-bind <path>", "Add read-only bind mount", collect, [])
  .option(
    "--env <KEY=VALUE>",
    "Set environment variable in sandbox",
    collect,
    [],
  )
  .option("-t, --terminal <wrapper>", "Terminal wrapper: tmux")
  .option("--terminal-session <name>", "Terminal session name")
  .option(
    "--terminal-window <name>",
    "Terminal window name (defaults to directory name)",
  )
  .option("--terminal-detach", "Detach terminal session (run in background)")
  .option(
    "--agent-id <id>",
    "Agent identifier (auto-generated if not provided)",
  )
  .option("-v, --verbose", "Enable verbose output")
  .option("-d, --debug", "Enable debug mode")
  .option("-n, --dry-run", "Show configuration without executing")
  .argument("[agent-args...]", "Arguments to pass to the agent")
  .allowUnknownOption(true)
  .action(async (agentArgs: string[], options: RunOptions, cmd: Command) => {
    // Generate or use provided agent ID
    const agentId = options.agentId ?? generateAgentId();

    // Load config file if specified
    const fileConfig = loadConfigFile(options.config);
    const mergedConfig = mergeConfig(fileConfig, {
      method: options.sandbox as SandboxMethod,
      homeDir: options.homeDir,
      network: options.network,
      bindPaths: options.bind,
      roBindPaths: options.roBind,
      customEnv: options.env,
    });

    // Resolve agent
    const agentType = options.agent as AgentType;
    const agent = getAgent(agentType);
    if (!agent) {
      logError("Unknown agent: " + agentType);
      logInfo("Available agents: " + getAgentTypes().join(", "));
      process.exit(1);
    }

    // Resolve provider (scoped to agent type)
    let provider;
    const providerName = options.provider ?? fileConfig.provider;
    if (providerName) {
      provider = getProvider(providerName, agent.type);
      if (!provider) {
        logError(
          "Unknown provider: " +
            providerName +
            " for agent " +
            agent.displayName,
        );
        const supportedProviders = getProviderNamesForAgent(agent.type);
        if (supportedProviders.length > 0) {
          logInfo(
            "Available providers for " +
              agent.displayName +
              ": " +
              supportedProviders.join(", "),
          );
        } else {
          logInfo("No providers currently available for " + agent.displayName);
        }
        process.exit(1);
      }
    }

    // Resolve work directory from -w option or current directory
    const originalWorkDir = resolve(options.workDir ?? process.cwd());
    let workDir = originalWorkDir;
    let worktreeSourceDir: string | undefined;

    // Setup worktree if configured
    if (options.worktreeBranch) {
      const repoName = basename(workDir);
      const worktreeDir =
        options.worktreeDir ??
        getDefaultWorktreeDir(options.worktreeBranch, repoName);
      const worktreeConfig: WorktreeConfig = {
        sourceBranch: options.worktreeSourceBranch,
        branch: options.worktreeBranch,
        dir: worktreeDir,
      };
      try {
        workDir = await setupWorktree(
          worktreeConfig,
          originalWorkDir,
          options.verbose ?? false,
        );
        // Track the original repo dir for sandbox binding (worktree needs access to .git)
        worktreeSourceDir = originalWorkDir;
      } catch {
        process.exit(1);
      }
    }

    // Run init commands if specified
    if (options.initCommand.length > 0) {
      try {
        executeInitCommands(
          options.initCommand,
          workDir,
          options.verbose ?? false,
        );
      } catch {
        process.exit(1);
      }
    }

    // Build terminal config with git context
    // Session: <parent>-<repo>/<sourceBranch>, Window: <worktreeBranch>/<id> or <id>
    // For worktree: session uses source branch, window uses worktree branch
    const sessionBranch = options.worktreeBranch
      ? (options.worktreeSourceBranch ?? getCurrentBranchSync(originalWorkDir))
      : undefined;
    const terminalConfig = buildTerminalConfig(
      options,
      agentId,
      originalWorkDir,
      sessionBranch ?? undefined,
      options.worktreeBranch,
    );

    // Combine agent args from positional and unknown options (after --)
    const allAgentArgs = combineAgentArgs(agentArgs, cmd.args);

    const method = mergedConfig.method ?? "none";

    // If using sandbox (bwrap or docker), use sandboxed execution
    if (method !== "none") {
      // Build image spec for nix (encode preset/sets/packages in image field)
      let image = options.image;
      if (method === "nix" && !image) {
        if (options.nixPreset) {
          image = `nix:preset=${options.nixPreset}`;
        } else if (options.nixSets || options.nixPackages) {
          // Build JSON spec for sets and/or packages
          const spec: { sets?: string; packages?: string } = {};
          if (options.nixSets) {
            spec.sets = options.nixSets;
          }
          if (options.nixPackages) {
            spec.packages = options.nixPackages;
          }
          image = `nix:${JSON.stringify(spec)}`;
        }
      }

      // Include original repo dir in bind paths if using worktree (for .git access)
      const bindPaths = buildBindPathsWithWorktree(
        mergedConfig.bindPaths ?? [],
        worktreeSourceDir,
      );

      if (options.verbose) {
        logInfo(`Agent ID: ${agentId}`);
      }

      const config: SandboxConfig = {
        agentId,
        method,
        agent,
        homeDir: mergedConfig.homeDir ?? getDefaultSandboxHome(),
        image,
        composeFile: options.composeFile,
        service: options.service,
        workDir,
        repoDir: worktreeSourceDir,
        network: mergedConfig.network ?? true,
        bindPaths,
        roBindPaths: mergedConfig.roBindPaths ?? [],
        customEnv: mergedConfig.customEnv ?? [],
        provider,
        agentArgs: allAgentArgs,
        sandboxInitCommands:
          options.sandboxInitCommand.length > 0
            ? options.sandboxInitCommand
            : undefined,
        verbose: options.verbose ?? false,
        debug: options.debug ?? false,
        dryRun: options.dryRun ?? false,
        terminal: terminalConfig,
      };

      const code = await executeSandboxed(config);
      process.exit(code);
    }

    // Direct execution (no sandbox)
    if (options.dryRun) {
      logInfo("Dry run mode - would execute:");
      logInfo("  Agent: " + agent.displayName);
      if (provider) {
        logInfo("  Provider: " + provider.displayName);
      }
      logInfo("  Work directory: " + workDir);
      logInfo(
        "  Arguments: " +
          (allAgentArgs.length > 0 ? allAgentArgs.join(" ") : "(none)"),
      );
      process.exit(0);
    }

    const code = await executeAgent({
      agent,
      provider,
      args: allAgentArgs,
      cwd: workDir,
      verbose: options.verbose ?? false,
      terminal: terminalConfig,
    });
    process.exit(code);
  });

// Completion command
program
  .command("completion")
  .description("Generate shell completion script")
  .action(() => {
    console.log(generateBashCompletion());
  });

// Error handling
program.configureOutput({
  writeErr: (str) => {
    logError(str.trim());
  },
});

program.exitOverride((err) => {
  if (err.code === "commander.help") {
    process.exit(0);
  }
  if (err.code === "commander.version") {
    process.exit(0);
  }
  logError(err.message);
  process.exit(1);
});

// Parse command line
program.parse();
