import { describe, expect, it } from "vitest";
import { combineAgentArgs } from "../../src/run/args.js";

describe("args", () => {
  describe("combineAgentArgs", () => {
    it("should combine positional and unknown args", () => {
      const result = combineAgentArgs(["send to agent"], ["--passed-to-agent"]);

      expect(result).toEqual(["send to agent", "--passed-to-agent"]);
    });

    it("should deduplicate args that appear in both", () => {
      const result = combineAgentArgs(
        ["/git-commit"],
        ["/git-commit", "--verbose"],
      );

      expect(result).toEqual(["/git-commit", "--verbose"]);
    });

    it("should handle empty positional args", () => {
      const result = combineAgentArgs([], ["--flag", "value"]);

      expect(result).toEqual(["--flag", "value"]);
    });

    it("should handle empty unknown args", () => {
      const result = combineAgentArgs(["say hello"], []);

      expect(result).toEqual(["say hello"]);
    });

    it("should handle both empty", () => {
      const result = combineAgentArgs([], []);

      expect(result).toEqual([]);
    });

    it("should preserve order with positional first", () => {
      const result = combineAgentArgs(["first", "second"], ["third", "fourth"]);

      expect(result).toEqual(["first", "second", "third", "fourth"]);
    });

    it("should handle typical slash command usage", () => {
      // agent-cli -- '/git-commit'
      const result = combineAgentArgs(["/git-commit"], []);

      expect(result).toEqual(["/git-commit"]);
    });

    it("should handle mixed args with flags after --", () => {
      // agent-cli 'say hello' -- --permission-mode bypass
      const result = combineAgentArgs(
        ["say hello"],
        ["--permission-mode", "bypass"],
      );

      expect(result).toEqual(["say hello", "--permission-mode", "bypass"]);
    });
  });
});
