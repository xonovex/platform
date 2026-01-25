/**
 * Shared constants for the script-agent package
 */

/**
 * Bash reserved environment variables that should not be modified or passed through sandboxes.
 * These are read-only in bash and attempting to set them can cause issues.
 */
export const BASH_RESERVED_ENV_VARS = ["UID", "EUID", "GID", "GROUPS"] as const;

/**
 * Default UID fallback when process.getuid() is unavailable (e.g., on Windows).
 * Linux/Unix systems typically start user IDs at 1000.
 */
export const DEFAULT_FALLBACK_UID = 1000;

/**
 * Default GID fallback when process.getgid() is unavailable (e.g., on Windows).
 * Linux/Unix systems typically start group IDs at 1000.
 */
export const DEFAULT_FALLBACK_GID = 1000;

/**
 * Default description for custom agents when none is provided.
 */
export const DEFAULT_AGENT_DESCRIPTION = "Custom agent";
