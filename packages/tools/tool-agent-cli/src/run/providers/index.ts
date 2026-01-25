import type { AgentType } from "../agents/types.js";
import { claudeProviders } from "./claude/index.js";
import { opencodeProviders } from "./opencode/index.js";
import type { ModelProvider } from "./types.js";

/**
 * All providers from all agents
 */
const allProviders: ModelProvider[] = [
  ...claudeProviders,
  ...opencodeProviders,
];

/**
 * Registry of all available model providers (indexed by agentType:name)
 */
export const providers = new Map<string, ModelProvider>(
  allProviders.map((p) => [p.agentType + ":" + p.name, p]),
);

/**
 * Get a provider by name, optionally filtered by agent type
 * If agentType is provided, looks for agentType:name first
 * Otherwise searches all providers for matching name
 */
export function getProvider(
  name: string,
  agentType?: AgentType,
): ModelProvider | undefined {
  // If agent type specified, try direct lookup
  if (agentType) {
    const key = agentType + ":" + name;
    const provider = providers.get(key);
    if (provider) return provider;
  }

  // Search for first matching provider by name
  for (const provider of allProviders) {
    if (provider.name === name) {
      // If agent type specified, only return if it matches
      if (agentType && provider.agentType !== agentType) {
        continue;
      }
      return provider;
    }
  }

  return undefined;
}

/**
 * Get providers that are for a specific agent type
 */
function getProvidersForAgent(agentType: AgentType): ModelProvider[] {
  return allProviders.filter((p) => p.agentType === agentType);
}

/**
 * Get provider names for a specific agent type
 */
export function getProviderNamesForAgent(agentType: AgentType): string[] {
  return getProvidersForAgent(agentType).map((p) => p.name);
}
