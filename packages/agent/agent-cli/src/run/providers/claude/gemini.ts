import type {ModelProvider} from "../types.js";

/**
 * Google Gemini 3.x provider configuration for Claude
 * Routes through local CLI Proxy
 */
export const geminiProvider: ModelProvider = {
  name: "gemini",
  displayName: "Google Gemini 3.x",
  agentType: "claude",
  authTokenEnv: "CLI_PROXY_API_KEY",
  environment: {
    ANTHROPIC_BASE_URL: "http://127.0.0.1:8317",
    API_TIMEOUT_MS: "3000000",
    CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC: "1",
    ANTHROPIC_DEFAULT_OPUS_MODEL: "gemini-3-pro-preview",
    ANTHROPIC_DEFAULT_SONNET_MODEL: "gemini-3-flash-preview",
    ANTHROPIC_DEFAULT_HAIKU_MODEL: "gemini-2.5-flash-lite",
    CLAUDE_CODE_SUBAGENT_MODEL: "gemini-3-flash-preview",
  },
};
