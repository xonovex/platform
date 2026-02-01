import type {ModelProvider} from "../types.js";

/**
 * OpenAI GPT-5.2 Codex provider configuration for Claude
 * Routes through local CLI Proxy
 */
export const gpt5CodexProvider: ModelProvider = {
  name: "gpt5-codex",
  displayName: "OpenAI GPT-5.2 Codex",
  agentType: "claude",
  authTokenEnv: "CLI_PROXY_API_KEY",
  environment: {
    ANTHROPIC_BASE_URL: "http://127.0.0.1:8317",
    API_TIMEOUT_MS: "3000000",
    CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC: "1",
    ANTHROPIC_DEFAULT_OPUS_MODEL: "gpt-5.2-codex(high)",
    ANTHROPIC_DEFAULT_SONNET_MODEL: "gpt-5.2-codex(medium)",
    ANTHROPIC_DEFAULT_HAIKU_MODEL: "gpt-5.2-codex(low)",
    CLAUDE_CODE_SUBAGENT_MODEL: "gpt-5.2-codex(medium)",
  },
};
