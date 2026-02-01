import {basename, resolve} from "node:path";
import {afterEach, beforeEach, describe, expect, it, vi} from "vitest";
import {
  buildBindPathsWithWorktree,
  checkExistingWorktree,
  getCurrentBranch,
  getDefaultWorktreeDir,
  isWorktreeDirectory,
  sanitizeBranchName,
  setupWorktree,
  type WorktreeConfig,
} from "../../src/run/worktree/index.js";

// Create mocks
const mockExecFile = vi.fn();
const mockExistsSync = vi.fn();
const mockStatSync = vi.fn();
const mockReadFileSync = vi.fn();

// Mock node:child_process
vi.mock("node:child_process", () => ({
  execFile: (...args: unknown[]): unknown => mockExecFile(...args),
}));

// Mock node:fs
vi.mock("node:fs", () => ({
  existsSync: (...args: unknown[]): unknown => mockExistsSync(...args),
  statSync: (...args: unknown[]): unknown => mockStatSync(...args),
  readFileSync: (...args: unknown[]): unknown => mockReadFileSync(...args),
}));

// Mock @xonovex/core logging to avoid console output in tests
vi.mock("@xonovex/core", () => ({
  logError: vi.fn(),
  logInfo: vi.fn(),
}));

interface MockExecFileCall {
  args: string[];
  cwd: string;
}

describe("worktree", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe("sanitizeBranchName", () => {
    it("should replace forward slashes with hyphens", () => {
      expect(sanitizeBranchName("feature/my-feature")).toBe(
        "feature-my-feature",
      );
      expect(sanitizeBranchName("feat/sub/deep")).toBe("feat-sub-deep");
    });

    it("should replace backslashes with hyphens", () => {
      expect(sanitizeBranchName(String.raw`feature\my-feature`)).toBe(
        "feature-my-feature",
      );
    });

    it("should replace special characters with hyphens", () => {
      expect(sanitizeBranchName("feature@test")).toBe("feature-test");
      expect(sanitizeBranchName("feature:test")).toBe("feature-test");
      expect(sanitizeBranchName("feature.test")).toBe("feature-test");
    });

    it("should collapse multiple hyphens", () => {
      expect(sanitizeBranchName("feature//test")).toBe("feature-test");
      expect(sanitizeBranchName("a---b")).toBe("a-b");
    });

    it("should remove leading and trailing hyphens", () => {
      expect(sanitizeBranchName("/feature")).toBe("feature");
      expect(sanitizeBranchName("feature/")).toBe("feature");
      expect(sanitizeBranchName("/feature/")).toBe("feature");
    });

    it("should preserve underscores and alphanumerics", () => {
      expect(sanitizeBranchName("my_feature_123")).toBe("my_feature_123");
    });
  });

  describe("getDefaultWorktreeDir", () => {
    it("should return parent directory with repo name and sanitized branch name", () => {
      expect(getDefaultWorktreeDir("feature/my-feature", "myrepo")).toBe(
        "../myrepo-feature-my-feature",
      );
      expect(getDefaultWorktreeDir("main", "project")).toBe("../project-main");
    });

    it("should handle complex branch names", () => {
      expect(getDefaultWorktreeDir("feat/user/JIRA-123", "myrepo")).toBe(
        "../myrepo-feat-user-JIRA-123",
      );
    });

    it("should sanitize repo name as well", () => {
      expect(getDefaultWorktreeDir("feature/test", "my.repo")).toBe(
        "../my-repo-feature-test",
      );
      expect(getDefaultWorktreeDir("main", "repo/name")).toBe(
        "../repo-name-main",
      );
    });
  });

  describe("WorktreeConfig", () => {
    it("should accept config with all fields", () => {
      const config: WorktreeConfig = {
        sourceBranch: "main",
        branch: "feature/test",
        dir: "/path/to/worktree",
      };

      expect(config.sourceBranch).toBe("main");
      expect(config.branch).toBe("feature/test");
      expect(config.dir).toBe("/path/to/worktree");
    });

    it("should accept config without sourceBranch", () => {
      const config: WorktreeConfig = {
        branch: "feature/test",
        dir: "/path/to/worktree",
      };

      expect(config.sourceBranch).toBeUndefined();
      expect(config.branch).toBe("feature/test");
      expect(config.dir).toBe("/path/to/worktree");
    });
  });

  describe("getCurrentBranch", () => {
    it("should return current branch name", async () => {
      mockExecFile.mockImplementation(
        (
          _cmd: string,
          _args: string[],
          _options: {cwd: string},
          callback: (
            err: Error | null,
            result: {stdout: string; stderr: string},
          ) => void,
        ) => {
          callback(null, {stdout: "main\n", stderr: ""});
          return {};
        },
      );

      const branch = await getCurrentBranch("/test/repo");

      expect(mockExecFile).toHaveBeenCalledWith(
        "git",
        ["rev-parse", "--abbrev-ref", "HEAD"],
        {cwd: "/test/repo"},
        expect.any(Function),
      );
      expect(branch).toBe("main");
    });

    it("should trim whitespace from branch name", async () => {
      mockExecFile.mockImplementation(
        (
          _cmd: string,
          _args: string[],
          _options: {cwd: string},
          callback: (
            err: Error | null,
            result: {stdout: string; stderr: string},
          ) => void,
        ) => {
          callback(null, {stdout: "  feature/branch  \n", stderr: ""});
          return {};
        },
      );

      const branch = await getCurrentBranch("/test/repo");
      expect(branch).toBe("feature/branch");
    });
  });

  describe("setupWorktree", () => {
    it("should create worktree with explicit source branch (new branch)", async () => {
      const calls: MockExecFileCall[] = [];

      mockExecFile.mockImplementation(
        (
          _cmd: string,
          args: string[],
          options: {cwd: string},
          callback: (
            err: Error | null,
            result: {stdout: string; stderr: string},
          ) => void,
        ) => {
          calls.push({args: [...args], cwd: options.cwd});
          // branchExists check should fail (branch doesn't exist)
          if (args[0] === "rev-parse" && args[1] === "--verify") {
            callback(new Error("branch not found"), {stdout: "", stderr: ""});
          } else {
            callback(null, {stdout: "", stderr: ""});
          }
          return {};
        },
      );

      const config: WorktreeConfig = {
        sourceBranch: "develop",
        branch: "feature/test",
        dir: "../my-worktree",
      };

      const result = await setupWorktree(config, "/home/user/repo");

      // First call should check if branch exists
      expect(calls[0]).toEqual({
        args: ["rev-parse", "--verify", "refs/heads/feature/test"],
        cwd: "/home/user/repo",
      });

      // Second call should create worktree with new branch
      expect(calls[1]).toEqual({
        args: [
          "worktree",
          "add",
          "../my-worktree",
          "-b",
          "feature/test",
          "develop",
        ],
        cwd: "/home/user/repo",
      });

      // Third call should set mergeBackTo config
      expect(calls[2]).toEqual({
        args: ["config", "branch.feature/test.mergeBackTo", "develop"],
        cwd: "/home/user/repo",
      });

      // Should return resolved worktree path
      expect(result).toBe(resolve("/home/user/repo", "../my-worktree"));
    });

    it("should use current branch when sourceBranch not specified", async () => {
      const calls: MockExecFileCall[] = [];

      mockExecFile.mockImplementation(
        (
          _cmd: string,
          args: string[],
          options: {cwd: string},
          callback: (
            err: Error | null,
            result: {stdout: string; stderr: string},
          ) => void,
        ) => {
          calls.push({args: [...args], cwd: options.cwd});

          // branchExists check should fail (branch doesn't exist)
          if (args[0] === "rev-parse" && args[1] === "--verify") {
            callback(new Error("branch not found"), {stdout: "", stderr: ""});
          } else if (args[0] === "rev-parse" && args[1] === "--abbrev-ref") {
            // getCurrentBranch
            callback(null, {stdout: "main\n", stderr: ""});
          } else {
            callback(null, {stdout: "", stderr: ""});
          }
          return {};
        },
      );

      const config: WorktreeConfig = {
        branch: "feature/new",
        dir: "./worktree-dir",
      };

      await setupWorktree(config, "/project");

      // First call should check if branch exists
      expect(calls[0]).toEqual({
        args: ["rev-parse", "--verify", "refs/heads/feature/new"],
        cwd: "/project",
      });

      // Second call should get current branch
      expect(calls[1]).toEqual({
        args: ["rev-parse", "--abbrev-ref", "HEAD"],
        cwd: "/project",
      });

      // Third call should create worktree with detected branch
      expect(calls[2]).toEqual({
        args: [
          "worktree",
          "add",
          "./worktree-dir",
          "-b",
          "feature/new",
          "main",
        ],
        cwd: "/project",
      });

      // Fourth call should set mergeBackTo config
      expect(calls[3]).toEqual({
        args: ["config", "branch.feature/new.mergeBackTo", "main"],
        cwd: "/project",
      });
    });

    it("should throw error when worktree creation fails", async () => {
      mockExecFile.mockImplementation(
        (
          _cmd: string,
          _args: string[],
          _options: {cwd: string},
          callback: (
            err: Error | null,
            result: {stdout: string; stderr: string},
          ) => void,
        ) => {
          callback(new Error("fatal: branch already exists"), {
            stdout: "",
            stderr: "",
          });
          return {};
        },
      );

      const config: WorktreeConfig = {
        sourceBranch: "main",
        branch: "existing-branch",
        dir: "./worktree",
      };

      await expect(setupWorktree(config, "/project")).rejects.toThrow(
        "fatal: branch already exists",
      );
    });

    it("should log info when verbose is enabled", async () => {
      const {logInfo} = await import("@xonovex/core");
      const mockLogInfo = vi.mocked(logInfo);

      mockExecFile.mockImplementation(
        (
          _cmd: string,
          _args: string[],
          _options: {cwd: string},
          callback: (
            err: Error | null,
            result: {stdout: string; stderr: string},
          ) => void,
        ) => {
          callback(null, {stdout: "", stderr: ""});
          return {};
        },
      );

      const config: WorktreeConfig = {
        sourceBranch: "main",
        branch: "feature/verbose-test",
        dir: "./verbose-worktree",
      };

      await setupWorktree(config, "/project", true);

      expect(mockLogInfo).toHaveBeenCalledWith(
        expect.stringContaining("Creating worktree"),
      );
      expect(mockLogInfo).toHaveBeenCalledWith(
        expect.stringContaining("Worktree created successfully"),
      );
    });

    it("should not log when verbose is disabled", async () => {
      const {logInfo} = await import("@xonovex/core");
      const mockLogInfo = vi.mocked(logInfo);

      mockExecFile.mockImplementation(
        (
          _cmd: string,
          _args: string[],
          _options: {cwd: string},
          callback: (
            err: Error | null,
            result: {stdout: string; stderr: string},
          ) => void,
        ) => {
          callback(null, {stdout: "", stderr: ""});
          return {};
        },
      );

      const config: WorktreeConfig = {
        sourceBranch: "main",
        branch: "feature/quiet-test",
        dir: "./quiet-worktree",
      };

      await setupWorktree(config, "/project", false);

      expect(mockLogInfo).not.toHaveBeenCalled();
    });
  });

  describe("worktree sandbox binding", () => {
    it("should include worktree source dir in bind paths when worktree is used", () => {
      const baseBindPaths = ["/custom/path"];
      const worktreeSourceDir = "/home/user/myrepo";

      const bindPaths = buildBindPathsWithWorktree(
        baseBindPaths,
        worktreeSourceDir,
      );

      expect(bindPaths).toEqual(["/custom/path", "/home/user/myrepo"]);
    });

    it("should not modify bind paths when worktree is not used", () => {
      const baseBindPaths = ["/custom/path"];
      const worktreeSourceDir = undefined;

      const bindPaths = buildBindPathsWithWorktree(
        baseBindPaths,
        worktreeSourceDir,
      );

      expect(bindPaths).toEqual(["/custom/path"]);
    });

    it("should handle empty base bind paths with worktree", () => {
      const baseBindPaths: string[] = [];
      const worktreeSourceDir = "/home/user/myrepo";

      const bindPaths = buildBindPathsWithWorktree(
        baseBindPaths,
        worktreeSourceDir,
      );

      expect(bindPaths).toEqual(["/home/user/myrepo"]);
    });

    it("should preserve order with worktree source at end", () => {
      const baseBindPaths = ["/path/a", "/path/b"];
      const worktreeSourceDir = "/home/user/myrepo";

      const bindPaths = buildBindPathsWithWorktree(
        baseBindPaths,
        worktreeSourceDir,
      );

      expect(bindPaths).toEqual(["/path/a", "/path/b", "/home/user/myrepo"]);
    });
  });

  describe("worktree directory naming", () => {
    it("should derive repo name from workDir basename", () => {
      const workDir = "/home/user/my-project";
      const repoName = basename(workDir);

      expect(repoName).toBe("my-project");
      expect(getDefaultWorktreeDir("feature/test", repoName)).toBe(
        "../my-project-feature-test",
      );
    });

    it("should handle nested workDir paths", () => {
      const workDir = "/home/user/projects/platform/packages/myrepo";
      const repoName = basename(workDir);

      expect(repoName).toBe("myrepo");
      expect(getDefaultWorktreeDir("feat/JIRA-123", repoName)).toBe(
        "../myrepo-feat-JIRA-123",
      );
    });
  });

  describe("isWorktreeDirectory", () => {
    it("should return false if .git does not exist", () => {
      mockExistsSync.mockReturnValue(false);

      expect(isWorktreeDirectory("/some/dir")).toBe(false);
    });

    it("should return false if .git is a directory (main repo)", () => {
      mockExistsSync.mockReturnValue(true);
      mockStatSync.mockReturnValue({isFile: () => false});

      expect(isWorktreeDirectory("/some/dir")).toBe(false);
    });

    it("should return false if .git file does not contain gitdir", () => {
      mockExistsSync.mockReturnValue(true);
      mockStatSync.mockReturnValue({isFile: () => true});
      mockReadFileSync.mockReturnValue("something else");

      expect(isWorktreeDirectory("/some/dir")).toBe(false);
    });

    it("should return true if .git file contains gitdir pointer", () => {
      mockExistsSync.mockReturnValue(true);
      mockStatSync.mockReturnValue({isFile: () => true});
      mockReadFileSync.mockReturnValue("gitdir: /repo/.git/worktrees/feature");

      expect(isWorktreeDirectory("/some/dir")).toBe(true);
    });
  });

  describe("checkExistingWorktree", () => {
    it("should return exists=false if directory does not exist", async () => {
      mockExistsSync.mockReturnValue(false);

      const result = await checkExistingWorktree("./worktree", "/repo");

      expect(result).toEqual({
        exists: false,
        isWorktree: false,
        isForThisRepo: false,
      });
    });

    it("should return isWorktree=false if directory exists but is not a worktree", async () => {
      mockExistsSync.mockImplementation((path: string) => {
        // Directory exists but .git doesn't
        return !path.endsWith(".git");
      });

      const result = await checkExistingWorktree("./worktree", "/repo");

      expect(result.exists).toBe(true);
      expect(result.isWorktree).toBe(false);
    });
  });

  describe("setupWorktree reuse behavior", () => {
    it("should reuse existing worktree with matching branch", async () => {
      // Setup: worktree exists with correct branch
      mockExistsSync.mockReturnValue(true);
      mockStatSync.mockReturnValue({isFile: () => true});
      mockReadFileSync.mockReturnValue(
        "gitdir: /repo/.git/worktrees/feature-test",
      );

      mockExecFile.mockImplementation(
        (
          _cmd: string,
          args: string[],
          options: {cwd: string},
          callback: (
            err: Error | null,
            result: {stdout: string; stderr: string},
          ) => void,
        ) => {
          // Handle different git commands based on cwd
          if (args[0] === "rev-parse" && args[1] === "--abbrev-ref") {
            callback(null, {stdout: "feature/test\n", stderr: ""});
          } else if (args[0] === "rev-parse" && args[1] === "--git-dir") {
            // Return different paths based on which directory is being queried
            if (options.cwd.includes("worktree")) {
              // Worktree's git dir is inside repo's .git/worktrees/
              callback(null, {
                stdout: "/repo/.git/worktrees/feature-test",
                stderr: "",
              });
            } else {
              // Main repo's git dir
              callback(null, {stdout: "/repo/.git", stderr: ""});
            }
          } else {
            callback(null, {stdout: "", stderr: ""});
          }
          return {};
        },
      );

      const config: WorktreeConfig = {
        branch: "feature/test",
        dir: "./worktree",
      };

      const result = await setupWorktree(config, "/repo");

      // Should return the path without creating new worktree
      expect(result).toBe(resolve("/repo", "./worktree"));
      // Should not have called worktree add
      const calls = mockExecFile.mock.calls;
      const worktreeAddCalls = calls.filter(
        (call) => Array.isArray(call[1]) && call[1][0] === "worktree",
      );
      expect(worktreeAddCalls).toHaveLength(0);
    });

    it("should throw error if directory exists but is not a worktree", async () => {
      // Setup: directory exists but no .git file
      mockExistsSync.mockImplementation(
        (path: string) => !path.endsWith(".git"),
      );

      const config: WorktreeConfig = {
        branch: "feature/test",
        dir: "./existing-dir",
      };

      await expect(setupWorktree(config, "/repo")).rejects.toThrow(
        "Directory exists but is not a worktree",
      );
    });

    it("should throw error if worktree exists with different branch", async () => {
      // Setup: worktree exists but on different branch
      mockExistsSync.mockReturnValue(true);
      mockStatSync.mockReturnValue({isFile: () => true});
      mockReadFileSync.mockReturnValue("gitdir: /repo/.git/worktrees/other");

      mockExecFile.mockImplementation(
        (
          _cmd: string,
          args: string[],
          options: {cwd: string},
          callback: (
            err: Error | null,
            result: {stdout: string; stderr: string},
          ) => void,
        ) => {
          if (args[0] === "rev-parse" && args[1] === "--abbrev-ref") {
            callback(null, {stdout: "other-branch\n", stderr: ""});
          } else if (args[0] === "rev-parse" && args[1] === "--git-dir") {
            // Return different paths based on which directory is being queried
            if (options.cwd.includes("worktree")) {
              callback(null, {
                stdout: "/repo/.git/worktrees/other",
                stderr: "",
              });
            } else {
              callback(null, {stdout: "/repo/.git", stderr: ""});
            }
          } else {
            callback(null, {stdout: "", stderr: ""});
          }
          return {};
        },
      );

      const config: WorktreeConfig = {
        branch: "feature/test",
        dir: "./worktree",
      };

      await expect(setupWorktree(config, "/repo")).rejects.toThrow(
        "Worktree on wrong branch",
      );
    });

    it("should create new worktree with new branch if directory and branch do not exist", async () => {
      const calls: MockExecFileCall[] = [];

      // Setup: directory doesn't exist
      mockExistsSync.mockReturnValue(false);

      mockExecFile.mockImplementation(
        (
          _cmd: string,
          args: string[],
          options: {cwd: string},
          callback: (
            err: Error | null,
            result: {stdout: string; stderr: string},
          ) => void,
        ) => {
          calls.push({args: [...args], cwd: options.cwd});
          // branchExists check should fail (branch doesn't exist)
          if (args[0] === "rev-parse" && args[1] === "--verify") {
            callback(new Error("branch not found"), {stdout: "", stderr: ""});
          } else if (args[0] === "rev-parse" && args[1] === "--abbrev-ref") {
            callback(null, {stdout: "main\n", stderr: ""});
          } else {
            callback(null, {stdout: "", stderr: ""});
          }
          return {};
        },
      );

      const config: WorktreeConfig = {
        branch: "feature/new",
        dir: "./new-worktree",
      };

      await setupWorktree(config, "/repo");

      // Should have called worktree add with -b flag for new branch
      const worktreeAddCall = calls.find((c) => c.args[0] === "worktree");
      expect(worktreeAddCall).toBeDefined();
      expect(worktreeAddCall?.args).toEqual([
        "worktree",
        "add",
        "./new-worktree",
        "-b",
        "feature/new",
        "main",
      ]);
    });

    it("should create worktree for existing branch without -b flag", async () => {
      const calls: MockExecFileCall[] = [];

      // Setup: directory doesn't exist
      mockExistsSync.mockReturnValue(false);

      mockExecFile.mockImplementation(
        (
          _cmd: string,
          args: string[],
          options: {cwd: string},
          callback: (
            err: Error | null,
            result: {stdout: string; stderr: string},
          ) => void,
        ) => {
          calls.push({args: [...args], cwd: options.cwd});
          // branchExists check should succeed (branch exists)
          if (args[0] === "rev-parse" && args[1] === "--verify") {
            callback(null, {stdout: "abc123", stderr: ""});
          } else {
            callback(null, {stdout: "", stderr: ""});
          }
          return {};
        },
      );

      const config: WorktreeConfig = {
        branch: "feature/existing",
        dir: "./existing-branch-worktree",
      };

      await setupWorktree(config, "/repo");

      // Should have called worktree add WITHOUT -b flag
      const worktreeAddCall = calls.find((c) => c.args[0] === "worktree");
      expect(worktreeAddCall).toBeDefined();
      expect(worktreeAddCall?.args).toEqual([
        "worktree",
        "add",
        "./existing-branch-worktree",
        "feature/existing",
      ]);

      // Should NOT have called config (no mergeBackTo for existing branches)
      const configCall = calls.find((c) => c.args[0] === "config");
      expect(configCall).toBeUndefined();
    });
  });
});
