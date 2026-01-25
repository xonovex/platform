import { execSync } from "node:child_process";
import { beforeEach, describe, expect, it, vi } from "vitest";
import {
  executeInitCommand,
  executeInitCommands,
} from "../../src/run/init/index.js";

vi.mock("node:child_process", () => ({
  execSync: vi.fn(),
}));

vi.mock("@xonovex/tool-lib", () => ({
  logInfo: vi.fn(),
  logError: vi.fn(),
}));

describe("init", () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  describe("executeInitCommand", () => {
    it("should execute command with correct options", () => {
      executeInitCommand("npm install", "/work/dir");

      expect(execSync).toHaveBeenCalledWith("npm install", {
        cwd: "/work/dir",
        stdio: "pipe",
        shell: "/bin/sh",
      });
    });

    it("should use inherit stdio when verbose", () => {
      executeInitCommand("npm install", "/work/dir", true);

      expect(execSync).toHaveBeenCalledWith("npm install", {
        cwd: "/work/dir",
        stdio: "inherit",
        shell: "/bin/sh",
      });
    });

    it("should throw when command fails", () => {
      vi.mocked(execSync).mockImplementationOnce(() => {
        throw new Error("command failed");
      });

      expect(() => {
        executeInitCommand("bad-command", "/work/dir");
      }).toThrow();
    });
  });

  describe("executeInitCommands", () => {
    it("should execute all commands in sequence", () => {
      executeInitCommands(["npm install", "npm run build"], "/work/dir");

      expect(execSync).toHaveBeenCalledTimes(2);
      expect(execSync).toHaveBeenNthCalledWith(
        1,
        "npm install",
        expect.any(Object),
      );
      expect(execSync).toHaveBeenNthCalledWith(
        2,
        "npm run build",
        expect.any(Object),
      );
    });

    it("should stop on first failure", () => {
      vi.mocked(execSync)
        .mockReturnValueOnce(Buffer.from(""))
        .mockImplementationOnce(() => {
          throw new Error("second command failed");
        });

      expect(() => {
        executeInitCommands(["first", "second", "third"], "/work/dir");
      }).toThrow();

      expect(execSync).toHaveBeenCalledTimes(2);
    });

    it("should handle empty array", () => {
      executeInitCommands([], "/work/dir");

      expect(execSync).not.toHaveBeenCalled();
    });

    it("should pass verbose flag to all commands", () => {
      executeInitCommands(["cmd1", "cmd2"], "/work/dir", true);

      expect(execSync).toHaveBeenNthCalledWith(1, "cmd1", {
        cwd: "/work/dir",
        stdio: "inherit",
        shell: "/bin/sh",
      });
      expect(execSync).toHaveBeenNthCalledWith(2, "cmd2", {
        cwd: "/work/dir",
        stdio: "inherit",
        shell: "/bin/sh",
      });
    });
  });
});
