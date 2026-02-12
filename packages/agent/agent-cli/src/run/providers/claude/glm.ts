import type {ModelProvider} from "../types.js";

/**
 * Zhipu AI GLM provider configuration for Claude
 * Routes through Z.AI Anthropic-compatible API
 */
export const glmProvider: ModelProvider = {
  name: "glm",
  displayName: "Zhipu AI GLM",
  agentType: "claude",
  authTokenEnv: "ZAI_AUTH_TOKEN",
  environment: {
    ANTHROPIC_BASE_URL: "https://api.z.ai/api/anthropic",
    API_TIMEOUT_MS: "3000000",
    CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC: "1",
    ANTHROPIC_DEFAULT_OPUS_MODEL: "GLM-5",
    ANTHROPIC_DEFAULT_SONNET_MODEL: "GLM-4.7",
    ANTHROPIC_DEFAULT_HAIKU_MODEL: "GLM-4.5-Air",
  },
};
