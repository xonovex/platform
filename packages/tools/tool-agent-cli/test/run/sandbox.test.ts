import { homedir } from "node:os";
import { join } from "node:path";
import { describe, expect, it } from "vitest";
import { claudeAgent } from "../../src/run/agents/claude/index.js";
import { opencodeAgent } from "../../src/run/agents/opencode/index.js";
import {
  getApplicationBindMounts,
  getDefaultSandboxHome,
  getHostOverlayBindMounts,
  getSystemBindMounts,
  getUserConfigBindMounts,
} from "../../src/run/sandbox/bindings.js";
import {
  envToBwrapArgs,
  envToDockerArgs,
  getSandboxEnvironment,
  mergeCustomEnv,
  parseCustomEnv,
} from "../../src/run/sandbox/environment.js";
import { getExecutor } from "../../src/run/sandbox/index.js";
import type {
  SandboxConfig,
  SandboxMethod,
} from "../../src/run/sandbox/types.js";

describe("sandbox", () => {
  describe("getDefaultSandboxHome", () => {
    it("should return path in home directory", () => {
      const home = getDefaultSandboxHome();
      expect(home).toBe(join(homedir(), ".sandboxed-agent-home"));
    });
  });

  describe("getExecutor", () => {
    it("should return none executor for none method", () => {
      const executor = getExecutor("none");
      expect(executor).toBeDefined();
    });

    it("should return bwrap executor for bwrap method", () => {
      const executor = getExecutor("bwrap");
      expect(executor).toBeDefined();
    });

    it("should return docker executor for docker method", () => {
      const executor = getExecutor("docker");
      expect(executor).toBeDefined();
    });

    it("should return none executor for unknown method", () => {
      // @ts-expect-error Testing fallback behavior
      const executor = getExecutor("unknown");
      expect(executor).toBeDefined();
    });
  });

  describe("SandboxConfig", () => {
    it("should accept valid config with claude agent", () => {
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

      expect(config.method).toBe("none");
      expect(config.agent).toBe(claudeAgent);
      expect(config.workDir).toBe("/home/user/project");
      expect(config.network).toBe(true);
    });

    it("should accept valid config with opencode agent", () => {
      const config: SandboxConfig = {
        agentId: "test-agent",
        method: "bwrap",
        agent: opencodeAgent,
        workDir: "/home/user/project",
        network: false,
        bindPaths: ["/extra/path"],
        roBindPaths: ["/readonly/path"],
        customEnv: [],
        agentArgs: [],
        verbose: true,
        debug: true,
        dryRun: false,
      };

      expect(config.method).toBe("bwrap");
      expect(config.agent).toBe(opencodeAgent);
      expect(config.network).toBe(false);
      expect(config.bindPaths).toEqual(["/extra/path"]);
      expect(config.roBindPaths).toEqual(["/readonly/path"]);
    });

    it("should accept config without agent (defaults to claude)", () => {
      const config: SandboxConfig = {
        agentId: "test-agent",
        method: "docker",
        workDir: "/home/user/project",
        network: true,
        bindPaths: [],
        roBindPaths: [],
        customEnv: [],
        agentArgs: [],
        verbose: false,
        debug: false,
        dryRun: true,
      };

      expect(config.method).toBe("docker");
      expect(config.agent).toBeUndefined();
      expect(config.dryRun).toBe(true);
    });
  });

  describe("SandboxMethod", () => {
    it("should accept valid methods", () => {
      const methods: SandboxMethod[] = [
        "none",
        "bwrap",
        "docker",
        "compose",
        "nix",
      ];
      expect(methods).toHaveLength(5);
    });
  });

  describe("bindings", () => {
    describe("getSystemBindMounts", () => {
      it("should return system-level mounts", () => {
        const mounts = getSystemBindMounts();
        expect(mounts).toHaveLength(4);

        const sources = mounts.map((m) => m.source);
        expect(sources).toContain("/");
        expect(sources).toContain("/dev");
        expect(sources).toContain("/proc");
        expect(sources).toContain("/tmp");
      });

      it("should have root mounted read-only", () => {
        const mounts = getSystemBindMounts();
        const rootMount = mounts.find((m) => m.source === "/");
        expect(rootMount?.readOnly).toBe(true);
      });

      it("should have special mount types for dev and proc", () => {
        const mounts = getSystemBindMounts();
        const devMount = mounts.find((m) => m.source === "/dev");
        const procMount = mounts.find((m) => m.source === "/proc");
        expect(devMount?.type).toBe("dev-bind");
        expect(procMount?.type).toBe("proc");
      });
    });

    describe("getUserConfigBindMounts", () => {
      it("should map config paths to target home", () => {
        const home = homedir();
        const targetHome = "/sandbox/home";
        const mounts = getUserConfigBindMounts(targetHome);

        // All mounts should target the specified home
        for (const mount of mounts) {
          expect(mount.target.startsWith(targetHome)).toBe(true);
          expect(mount.source.startsWith(home)).toBe(true);
        }
      });

      it("should support read-only option", () => {
        const mounts = getUserConfigBindMounts("/home", true);
        for (const mount of mounts) {
          expect(mount.readOnly).toBe(true);
        }
      });
    });

    describe("getApplicationBindMounts", () => {
      it("should include workdir mount", () => {
        const workDir = "/home/user/project";
        const mounts = getApplicationBindMounts(workDir, homedir());

        const workMount = mounts.find((m) => m.source === workDir);
        expect(workMount).toBeDefined();
        expect(workMount?.target).toBe(workDir);
        expect(workMount?.readOnly).toBe(false);
      });

      it("should include user config mounts", () => {
        const mounts = getApplicationBindMounts("/work", homedir());
        // Should have more than just the workdir
        expect(mounts.length).toBeGreaterThan(1);
      });

      it("should include custom bind paths", () => {
        const customPath = "/tmp"; // Use /tmp as it exists
        const mounts = getApplicationBindMounts(
          "/work",
          homedir(),
          [customPath],
          [],
        );

        const customMount = mounts.find((m) => m.source === customPath);
        expect(customMount).toBeDefined();
        expect(customMount?.readOnly).toBe(false);
      });

      it("should include custom read-only bind paths", () => {
        const customPath = "/tmp";
        const mounts = getApplicationBindMounts(
          "/work",
          homedir(),
          [],
          [customPath],
        );

        const customMount = mounts.find((m) => m.source === customPath);
        expect(customMount).toBeDefined();
        expect(customMount?.readOnly).toBe(true);
      });
    });

    describe("getHostOverlayBindMounts", () => {
      it("should include system mounts", () => {
        const mounts = getHostOverlayBindMounts("/sandbox/home", "/work");
        const sources = mounts.map((m) => m.source);

        expect(sources).toContain("/");
        expect(sources).toContain("/dev");
        expect(sources).toContain("/proc");
        expect(sources).toContain("/tmp");
      });

      it("should include sandbox home mapping", () => {
        const sandboxHome = "/sandbox/home";
        const mounts = getHostOverlayBindMounts(sandboxHome, "/work");

        const homeMount = mounts.find((m) => m.source === sandboxHome);
        expect(homeMount).toBeDefined();
        expect(homeMount?.target).toBe(homedir());
      });

      it("should include application mounts", () => {
        const workDir = "/home/user/project";
        const mounts = getHostOverlayBindMounts("/sandbox/home", workDir);

        const workMount = mounts.find((m) => m.source === workDir);
        expect(workMount).toBeDefined();
      });
    });
  });

  describe("environment", () => {
    describe("getSandboxEnvironment", () => {
      it("should return HOME, TMPDIR, and PATH", () => {
        const env = getSandboxEnvironment("/home/user");

        expect(env.HOME).toBe("/home/user");
        expect(env.TMPDIR).toBe("/tmp");
        expect(env.PATH).toBeDefined();
      });

      it("should include .local/bin in PATH", () => {
        const env = getSandboxEnvironment("/home/user");

        expect(env.PATH).toContain("/home/user/.local/bin");
      });

      it("should include standard paths in PATH", () => {
        const env = getSandboxEnvironment("/home/user");

        expect(env.PATH).toContain("/usr/local/bin");
        expect(env.PATH).toContain("/usr/bin");
        expect(env.PATH).toContain("/bin");
      });

      it("should prepend extra paths to PATH", () => {
        const env = getSandboxEnvironment("/home/user", [
          "/env/bin",
          "/custom/bin",
        ]);

        expect(env.PATH.startsWith("/env/bin:")).toBe(true);
        expect(env.PATH).toContain("/custom/bin");
      });

      it("should order paths correctly with extra prefixes first", () => {
        const env = getSandboxEnvironment("/home/user", ["/env/bin"]);
        const paths = env.PATH.split(":");

        expect(paths[0]).toBe("/env/bin");
        expect(paths[1]).toBe("/home/user/.local/bin");
      });

      it("should include SHELL in environment", () => {
        const env = getSandboxEnvironment("/home/user");

        expect(env.SHELL).toBeDefined();
      });

      it("should default SHELL to /bin/bash", () => {
        const env = getSandboxEnvironment("/home/user");

        expect(env.SHELL).toBe("/bin/bash");
      });

      it("should allow custom shell path", () => {
        const env = getSandboxEnvironment(
          "/home/user",
          [],
          true,
          "/env/bin/bash",
        );

        expect(env.SHELL).toBe("/env/bin/bash");
      });

      it("should use custom shell for nix sandbox", () => {
        const env = getSandboxEnvironment(
          "/home/user",
          ["/env/bin"],
          false,
          "/env/bin/bash",
        );

        expect(env.SHELL).toBe("/env/bin/bash");
        expect(env.PATH.startsWith("/env/bin:")).toBe(true);
      });
    });

    describe("envToBwrapArgs", () => {
      it("should convert env to --setenv arguments", () => {
        const env = {
          HOME: "/home/user",
          TMPDIR: "/tmp",
          PATH: "/usr/bin",
          SHELL: "/bin/bash",
        };
        const args = envToBwrapArgs(env);

        expect(args).toContain("--setenv");
        expect(args).toContain("HOME");
        expect(args).toContain("/home/user");
      });

      it("should include all environment variables", () => {
        const env = {
          HOME: "/home",
          TMPDIR: "/tmp",
          PATH: "/bin",
          SHELL: "/bin/bash",
          CUSTOM: "value",
        };
        const args = envToBwrapArgs(env);

        // Each env var produces 3 args: --setenv, key, value
        expect(args.length).toBe(15);
      });

      it("should produce valid bwrap argument pairs", () => {
        const env = {
          HOME: "/home/user",
          TMPDIR: "/tmp",
          PATH: "/usr/bin",
          SHELL: "/bin/bash",
        };
        const args = envToBwrapArgs(env);

        // Check structure: --setenv KEY VALUE
        for (let i = 0; i < args.length; i += 3) {
          expect(args[i]).toBe("--setenv");
          expect(typeof args[i + 1]).toBe("string");
          expect(typeof args[i + 2]).toBe("string");
        }
      });
    });

    describe("envToDockerArgs", () => {
      it("should convert env to -e arguments", () => {
        const env = {
          HOME: "/home/user",
          TMPDIR: "/tmp",
          PATH: "/usr/bin",
          SHELL: "/bin/bash",
        };
        const args = envToDockerArgs(env);

        expect(args).toContain("-e");
        expect(args).toContain("HOME=/home/user");
      });

      it("should include all environment variables", () => {
        const env = {
          HOME: "/home",
          TMPDIR: "/tmp",
          PATH: "/bin",
          SHELL: "/bin/bash",
        };
        const args = envToDockerArgs(env);

        // Each env var produces 2 args: -e, KEY=VALUE
        expect(args.length).toBe(8);
      });

      it("should produce valid docker argument pairs", () => {
        const env = {
          HOME: "/home/user",
          TMPDIR: "/tmp",
          PATH: "/usr/bin",
          SHELL: "/bin/bash",
        };
        const args = envToDockerArgs(env);

        // Check structure: -e KEY=VALUE
        for (let i = 0; i < args.length; i += 2) {
          expect(args[i]).toBe("-e");
          expect(args[i + 1]).toContain("=");
        }
      });
    });

    describe("parseCustomEnv", () => {
      it("should parse KEY=VALUE strings", () => {
        const result = parseCustomEnv(["FOO=bar", "BAZ=qux"]);

        expect(result.FOO).toBe("bar");
        expect(result.BAZ).toBe("qux");
      });

      it("should handle values with equals signs", () => {
        const result = parseCustomEnv(["URL=https://example.com?foo=bar"]);

        expect(result.URL).toBe("https://example.com?foo=bar");
      });

      it("should ignore invalid entries without equals", () => {
        const result = parseCustomEnv([
          "VALID=value",
          "INVALID",
          "ALSO_VALID=x",
        ]);

        expect(result.VALID).toBe("value");
        expect(result.ALSO_VALID).toBe("x");
        expect(result.INVALID).toBeUndefined();
      });

      it("should handle empty values", () => {
        const result = parseCustomEnv(["EMPTY="]);

        expect(result.EMPTY).toBe("");
      });

      it("should return empty object for empty array", () => {
        const result = parseCustomEnv([]);

        expect(Object.keys(result)).toHaveLength(0);
      });
    });

    describe("mergeCustomEnv", () => {
      it("should merge custom env into base env", () => {
        const base = {
          HOME: "/home/user",
          TMPDIR: "/tmp",
          PATH: "/bin",
          SHELL: "/bin/bash",
        };
        const result = mergeCustomEnv(base, ["CUSTOM=value"]);

        expect(result.HOME).toBe("/home/user");
        expect(result.CUSTOM).toBe("value");
      });

      it("should override base values with custom values", () => {
        const base = {
          HOME: "/home/user",
          TMPDIR: "/tmp",
          PATH: "/bin",
          SHELL: "/bin/bash",
        };
        const result = mergeCustomEnv(base, ["PATH=/custom/bin"]);

        expect(result.PATH).toBe("/custom/bin");
      });

      it("should preserve base values not overridden", () => {
        const base = {
          HOME: "/home/user",
          TMPDIR: "/tmp",
          PATH: "/bin",
          SHELL: "/bin/bash",
        };
        const result = mergeCustomEnv(base, ["CUSTOM=value"]);

        expect(result.HOME).toBe("/home/user");
        expect(result.TMPDIR).toBe("/tmp");
        expect(result.PATH).toBe("/bin");
      });

      it("should handle empty custom env", () => {
        const base = {
          HOME: "/home/user",
          TMPDIR: "/tmp",
          PATH: "/bin",
          SHELL: "/bin/bash",
        };
        const result = mergeCustomEnv(base, []);

        expect(result).toEqual(base);
      });
    });
  });

  describe("executor isAvailable", () => {
    it("none executor should always be available", async () => {
      const executor = getExecutor("none");
      const available = await executor.isAvailable();
      expect(available).toBe(true);
    });
  });

  describe("executor getCommand", () => {
    it("none executor should return agent command", () => {
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

    it("none executor should use opencode binary for opencode agent", () => {
      const executor = getExecutor("none");
      const config: SandboxConfig = {
        agentId: "test-agent",
        method: "none",
        agent: opencodeAgent,
        workDir: "/home/user/project",
        network: true,
        bindPaths: [],
        roBindPaths: [],
        customEnv: [],
        agentArgs: ["."],
        verbose: false,
        debug: false,
        dryRun: false,
      };

      const command = executor.getCommand(config);
      expect(command).toEqual(["opencode", "."]);
    });

    it("bwrap executor should include bwrap command", () => {
      const executor = getExecutor("bwrap");
      const config: SandboxConfig = {
        agentId: "test-agent",
        method: "bwrap",
        agent: claudeAgent,
        workDir: "/home/user/project",
        network: true,
        bindPaths: [],
        roBindPaths: [],
        customEnv: [],
        agentArgs: [],
        verbose: false,
        debug: false,
        dryRun: false,
      };

      const command = executor.getCommand(config);
      expect(command[0]).toBe("bwrap");
      // Individual unshare flags instead of --unshare-all to avoid Bun segfaults
      expect(command).toContain("--unshare-uts");
      expect(command).toContain("--unshare-ipc");
      expect(command).toContain("--unshare-pid");
      expect(command).toContain("--unshare-cgroup");
      expect(command).toContain("--share-net");
      expect(command).toContain("claude");
      expect(command).toContain("--permission-mode");
      expect(command).toContain("bypassPermissions");
    });

    it("bwrap executor should not include --share-net when network is disabled", () => {
      const executor = getExecutor("bwrap");
      const config: SandboxConfig = {
        agentId: "test-agent",
        method: "bwrap",
        agent: claudeAgent,
        workDir: "/home/user/project",
        network: false,
        bindPaths: [],
        roBindPaths: [],
        customEnv: [],
        agentArgs: [],
        verbose: false,
        debug: false,
        dryRun: false,
      };

      const command = executor.getCommand(config);
      expect(command).not.toContain("--share-net");
    });

    it("docker executor should include docker run command", () => {
      const executor = getExecutor("docker");
      const config: SandboxConfig = {
        agentId: "test-agent",
        method: "docker",
        agent: claudeAgent,
        workDir: "/home/user/project",
        network: true,
        bindPaths: [],
        roBindPaths: [],
        customEnv: [],
        agentArgs: [],
        verbose: false,
        debug: false,
        dryRun: false,
      };

      const command = executor.getCommand(config);
      expect(command[0]).toBe("docker");
      expect(command).toContain("run");
      expect(command).toContain("--rm");
      expect(command).toContain("-it");
      expect(command).toContain("claude");
    });

    it("docker executor should include --network none when network is disabled", () => {
      const executor = getExecutor("docker");
      const config: SandboxConfig = {
        agentId: "test-agent",
        method: "docker",
        agent: claudeAgent,
        workDir: "/home/user/project",
        network: false,
        bindPaths: [],
        roBindPaths: [],
        customEnv: [],
        agentArgs: [],
        verbose: false,
        debug: false,
        dryRun: false,
      };

      const command = executor.getCommand(config);
      const networkIndex = command.indexOf("--network");
      expect(networkIndex).toBeGreaterThan(-1);
      expect(command[networkIndex + 1]).toBe("none");
    });

    it("nix executor should return nix executor", () => {
      const executor = getExecutor("nix");
      expect(executor).toBeDefined();
    });

    it("nix executor getCommand should include bwrap and nix-specific mounts", () => {
      const executor = getExecutor("nix");
      const config: SandboxConfig = {
        agentId: "test-agent",
        method: "nix",
        agent: claudeAgent,
        workDir: "/home/user/project",
        network: true,
        bindPaths: [],
        roBindPaths: [],
        customEnv: [],
        agentArgs: [],
        verbose: false,
        debug: false,
        dryRun: false,
      };

      const command = executor.getCommand(config);
      // Should include comment about nix-build
      expect(command.some((c) => c.includes("nix-build"))).toBe(true);
      // Should include bwrap command
      expect(command.some((c) => c.includes("bwrap"))).toBe(true);
      // Should include /nix/store mount
      expect(command.some((c) => c.includes("/nix/store"))).toBe(true);
    });

    it("compose executor should return compose executor", () => {
      const executor = getExecutor("compose");
      expect(executor).toBeDefined();
    });
  });
});
