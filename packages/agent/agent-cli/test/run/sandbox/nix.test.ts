import {describe, expect, it, vi} from "vitest";
// Import after mocks
import {nixExecutor} from "../../../src/run/sandbox/nix/index.js";
import type {SandboxConfig} from "../../../src/run/sandbox/types.js";

// We need to test the internal functions, so we'll test via the exported executor behavior
// by checking the getCommand output which uses parseNixConfig internally

vi.mock("node:child_process", () => ({
  spawn: vi.fn(),
  spawnSync: vi.fn(() => ({status: 0})),
}));

vi.mock("node:fs", () => ({
  existsSync: vi.fn(() => true),
  mkdirSync: vi.fn(),
  realpathSync: vi.fn((p: string) => p),
}));

vi.mock("@xonovex/core", () => ({
  logInfo: vi.fn(),
  logError: vi.fn(),
  logDebug: vi.fn(),
  logWarning: vi.fn(),
}));

function createConfig(overrides: Partial<SandboxConfig> = {}): SandboxConfig {
  return {
    agentId: "test-agent-123",
    method: "nix",
    workDir: "/test/dir",
    network: true,
    bindPaths: [],
    roBindPaths: [],
    customEnv: [],
    agentArgs: [],
    verbose: false,
    debug: false,
    dryRun: false,
    ...overrides,
  };
}

describe("nix sandbox", () => {
  describe("package sets", () => {
    it("should expand nodejs set with build tools", () => {
      const config = createConfig({image: "nix:sets=nodejs"});
      const command = nixExecutor.getCommand(config);
      const packagesLine = command.find((c) => c.includes("Packages:"));

      expect(packagesLine).toContain("nodejs_24");
      expect(packagesLine).toContain("gnumake");
      expect(packagesLine).toContain("gcc");
      expect(packagesLine).toContain("gnused");
    });

    it("should expand python set", () => {
      const config = createConfig({image: "nix:sets=python"});
      const command = nixExecutor.getCommand(config);
      const packagesLine = command.find((c) => c.includes("Packages:"));

      expect(packagesLine).toContain("python312");
      expect(packagesLine).toContain("python312Packages.pip");
    });

    it("should expand multiple sets", () => {
      const config = createConfig({
        image: "nix:sets=nodejs,python,kubernetes",
      });
      const command = nixExecutor.getCommand(config);
      const packagesLine = command.find((c) => c.includes("Packages:"));

      expect(packagesLine).toContain("nodejs_24");
      expect(packagesLine).toContain("python312");
      expect(packagesLine).toContain("kubectl");
    });

    it("should expand rust set", () => {
      const config = createConfig({image: "nix:sets=rust"});
      const command = nixExecutor.getCommand(config);
      const packagesLine = command.find((c) => c.includes("Packages:"));

      expect(packagesLine).toContain("rustc");
      expect(packagesLine).toContain("cargo");
    });

    it("should expand kubernetes set", () => {
      const config = createConfig({image: "nix:sets=kubernetes"});
      const command = nixExecutor.getCommand(config);
      const packagesLine = command.find((c) => c.includes("Packages:"));

      expect(packagesLine).toContain("kubectl");
      expect(packagesLine).toContain("kubernetes-helm");
      expect(packagesLine).toContain("k9s");
    });
  });

  describe("JSON config parsing", () => {
    it("should parse sets from JSON", () => {
      const config = createConfig({image: 'nix:{"sets":"nodejs,python"}'});
      const command = nixExecutor.getCommand(config);
      const packagesLine = command.find((c) => c.includes("Packages:"));

      expect(packagesLine).toContain("nodejs_24");
      expect(packagesLine).toContain("python312");
    });

    it("should parse packages from JSON string", () => {
      const config = createConfig({image: 'nix:{"packages":"tree,htop"}'});
      const command = nixExecutor.getCommand(config);
      const packagesLine = command.find((c) => c.includes("Packages:"));

      expect(packagesLine).toContain("tree");
      expect(packagesLine).toContain("htop");
    });

    it("should parse packages from JSON array", () => {
      const config = createConfig({
        image: 'nix:{"packages":["tree","htop"]}',
      });
      const command = nixExecutor.getCommand(config);
      const packagesLine = command.find((c) => c.includes("Packages:"));

      expect(packagesLine).toContain("tree");
      expect(packagesLine).toContain("htop");
    });

    it("should combine sets and extra packages", () => {
      const config = createConfig({
        image: 'nix:{"sets":"nodejs","packages":"tree,htop"}',
      });
      const command = nixExecutor.getCommand(config);
      const packagesLine = command.find((c) => c.includes("Packages:"));

      expect(packagesLine).toContain("nodejs_24");
      expect(packagesLine).toContain("tree");
      expect(packagesLine).toContain("htop");
    });
  });

  describe("simple key=value parsing", () => {
    it("should parse preset=claude", () => {
      const config = createConfig({image: "nix:preset=claude"});
      const command = nixExecutor.getCommand(config);

      // Should include default packages
      expect(command.some((c) => c.includes("Packages:"))).toBe(true);
    });

    it("should parse packages= with noDefaults", () => {
      const config = createConfig({image: "nix:packages=nodejs_24,git"});
      const command = nixExecutor.getCommand(config);
      const packagesLine = command.find((c) => c.includes("Packages:"));

      expect(packagesLine).toContain("nodejs_24");
      expect(packagesLine).toContain("git");
    });
  });

  describe("default packages", () => {
    it("should include base packages with sets", () => {
      const config = createConfig({image: "nix:sets=nodejs"});
      const command = nixExecutor.getCommand(config);
      const packagesLine = command.find((c) => c.includes("Packages:"));

      // Base packages should be included
      expect(packagesLine).toContain("git");
      expect(packagesLine).toContain("ripgrep");
      expect(packagesLine).toContain("coreutils");
    });
  });
});
