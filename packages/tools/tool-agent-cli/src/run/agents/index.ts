import { claudeAgent } from "./claude/index.js";
import { opencodeAgent } from "./opencode/index.js";
import type { AgentConfig, AgentType } from "./types.js";

/**
 * Registry of all available agents
 */
export const agents = new Map<AgentType, AgentConfig>([
  [claudeAgent.type, claudeAgent],
  [opencodeAgent.type, opencodeAgent],
]);

/**
 * Get an agent by type
 */
export function getAgent(type: AgentType): AgentConfig | undefined {
  return agents.get(type);
}

/**
 * Get all available agent types
 */
export function getAgentTypes(): AgentType[] {
  return Array.from(agents.keys());
}
