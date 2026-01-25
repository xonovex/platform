import type { AgentType } from "../agents/types.js";

/**
 * Base configuration for a model provider
 */
export interface ModelProvider {
  /** Provider identifier (e.g., "gemini", "glm") */
  name: string;

  /** Display name for the provider */
  displayName: string;

  /** Which agent type this provider is for */
  agentType: AgentType;

  /** Environment variable name containing the auth token (optional) */
  authTokenEnv?: string;

  /** Environment variables to set when using this provider */
  environment: Record<string, string>;

  /** CLI arguments to add when using this provider (e.g., --model for opencode) */
  cliArgs?: string[];
}

/**
 * Build environment variables from a provider configuration
 * Returns a record that can be spread into process.env
 */
export function buildProviderEnvironment(
  provider: ModelProvider,
): Record<string, string> {
  // Create a copy of the environment
  const env: Record<string, string> = { ...provider.environment };

  // If auth token is required, inject it
  if (provider.authTokenEnv) {
    const authToken = process.env[provider.authTokenEnv];
    if (!authToken) {
      throw new Error(
        `Missing authentication token: ${provider.authTokenEnv} environment variable is not set`,
      );
    }

    // Inject auth token for Anthropic-compatible providers (Claude)
    if (env.ANTHROPIC_BASE_URL) {
      env.ANTHROPIC_AUTH_TOKEN = authToken;
    }
  }

  return env;
}

/**
 * Get CLI arguments to add for a provider
 */
export function getProviderCliArgs(provider: ModelProvider): string[] {
  return provider.cliArgs ?? [];
}
