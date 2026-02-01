import {describe, expect, it} from "vitest";
import {claudeAgent} from "../../src/run/agents/claude/index.js";
import {agents, getAgent, getAgentTypes} from "../../src/run/agents/index.js";
import {opencodeAgent} from "../../src/run/agents/opencode/index.js";
import {type AgentExecOptions} from "../../src/run/agents/types.js";

describe("agents", () => {
  describe("claudeAgent", () => {
    it("should have correct configuration", () => {
      expect(claudeAgent.type).toBe("claude");
      expect(claudeAgent.displayName).toBe("Claude Code");
      expect(claudeAgent.binary).toBe("claude");
    });

    it("should have nixPackage for nix sandbox support", () => {
      expect(claudeAgent.nixPackage).toBe("claude-code");
    });

    it("should build args without sandbox", () => {
      const options: AgentExecOptions = {
        sandbox: false,
        providerCliArgs: [],
      };
      const args = claudeAgent.buildArgs(["--help"], options);
      expect(args).toEqual(["--help"]);
    });

    it("should build args with sandbox (adds permission bypass)", () => {
      const options: AgentExecOptions = {
        sandbox: true,
        providerCliArgs: [],
      };
      const args = claudeAgent.buildArgs(["--help"], options);
      expect(args).toEqual([
        "--permission-mode",
        "bypassPermissions",
        "--help",
      ]);
    });

    it("should ignore providerCliArgs (uses env instead)", () => {
      const options: AgentExecOptions = {
        sandbox: false,
        providerCliArgs: ["--model", "test"],
      };
      const args = claudeAgent.buildArgs([], options);
      expect(args).toEqual([]);
    });

    it("should pass through provider env", () => {
      const providerEnv = {
        ANTHROPIC_BASE_URL: "https://test.example.com",
        ANTHROPIC_AUTH_TOKEN: "token",
      };
      const env = claudeAgent.buildEnv(providerEnv);
      expect(env).toEqual(providerEnv);
    });
  });

  describe("opencodeAgent", () => {
    it("should have correct configuration", () => {
      expect(opencodeAgent.type).toBe("opencode");
      expect(opencodeAgent.displayName).toBe("OpenCode");
      expect(opencodeAgent.binary).toBe("opencode");
    });

    it("should have nixPackage for nix sandbox support", () => {
      expect(opencodeAgent.nixPackage).toBe("opencode");
    });

    it("should build args with providerCliArgs", () => {
      const options: AgentExecOptions = {
        sandbox: false,
        providerCliArgs: ["--model", "google/gemini-2.5-pro"],
      };
      const args = opencodeAgent.buildArgs(["project"], options);
      expect(args).toEqual(["--model", "google/gemini-2.5-pro", "project"]);
    });

    it("should not add permission mode in sandbox (not supported)", () => {
      const options: AgentExecOptions = {
        sandbox: true,
        providerCliArgs: [],
      };
      const args = opencodeAgent.buildArgs([], options);
      expect(args).toEqual([]);
    });

    it("should return empty env (ignores provider env)", () => {
      const providerEnv = {
        SOME_VAR: "value",
      };
      const env = opencodeAgent.buildEnv(providerEnv);
      expect(env).toEqual({});
    });
  });

  describe("getAgent", () => {
    it("should return claude agent", () => {
      expect(getAgent("claude")).toBe(claudeAgent);
    });

    it("should return opencode agent", () => {
      expect(getAgent("opencode")).toBe(opencodeAgent);
    });

    it("should return undefined for unknown agent", () => {
      // @ts-expect-error Testing unknown agent
      expect(getAgent("unknown")).toBeUndefined();
    });
  });

  describe("getAgentTypes", () => {
    it("should return all agent types", () => {
      const types = getAgentTypes();
      expect(types).toContain("claude");
      expect(types).toContain("opencode");
      expect(types).toHaveLength(2);
    });
  });

  describe("agents map", () => {
    it("should contain all agents", () => {
      expect(agents.size).toBe(2);
      expect(agents.get("claude")).toBe(claudeAgent);
      expect(agents.get("opencode")).toBe(opencodeAgent);
    });
  });
});
