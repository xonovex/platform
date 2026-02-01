import type {ModelProvider} from "../types.js";

/**
 * Google Gemini-Claude hybrid thinking models provider configuration for Claude
 * Routes through local CLI Proxy
 */
export const geminiClaudeProvider: ModelProvider = {
  name: "gemini-claude",
  displayName: "Gemini-Claude Thinking",
  agentType: "claude",
  authTokenEnv: "CLI_PROXY_API_KEY",
  environment: {
    ANTHROPIC_BASE_URL: "http://127.0.0.1:8317",
    API_TIMEOUT_MS: "3000000",
    CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC: "1",
    ANTHROPIC_DEFAULT_OPUS_MODEL: "gemini-claude-opus-4-5-thinking",
    ANTHROPIC_DEFAULT_SONNET_MODEL: "gemini-claude-sonnet-4-5-thinking",
    ANTHROPIC_DEFAULT_HAIKU_MODEL: "gemini-2.5-flash-lite",
    CLAUDE_CODE_SUBAGENT_MODEL: "gemini-claude-sonnet-4-5",
  },
};
