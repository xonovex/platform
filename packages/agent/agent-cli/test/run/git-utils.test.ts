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
    it("should return string or null (null for detached HEAD)", () => {
      const branch = getCurrentBranchSync();

      // In CI with tag checkout, we may be in detached HEAD (returns null)
      // On a regular branch, it returns the branch name as string
      // Either way, the type should be string or null
      expect(branch === null || typeof branch === "string").toBe(true);
    });

    it("should return null for non-git directory", () => {
      const branch = getCurrentBranchSync("/tmp");

      expect(branch).toBeNull();
    });

    it("should return null for non-existent directory", () => {
      const branch = getCurrentBranchSync("/non/existent/path");

      expect(branch).toBeNull();
    });
  });

  describe("integration", () => {
    it("should return git root for a git repo", () => {
      const root = getGitRootSync();

      // Root should succeed in a git repo
      expect(root).not.toBeNull();
      // Root should be an absolute path
      expect(root).toMatch(/^\//);
    });
  });
});
