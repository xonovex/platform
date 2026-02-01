import {execFile, execSync} from "node:child_process";
import {existsSync, readFileSync, statSync} from "node:fs";
import {resolve} from "node:path";
import {promisify} from "node:util";
import {logError, logInfo} from "@xonovex/core";

const execFileAsync = promisify(execFile);

/**
 * Get the git repository root directory (synchronous)
 * Returns null if not in a git repository
 */
export function getGitRootSync(cwd?: string): string | null {
  try {
    // eslint-disable-next-line sonarjs/no-os-command-from-path
    const result = execSync("git rev-parse --show-toplevel", {
      cwd,
      stdio: ["pipe", "pipe", "pipe"],
      encoding: "utf8",
    });
    return result.trim();
  } catch {
    return null;
  }
}

/**
 * Get the current git branch name (synchronous)
 * Returns null if not in a git repository or on detached HEAD
 */
export function getCurrentBranchSync(cwd?: string): string | null {
  try {
    // eslint-disable-next-line sonarjs/no-os-command-from-path
    const result = execSync("git rev-parse --abbrev-ref HEAD", {
      cwd,
      stdio: ["pipe", "pipe", "pipe"],
      encoding: "utf8",
    });
    const branch = result.trim();
    return branch === "HEAD" ? null : branch;
  } catch {
    return null;
  }
}

/**
 * Sanitize a branch name to create a valid directory name
 * Replaces slashes and other special characters with hyphens
 */
export function sanitizeBranchName(branch: string): string {
  return branch
    .replaceAll(/[/\\]/g, "-") // Replace slashes with hyphens
    .replaceAll(/[^\w-]/g, "-") // Replace other special chars with hyphens
    .replaceAll(/-+/g, "-") // Collapse multiple hyphens
    .replaceAll(/^-|-$/g, ""); // Remove leading/trailing hyphens
}

/**
 * Generate a default worktree directory path from a branch name and repo name
 * Returns a path like "../<repo-name>-<sanitized-branch-name>"
 */
export function getDefaultWorktreeDir(
  branch: string,
  repoName: string,
): string {
  const sanitizedRepo = sanitizeBranchName(repoName);
  const sanitizedBranch = sanitizeBranchName(branch);
  return `../${sanitizedRepo}-${sanitizedBranch}`;
}

/**
 * Build bind paths array including the worktree source directory if present.
 * The worktree source directory (original repo with .git) must be bound
 * in sandbox mode for git operations to work.
 */
export function buildBindPathsWithWorktree(
  baseBindPaths: string[],
  worktreeSourceDir: string | undefined,
): string[] {
  return [...baseBindPaths, ...(worktreeSourceDir ? [worktreeSourceDir] : [])];
}

/**
 * Configuration for creating a git worktree
 */
export interface WorktreeConfig {
  /** Source branch for the worktree (defaults to current branch if not specified) */
  sourceBranch?: string;
  /** New branch name for the worktree */
  branch: string;
  /** Directory path for the worktree */
  dir: string;
}

/**
 * Execute a git command and return stdout
 */
async function execGit(args: string[], cwd: string): Promise<string> {
  const {stdout} = await execFileAsync("git", args, {cwd});
  return stdout.trim();
}

/**
 * Get the current branch name
 */
export async function getCurrentBranch(cwd: string): Promise<string> {
  return execGit(["rev-parse", "--abbrev-ref", "HEAD"], cwd);
}

/**
 * Check if a directory is a git worktree (has a .git file pointing to the main repo)
 */
export function isWorktreeDirectory(dir: string): boolean {
  const gitPath = resolve(dir, ".git");
  if (!existsSync(gitPath)) {
    return false;
  }
  // Worktrees have a .git file, not a directory
  const stats = statSync(gitPath);
  if (!stats.isFile()) {
    return false;
  }
  // Check if it contains a gitdir pointer
  const content = readFileSync(gitPath, "utf8");
  return content.startsWith("gitdir:");
}

/**
 * Check if a worktree belongs to a specific repository
 */
export async function isWorktreeForRepo(
  worktreeDir: string,
  repoDir: string,
): Promise<boolean> {
  try {
    // Get the git directory of the worktree
    const worktreeGitDir = await execGit(
      ["rev-parse", "--git-dir"],
      worktreeDir,
    );
    // Get the git directory of the main repo
    const repoGitDir = await execGit(["rev-parse", "--git-dir"], repoDir);

    // Worktree git dir should be inside the repo's .git/worktrees/
    const resolvedWorktreeGitDir = resolve(worktreeDir, worktreeGitDir);
    const resolvedRepoGitDir = resolve(repoDir, repoGitDir);

    return resolvedWorktreeGitDir.startsWith(resolvedRepoGitDir);
  } catch {
    return false;
  }
}

/**
 * Result of checking an existing worktree
 */
export interface ExistingWorktreeCheck {
  exists: boolean;
  isWorktree: boolean;
  isForThisRepo: boolean;
  currentBranch?: string;
}

/**
 * Check the status of an existing directory that might be a worktree
 */
export async function checkExistingWorktree(
  dir: string,
  repoDir: string,
): Promise<ExistingWorktreeCheck> {
  const resolvedDir = resolve(repoDir, dir);

  if (!existsSync(resolvedDir)) {
    return {exists: false, isWorktree: false, isForThisRepo: false};
  }

  if (!isWorktreeDirectory(resolvedDir)) {
    return {exists: true, isWorktree: false, isForThisRepo: false};
  }

  const isForThisRepo = await isWorktreeForRepo(resolvedDir, repoDir);
  if (!isForThisRepo) {
    return {exists: true, isWorktree: true, isForThisRepo: false};
  }

  try {
    const currentBranch = await getCurrentBranch(resolvedDir);
    return {
      exists: true,
      isWorktree: true,
      isForThisRepo: true,
      currentBranch,
    };
  } catch {
    return {exists: true, isWorktree: true, isForThisRepo: true};
  }
}

/**
 * Check if a branch exists
 */
async function branchExists(branch: string, cwd: string): Promise<boolean> {
  try {
    await execGit(["rev-parse", "--verify", `refs/heads/${branch}`], cwd);
    return true;
  } catch {
    return false;
  }
}

/**
 * Create a git worktree for an existing branch
 */
async function createWorktreeForExistingBranch(
  dir: string,
  branch: string,
  cwd: string,
): Promise<void> {
  await execGit(["worktree", "add", dir, branch], cwd);
}

/**
 * Create a git worktree with a new branch
 */
async function createWorktreeWithNewBranch(
  dir: string,
  branch: string,
  sourceBranch: string,
  cwd: string,
): Promise<void> {
  await execGit(["worktree", "add", dir, "-b", branch, sourceBranch], cwd);
}

/**
 * Set the mergeBackTo config for a branch
 */
async function setMergeBackConfig(
  branch: string,
  sourceBranch: string,
  cwd: string,
): Promise<void> {
  await execGit(["config", `branch.${branch}.mergeBackTo`, sourceBranch], cwd);
}

/**
 * Setup a git worktree and return the worktree directory path
 *
 * If the worktree already exists with the correct branch, it will be reused.
 * Otherwise, creates a new worktree at the specified directory on a new branch,
 * and sets the mergeBackTo config for tracking the source branch.
 *
 * @param config - Worktree configuration
 * @param cwd - Current working directory (git repository)
 * @param verbose - Enable verbose output
 * @returns Absolute path to the worktree directory
 */
export async function setupWorktree(
  config: WorktreeConfig,
  cwd: string,
  verbose = false,
): Promise<string> {
  const resolvedDir = resolve(cwd, config.dir);

  // Check if worktree already exists
  const existing = await checkExistingWorktree(config.dir, cwd);

  if (existing.exists) {
    // Directory exists - check if we can reuse it
    if (!existing.isWorktree) {
      logError(`Directory ${config.dir} exists but is not a git worktree`);
      throw new Error(`Directory exists but is not a worktree: ${config.dir}`);
    }

    if (!existing.isForThisRepo) {
      logError(
        `Worktree ${config.dir} exists but belongs to a different repository`,
      );
      throw new Error(
        `Worktree belongs to different repository: ${config.dir}`,
      );
    }

    const existingBranch = existing.currentBranch ?? "unknown";
    if (existingBranch !== config.branch) {
      logError(
        `Worktree ${config.dir} exists on branch '${existingBranch}', expected '${config.branch}'`,
      );
      throw new Error(
        `Worktree on wrong branch: expected '${config.branch}', found '${existingBranch}'`,
      );
    }

    // Worktree exists and is valid - reuse it
    if (verbose) {
      logInfo(
        `Reusing existing worktree at ${config.dir} on branch ${config.branch}`,
      );
    }
    return resolvedDir;
  }

  // Worktree doesn't exist - check if branch exists
  const branchAlreadyExists = await branchExists(config.branch, cwd);

  if (branchAlreadyExists) {
    // Branch exists - create worktree for existing branch
    if (verbose) {
      logInfo(
        `Creating worktree at ${config.dir} for existing branch ${config.branch}`,
      );
    }

    try {
      await createWorktreeForExistingBranch(config.dir, config.branch, cwd);
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      logError(`Failed to create worktree: ${message}`);
      throw error;
    }

    if (verbose) {
      logInfo(`Worktree created successfully for existing branch`);
      logInfo(`  Branch: ${config.branch}`);
    }
  } else {
    // Branch doesn't exist - create new branch and worktree
    const sourceBranch = config.sourceBranch ?? (await getCurrentBranch(cwd));

    if (verbose) {
      logInfo(
        `Creating worktree at ${config.dir} on new branch ${config.branch} from ${sourceBranch}`,
      );
    }

    try {
      await createWorktreeWithNewBranch(
        config.dir,
        config.branch,
        sourceBranch,
        cwd,
      );
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      logError(`Failed to create worktree: ${message}`);
      throw error;
    }

    // Set the mergeBackTo config only for new branches
    try {
      await setMergeBackConfig(config.branch, sourceBranch, cwd);
    } catch (error) {
      const message = error instanceof Error ? error.message : String(error);
      logError(`Failed to set mergeBackTo config: ${message}`);
      throw error;
    }

    if (verbose) {
      logInfo(`Worktree created successfully`);
      logInfo(`  Branch: ${config.branch}`);
      logInfo(`  Source: ${sourceBranch}`);
      logInfo(`  mergeBackTo: ${sourceBranch}`);
    }
  }

  // Return the absolute worktree path
  return resolvedDir;
}
