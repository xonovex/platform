import { afterEach, beforeEach, describe, expect, it } from "vitest";
import { geminiClaudeProvider } from "../../src/run/providers/claude/gemini-claude.js";
import { geminiProvider as claudeGeminiProvider } from "../../src/run/providers/claude/gemini.js";
import { glmProvider } from "../../src/run/providers/claude/glm.js";
import { gpt5CodexProvider } from "../../src/run/providers/claude/gpt5-codex.js";
import {
  getProvider,
  getProviderNamesForAgent,
  providers,
} from "../../src/run/providers/index.js";
import { geminiProvider as opencodeGeminiProvider } from "../../src/run/providers/opencode/gemini.js";
import {
  buildProviderEnvironment,
  getProviderCliArgs,
} from "../../src/run/providers/types.js";

describe("providers", () => {
  describe("claude provider configurations", () => {
    it("should have correct gemini provider config", () => {
      expect(claudeGeminiProvider.name).toBe("gemini");
      expect(claudeGeminiProvider.displayName).toBe("Google Gemini 3.x");
      expect(claudeGeminiProvider.agentType).toBe("claude");
      expect(claudeGeminiProvider.authTokenEnv).toBe("CLI_PROXY_API_KEY");
      expect(claudeGeminiProvider.environment.ANTHROPIC_BASE_URL).toBe(
        "http://127.0.0.1:8317",
      );
      expect(
        claudeGeminiProvider.environment.ANTHROPIC_DEFAULT_OPUS_MODEL,
      ).toBe("gemini-3-pro-preview");
    });

    it("should have correct gemini-claude provider config", () => {
      expect(geminiClaudeProvider.name).toBe("gemini-claude");
      expect(geminiClaudeProvider.displayName).toBe("Gemini-Claude Thinking");
      expect(geminiClaudeProvider.agentType).toBe("claude");
      expect(
        geminiClaudeProvider.environment.ANTHROPIC_DEFAULT_OPUS_MODEL,
      ).toBe("gemini-claude-opus-4-5-thinking");
    });

    it("should have correct glm provider config", () => {
      expect(glmProvider.name).toBe("glm");
      expect(glmProvider.displayName).toBe("Zhipu AI GLM");
      expect(glmProvider.agentType).toBe("claude");
      expect(glmProvider.authTokenEnv).toBe("ZAI_AUTH_TOKEN");
      expect(glmProvider.environment.ANTHROPIC_BASE_URL).toBe(
        "https://api.z.ai/api/anthropic",
      );
    });

    it("should have correct gpt5-codex provider config", () => {
      expect(gpt5CodexProvider.name).toBe("gpt5-codex");
      expect(gpt5CodexProvider.displayName).toBe("OpenAI GPT-5.2 Codex");
      expect(gpt5CodexProvider.agentType).toBe("claude");
      expect(gpt5CodexProvider.environment.ANTHROPIC_DEFAULT_OPUS_MODEL).toBe(
        "gpt-5.2-codex(high)",
      );
    });
  });

  describe("opencode provider configurations", () => {
    it("should have correct gemini provider config", () => {
      expect(opencodeGeminiProvider.name).toBe("gemini");
      expect(opencodeGeminiProvider.displayName).toBe("Google Gemini");
      expect(opencodeGeminiProvider.agentType).toBe("opencode");
      expect(opencodeGeminiProvider.cliArgs).toEqual([
        "--model",
        "google/gemini-2.5-pro",
      ]);
    });
  });

  describe("getProvider", () => {
    it("should return claude provider by name", () => {
      const provider = getProvider("gemini", "claude");
      expect(provider).toBe(claudeGeminiProvider);
    });

    it("should return opencode provider by name", () => {
      const provider = getProvider("gemini", "opencode");
      expect(provider).toBe(opencodeGeminiProvider);
    });

    it("should return first matching provider without agent type", () => {
      const provider = getProvider("gemini");
      expect(provider?.name).toBe("gemini");
    });

    it("should return undefined for unknown provider", () => {
      expect(getProvider("unknown")).toBeUndefined();
    });
  });

  describe("getProviderNamesForAgent", () => {
    it("should return claude provider names", () => {
      const names = getProviderNamesForAgent("claude");
      expect(names).toContain("gemini");
      expect(names).toContain("glm");
    });

    it("should return opencode provider names", () => {
      const names = getProviderNamesForAgent("opencode");
      expect(names).toContain("gemini");
    });
  });

  describe("getProviderCliArgs", () => {
    it("should return empty array for claude provider", () => {
      const args = getProviderCliArgs(claudeGeminiProvider);
      expect(args).toEqual([]);
    });

    it("should return cli args for opencode provider", () => {
      const args = getProviderCliArgs(opencodeGeminiProvider);
      expect(args).toEqual(["--model", "google/gemini-2.5-pro"]);
    });
  });

  describe("providers map", () => {
    it("should contain all providers with agentType:name key", () => {
      expect(providers.get("claude:gemini")).toBe(claudeGeminiProvider);
      expect(providers.get("claude:glm")).toBe(glmProvider);
      expect(providers.get("opencode:gemini")).toBe(opencodeGeminiProvider);
    });
  });

  describe("buildProviderEnvironment", () => {
    const originalEnv = process.env;

    beforeEach(() => {
      process.env = { ...originalEnv };
    });

    afterEach(() => {
      process.env = originalEnv;
    });

    it("should build environment for claude gemini provider", () => {
      process.env.CLI_PROXY_API_KEY = "test-api-key";

      const env = buildProviderEnvironment(claudeGeminiProvider);

      expect(env.ANTHROPIC_BASE_URL).toBe("http://127.0.0.1:8317");
      expect(env.ANTHROPIC_AUTH_TOKEN).toBe("test-api-key");
      expect(env.API_TIMEOUT_MS).toBe("3000000");
      expect(env.ANTHROPIC_DEFAULT_OPUS_MODEL).toBe("gemini-3-pro-preview");
    });

    it("should build environment for glm provider", () => {
      process.env.ZAI_AUTH_TOKEN = "zai-token";

      const env = buildProviderEnvironment(glmProvider);

      expect(env.ANTHROPIC_BASE_URL).toBe("https://api.z.ai/api/anthropic");
      expect(env.ANTHROPIC_AUTH_TOKEN).toBe("zai-token");
    });

    it("should throw error when auth token is missing", () => {
      delete process.env.CLI_PROXY_API_KEY;

      expect(() => buildProviderEnvironment(claudeGeminiProvider)).toThrow(
        "Missing authentication token: CLI_PROXY_API_KEY environment variable is not set",
      );
    });

    it("should not throw for provider without authTokenEnv", () => {
      const env = buildProviderEnvironment(opencodeGeminiProvider);
      expect(env).toEqual({});
    });
  });
});
