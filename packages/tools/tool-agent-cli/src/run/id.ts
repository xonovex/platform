/**
 * Agent ID generation and management
 */

/** Environment variable for the current agent ID */
export const AGENT_ID_ENV = "XONOVEX_AGENT_ID";

/** Environment variable for the parent agent ID */
export const PARENT_AGENT_ID_ENV = "XONOVEX_PARENT_AGENT_ID";

/**
 * Generate a random agent ID similar to a short git hash
 * Format: 7 lowercase hex characters
 * Example: a3f2b9c
 */
export function generateAgentId(): string {
  const bytes = new Uint8Array(4);
  crypto.getRandomValues(bytes);
  return Array.from(bytes)
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("")
    .slice(0, 7);
}

/**
 * Get agent ID environment variables to pass to child processes.
 * If there's an existing agent ID in the environment, it becomes the parent.
 */
export function getAgentIdEnv(agentId: string): Record<string, string> {
  const env: Record<string, string> = {
    [AGENT_ID_ENV]: agentId,
  };

  // If there's an existing agent ID, it becomes the parent
  const existingAgentId = process.env[AGENT_ID_ENV];
  if (existingAgentId) {
    env[PARENT_AGENT_ID_ENV] = existingAgentId;
  }

  return env;
}
