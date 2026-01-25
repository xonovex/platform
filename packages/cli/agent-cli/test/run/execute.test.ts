import {describe, expect, it} from "vitest";
import {claudeAgent} from "../../src/run/agents/claude/index.js";
import {parseCustomEnv} from "../../src/run/sandbox/environment.js";
import {getExecutor} from "../../src/run/sandbox/index.js";
import type {SandboxConfig} from "../../src/run/sandbox/types.js";

describe("executeAgent customEnv", () => {
  describe("parseCustomEnv integration", () => {
    it("parseCustomEnv should be used for building environment", () => {
      // Test that parseCustomEnv correctly parses custom env vars
      // This validates the integration point used by executeAgent
      const customEnv = ["MY_VAR=value", "ANOTHER=test"];
      const parsed = parseCustomEnv(customEnv);

      expect(parsed).toEqual({
        MY_VAR: "value",
        ANOTHER: "test",
      });
    });

    it("customEnv should override existing values when merged", () => {
      const baseEnv: Record<string, string> = {
        PATH: "/usr/bin",
        HOME: "/home/user",
      };
      const customEnv = ["PATH=/custom/bin", "NEW_VAR=new"];
      const parsed = parseCustomEnv(customEnv);

      const merged = {...baseEnv, ...parsed};

      expect(merged.PATH).toBe("/custom/bin");
      expect(merged.HOME).toBe("/home/user");
      expect(merged.NEW_VAR).toBe("new");
    });

    it("empty customEnv should not affect environment", () => {
      const baseEnv = {
        PATH: "/usr/bin",
        HOME: "/home/user",
      };
      const parsed = parseCustomEnv([]);

      const merged = {...baseEnv, ...parsed};

      expect(merged).toEqual(baseEnv);
    });

    it("customEnv with values containing equals signs should parse correctly", () => {
      const customEnv = ["URL=https://example.com?foo=bar&baz=qux"];
      const parsed = parseCustomEnv(customEnv);

      expect(parsed.URL).toBe("https://example.com?foo=bar&baz=qux");
    });

    it("customEnv should support empty values", () => {
      const customEnv = ["EMPTY_VAR="];
      const parsed = parseCustomEnv(customEnv);

      expect(parsed.EMPTY_VAR).toBe("");
    });
  });

  describe("noneExecutor customEnv config", () => {
    it("should accept config with customEnv", () => {
      const executor = getExecutor("none");

      const config: SandboxConfig = {
        agentId: "test-agent",
        method: "none",
        agent: claudeAgent,
        workDir: "/home/user/project",
        network: true,
        bindPaths: [],
        roBindPaths: [],
        customEnv: ["MY_VAR=test", "ANOTHER=value"],
        agentArgs: [],
        verbose: false,
        debug: false,
        dryRun: false,
      };

      // getCommand doesn't include env vars (they're applied at execute time)
      // This test verifies the config structure is accepted
      const command = executor.getCommand(config);
      expect(command).toContain("claude");
    });

    it("should work with empty customEnv", () => {
      const executor = getExecutor("none");

      const config: SandboxConfig = {
        agentId: "test-agent",
        method: "none",
        agent: claudeAgent,
        workDir: "/home/user/project",
        network: true,
        bindPaths: [],
        roBindPaths: [],
        customEnv: [],
        agentArgs: ["--help"],
        verbose: false,
        debug: false,
        dryRun: false,
      };

      const command = executor.getCommand(config);
      expect(command).toEqual(["claude", "--help"]);
    });

    it("customEnv should not appear in getCommand output", () => {
      const executor = getExecutor("none");

      const config: SandboxConfig = {
        agentId: "test-agent",
        method: "none",
        agent: claudeAgent,
        workDir: "/home/user/project",
        network: true,
        bindPaths: [],
        roBindPaths: [],
        customEnv: ["SECRET=value"],
        agentArgs: [],
        verbose: false,
        debug: false,
        dryRun: false,
      };

      // Environment variables are applied at execute() time, not in getCommand()
      const command = executor.getCommand(config);
      expect(command.join(" ")).not.toContain("SECRET");
      expect(command.join(" ")).not.toContain("value");
    });
  });
});
