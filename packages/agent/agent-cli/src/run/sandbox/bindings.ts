import {existsSync, mkdirSync} from "node:fs";
import {homedir} from "node:os";
import {join, relative} from "node:path";
import {logInfo} from "@xonovex/core";

/**
 * Mount type for bubblewrap
 */
export type MountType = "bind" | "ro-bind" | "dev-bind" | "proc";

/**
 * Default bind mount configuration for bubblewrap
 */
export interface BindMount {
  source: string;
  target: string;
  readOnly: boolean;
  type?: MountType;
}

/**
 * Get the default sandbox home directory
 */
export function getDefaultSandboxHome(): string {
  return join(homedir(), ".sandboxed-agent-home");
}

/**
 * User config directories/files that should be bind mounted into sandboxes
 * These are relative to the user's home directory
 */
const USER_CONFIG_PATHS = [
  // Claude Code config
  ".claude",
  ".claude.json",
  // Git config
  ".gitconfig",
  ".gitignore_global",
  // SSH (for git operations)
  ".ssh",
  // General config
  ".config",
  // NPM config
  ".npmrc",
  ".npm",
  ".npm-global",
  // Other tool configs
  ".cargo",
  ".rustup",
  ".local",
  ".cache",
];

/**
 * Get user config bind mounts for sandboxes
 * Maps from real home paths to target home directory
 *
 * @param targetHome - The home directory inside the sandbox (e.g., "/home" or the real homedir)
 * @param readOnly - Whether to mount read-only (default: false)
 */
export function getUserConfigBindMounts(
  targetHome: string,
  readOnly = false,
): BindMount[] {
  const home = homedir();
  const mounts: BindMount[] = [];

  for (const configPath of USER_CONFIG_PATHS) {
    const sourcePath = join(home, configPath);
    if (existsSync(sourcePath)) {
      mounts.push({
        source: sourcePath,
        target: join(targetHome, configPath),
        readOnly,
      });
    }
  }

  return mounts;
}

/**
 * Get system-level bind mounts (/, /dev, /proc, /tmp)
 * These are needed for bwrap-style sandboxes that overlay on the host filesystem
 */
export function getSystemBindMounts(): BindMount[] {
  return [
    {source: "/", target: "/", readOnly: true},
    {source: "/dev", target: "/dev", readOnly: false, type: "dev-bind"},
    {source: "/proc", target: "/proc", readOnly: false, type: "proc"},
    {source: "/tmp", target: "/tmp", readOnly: false},
  ];
}

/**
 * Get application-level bind mounts (workdir + user config)
 * These are needed by ALL sandbox types
 *
 * @param workDir - The working directory to mount
 * @param targetHome - The home directory inside the sandbox
 * @param customBindPaths - Additional read-write bind paths
 * @param customRoBindPaths - Additional read-only bind paths
 */
export function getApplicationBindMounts(
  workDir: string,
  targetHome: string,
  customBindPaths: string[] = [],
  customRoBindPaths: string[] = [],
): BindMount[] {
  const mounts: BindMount[] = [
    // Working directory at same path
    {source: workDir, target: workDir, readOnly: false},
    // User config directories
    ...getUserConfigBindMounts(targetHome),
    // Custom bind paths
    ...customBindPaths
      .filter((path) => existsSync(path))
      .map((path) => ({source: path, target: path, readOnly: false})),
    ...customRoBindPaths
      .filter((path) => existsSync(path))
      .map((path) => ({source: path, target: path, readOnly: true})),
  ];

  return mounts;
}

/**
 * Ensure the mount point directory exists in sandboxHome for paths under the home directory.
 *
 * When bwrap maps sandboxHome to the user's home directory, bind mounts for paths
 * under home need the mount point to exist in sandboxHome BEFORE bwrap runs.
 * bwrap evaluates mount points before applying the namespace, so it can't create
 * mount points inside an overlaid directory.
 *
 * @param sandboxHome - The sandbox home directory
 * @param targetPath - The path that will be bind mounted (the mount point)
 */
export function ensureSandboxMountPoint(
  sandboxHome: string,
  targetPath: string,
  verbose = false,
): void {
  const home = homedir();

  if (verbose) {
    logInfo(
      `Ensuring sandbox mount point: sandboxHome=${sandboxHome}, targetPath=${targetPath}`,
    );
  }

  // Check if targetPath is under home
  if (!targetPath.startsWith(home + "/")) {
    if (verbose) {
      logInfo(
        `Target path not under home (${home}), skipping sandbox mount point creation`,
      );
    }
    return;
  }

  // Get the relative path from home to the target (the mount point itself)
  const relativePath = relative(home, targetPath);

  if (relativePath && !relativePath.startsWith("..")) {
    const mountPointInSandbox = join(sandboxHome, relativePath);
    if (verbose) {
      logInfo(`Creating sandbox mount point: ${mountPointInSandbox}`);
    }
    mkdirSync(mountPointInSandbox, {recursive: true});
  }
}

/**
 * @deprecated Use ensureSandboxMountPoint instead
 */
export function ensureSandboxParentDirs(
  sandboxHome: string,
  targetPath: string,
  verbose = false,
): void {
  ensureSandboxMountPoint(sandboxHome, targetPath, verbose);
}

/**
 * Get base system mounts including sandbox home overlay.
 * These need to be mounted before --dir can create directories under home.
 */
export function getBaseMounts(sandboxHome: string): BindMount[] {
  const home = homedir();
  return [
    // System-level mounts
    ...getSystemBindMounts(),
    // Sandbox home mapped to real home (must come before --dir for paths under home)
    {source: sandboxHome, target: home, readOnly: false},
  ];
}

/**
 * Get bind mounts for host overlay sandboxes (bubblewrap)
 * These sandboxes mount the host root filesystem read-only and overlay writable directories
 * Includes: system mounts (/, /dev, /proc, /tmp) + sandbox home + application mounts
 */
export function getHostOverlayBindMounts(
  sandboxHome: string,
  workDir: string,
): BindMount[] {
  const home = homedir();
  // Use special mount types for /dev and /proc to avoid Bun segfaults
  // --dev-bind properly handles device nodes, --proc mounts a new procfs
  const mounts: BindMount[] = [
    // System-level mounts
    ...getSystemBindMounts(),
    // Sandbox home mapped to real home
    {source: sandboxHome, target: home, readOnly: false},
    // Application-level mounts (workdir + user config)
    ...getApplicationBindMounts(workDir, home),
  ];

  return mounts;
}

/**
 * Convert bind mounts to bubblewrap arguments
 */
export function bindMountsToBwrapArgs(mounts: BindMount[]): string[] {
  const args: string[] = [];

  for (const mount of mounts) {
    if (mount.type === "dev-bind") {
      args.push("--dev-bind", mount.source, mount.target);
    } else if (mount.type === "proc") {
      args.push("--proc", mount.target);
    } else if (mount.readOnly) {
      args.push("--ro-bind", mount.source, mount.target);
    } else {
      args.push("--bind", mount.source, mount.target);
    }
  }

  return args;
}

/**
 * Convert bind mounts to Docker volume arguments
 */
export function bindMountsToDockerArgs(mounts: BindMount[]): string[] {
  const args: string[] = [];

  for (const mount of mounts) {
    const mode = mount.readOnly ? "ro" : "rw";
    args.push("-v", `${mount.source}:${mount.target}:${mode}`);
  }

  return args;
}
