import { describe, expect, it } from "vitest";
import { generateBashCompletion } from "../src/completion/bash/generator.js";

describe("completion", () => {
  describe("generateBashCompletion", () => {
    it("should generate valid bash completion script", () => {
      const script = generateBashCompletion();

      expect(script).toContain("_agent_cli_completions");
      expect(script).toContain("complete -F _agent_cli_completions agent-cli");
      expect(script).toContain("gemini");
      expect(script).toContain("gemini-claude");
      expect(script).toContain("glm");
      expect(script).toContain("gpt5-codex");
      expect(script).toContain("run");
      expect(script).toContain("completion");
    });

    it("should include sandbox options", () => {
      const script = generateBashCompletion();

      expect(script).toContain("--home-dir");
      expect(script).toContain("--work-dir");
      expect(script).toContain("--provider");
      expect(script).toContain("--sandbox");
    });

    it("should include provider and sandbox lists", () => {
      const script = generateBashCompletion();

      expect(script).toContain(
        'providers="gemini gemini-claude glm gpt5-codex"',
      );
      expect(script).toContain('sandboxes="none bwrap docker"');
    });
  });
});
