import {describe, expect, it} from "vitest";
import {wrapWithInitCommands} from "../../../src/run/sandbox/command.js";

describe("command", () => {
  describe("wrapWithInitCommands", () => {
    it("should return command unchanged when no init commands", () => {
      const command = ["claude", "--verbose"];

      const result = wrapWithInitCommands(command);

      expect(result).toEqual(["claude", "--verbose"]);
    });

    it("should return command unchanged for empty init commands array", () => {
      const command = ["claude", "--verbose"];

      const result = wrapWithInitCommands(command, []);

      expect(result).toEqual(["claude", "--verbose"]);
    });

    it("should wrap with single init command", () => {
      const command = ["claude", "--verbose"];

      const result = wrapWithInitCommands(command, ["npm install"]);

      expect(result).toEqual([
        "sh",
        "-c",
        "npm install && exec claude --verbose",
      ]);
    });

    it("should chain multiple init commands", () => {
      const command = ["claude"];

      const result = wrapWithInitCommands(command, [
        "npm install",
        "npm run build",
      ]);

      expect(result).toEqual([
        "sh",
        "-c",
        "npm install && npm run build && exec claude",
      ]);
    });

    it("should quote arguments with special characters", () => {
      const command = ["claude", "--message", "hello world"];

      const result = wrapWithInitCommands(command, ["echo test"]);

      expect(result).toEqual([
        "sh",
        "-c",
        "echo test && exec claude --message 'hello world'",
      ]);
    });

    it("should handle arguments with single quotes", () => {
      const command = ["claude", "--message", "it's working"];

      const result = wrapWithInitCommands(command, ["npm install"]);

      expect(result).toEqual([
        "sh",
        "-c",
        "npm install && exec claude --message 'it'\"'\"'s working'",
      ]);
    });

    it("should handle paths without quoting", () => {
      const command = ["/env/bin/claude", "--work-dir", "/path/to/dir"];

      const result = wrapWithInitCommands(command, ["npm install"]);

      expect(result).toEqual([
        "sh",
        "-c",
        "npm install && exec /env/bin/claude --work-dir /path/to/dir",
      ]);
    });
  });
});
