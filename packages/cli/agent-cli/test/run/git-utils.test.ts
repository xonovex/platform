import {describe, expect, it} from "vitest";
import {
  getCurrentBranchSync,
  getGitRootSync,
} from "../../src/run/worktree/index.js";

describe("git utilities", () => {
  describe("getGitRootSync", () => {
    it("should return git root for current directory", () => {
      const root = getGitRootSync();

      // We're in a git repo, so it should return something
      expect(root).not.toBeNull();
      expect(root).toContain("platform");
    });

    it("should return git root for specified directory", () => {
      const root = getGitRootSync(process.cwd());

      expect(root).not.toBeNull();
      expect(root).toContain("platform");
    });

    it("should return null for non-git directory", () => {
      const root = getGitRootSync("/tmp");

      expect(root).toBeNull();
    });

    it("should return null for non-existent directory", () => {
      const root = getGitRootSync("/non/existent/path");

      expect(root).toBeNull();
    });
  });

  describe("getCurrentBranchSync", () => {
    it("should return current branch for current directory", () => {
      const branch = getCurrentBranchSync();

      // We're in a git repo, so it should return something
      expect(branch).not.toBeNull();
      expect(typeof branch).toBe("string");
    });

    it("should return current branch for specified directory", () => {
      const branch = getCurrentBranchSync(process.cwd());

      expect(branch).not.toBeNull();
      expect(typeof branch).toBe("string");
    });

    it("should return null for non-git directory", () => {
      const branch = getCurrentBranchSync("/tmp");

      expect(branch).toBeNull();
    });

    it("should return null for non-existent directory", () => {
      const branch = getCurrentBranchSync("/non/existent/path");

      expect(branch).toBeNull();
    });

    it("should return a valid branch name format", () => {
      const branch = getCurrentBranchSync();

      // We're in a git repo so branch should exist
      expect(branch).not.toBeNull();
      // Branch names shouldn't contain spaces or control characters
      expect(branch).not.toMatch(/\s/);
      // Branch name should not be "HEAD" (that would indicate detached)
      // Our function returns null for detached HEAD
      expect(branch).not.toBe("HEAD");
    });
  });

  describe("integration", () => {
    it("should work together to identify repo context", () => {
      const root = getGitRootSync();
      const branch = getCurrentBranchSync();

      // Both should succeed in a git repo
      expect(root).not.toBeNull();
      expect(branch).not.toBeNull();

      // Root should be an absolute path
      expect(root).toMatch(/^\//);
    });
  });
});
