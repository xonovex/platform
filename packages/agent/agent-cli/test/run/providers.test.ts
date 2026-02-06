import {afterEach, beforeEach, describe, expect, it} from "vitest";
import {geminiProvider as claudeGeminiProvider} from "../../src/run/providers/claude/gemini.js";
import {glmProvider} from "../../src/run/providers/claude/glm.js";
import {
  getProvider,
  getProviderNamesForAgent,
  providers,
} from "../../src/run/providers/index.js";
import {geminiProvider as opencodeGeminiProvider} from "../../src/run/providers/opencode/gemini.js";
import {
  buildProviderEnvironment,
  getProviderCliArgs,
} from "../../src/run/providers/types.js";

describe("providers", () => {
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
      process.env = {...originalEnv};
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
