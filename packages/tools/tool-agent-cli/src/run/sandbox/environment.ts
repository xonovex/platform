import { join } from "node:path";

/**
 * Environment variable configuration for sandboxes
 */
export interface SandboxEnv {
  HOME: string;
  TMPDIR: string;
  PATH: string;
  SHELL: string;
  [key: string]: string;
}

/**
 * Get base environment variables for sandboxes
 * These are the common environment variables needed by all sandbox types
 *
 * @param home - The home directory inside the sandbox
 * @param extraPathPrefixes - Additional paths to prepend to PATH (e.g., "/env/bin")
 * @param inheritHostPath - Whether to include the host's PATH (default: true for host-based sandboxes)
 * @param shell - Shell path to use (default: "/bin/bash")
 */
export function getSandboxEnvironment(
  home: string,
  extraPathPrefixes: string[] = [],
  inheritHostPath = true,
  shell = "/bin/bash",
): SandboxEnv {
  const pathParts = [
    ...extraPathPrefixes,
    join(home, ".local", "bin"),
    "/usr/local/bin",
    "/usr/bin",
    "/bin",
  ];

  // Include host PATH for host-based sandboxes (bwrap, none)
  // This ensures binaries installed in non-standard locations are available
  if (inheritHostPath && process.env.PATH) {
    pathParts.push(process.env.PATH);
  }

  return {
    HOME: home,
    TMPDIR: "/tmp",
    PATH: pathParts.join(":"),
    SHELL: shell,
  };
}

/**
 * Convert environment variables to bubblewrap --setenv arguments
 */
export function envToBwrapArgs(env: SandboxEnv): string[] {
  const args: string[] = [];
  for (const [key, value] of Object.entries(env)) {
    args.push("--setenv", key, value);
  }
  return args;
}

/**
 * Convert environment variables to Docker -e arguments
 */
export function envToDockerArgs(env: SandboxEnv): string[] {
  const args: string[] = [];
  for (const [key, value] of Object.entries(env)) {
    args.push("-e", `${key}=${value}`);
  }
  return args;
}

/**
 * Parse custom environment variables from KEY=VALUE strings
 * Returns a record of key-value pairs
 */
export function parseCustomEnv(customEnv: string[]): Record<string, string> {
  const env: Record<string, string> = {};
  for (const entry of customEnv) {
    const eqIndex = entry.indexOf("=");
    if (eqIndex > 0) {
      const key = entry.slice(0, eqIndex);
      const value = entry.slice(eqIndex + 1);
      env[key] = value;
    }
  }
  return env;
}

/**
 * Merge custom environment variables into sandbox environment
 * Custom values override defaults
 */
export function mergeCustomEnv(
  baseEnv: SandboxEnv,
  customEnv: string[],
): SandboxEnv {
  const custom = parseCustomEnv(customEnv);
  return {
    ...baseEnv,
    ...custom,
  };
}
