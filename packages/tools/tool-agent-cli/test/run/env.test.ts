import { homedir } from "node:os";
import { join } from "node:path";
import { describe, expect, it } from "vitest";
import { renderNixExpression } from "../../src/run/env/render.js";
import {
  computeEnvId,
  getAgentNixDir,
  getAgentsDir,
  getEnvsDir,
  getSpecsDir,
  normalizeEnvSpec,
  resolveEnv,
  validateEnvSpec,
  validatePackageName,
} from "../../src/run/env/resolve.js";
import {
  DEFAULT_BASE_PACKAGES,
  DEFAULT_NIXPKGS_PIN,
  NIXPKGS_PINS,
} from "../../src/run/env/types.js";

describe("env", () => {
  describe("types", () => {
    describe("NIXPKGS_PINS", () => {
      it("should have nixos-24.11 pin", () => {
        const pin = NIXPKGS_PINS["nixos-24.11"];
        expect(pin).toBeDefined();
        expect(pin?.ref).toBe("nixos-24.11");
      });

      it("should have nixos-unstable pin", () => {
        expect(NIXPKGS_PINS["nixos-unstable"]).toBeDefined();
      });
    });

    describe("DEFAULT_BASE_PACKAGES", () => {
      it("should include nodejs_24", () => {
        expect(DEFAULT_BASE_PACKAGES).toContain("nodejs_24");
      });

      it("should not include agent-specific packages", () => {
        // Agent packages are added dynamically based on the agent being used
        expect(DEFAULT_BASE_PACKAGES).not.toContain("claude-code");
        expect(DEFAULT_BASE_PACKAGES).not.toContain("opencode");
      });

      it("should include essential tools", () => {
        expect(DEFAULT_BASE_PACKAGES).toContain("git");
        expect(DEFAULT_BASE_PACKAGES).toContain("ripgrep");
        expect(DEFAULT_BASE_PACKAGES).toContain("fd");
        expect(DEFAULT_BASE_PACKAGES).toContain("jq");
      });
    });

    describe("DEFAULT_NIXPKGS_PIN", () => {
      it("should be a valid pin", () => {
        expect(NIXPKGS_PINS[DEFAULT_NIXPKGS_PIN]).toBeDefined();
      });
    });
  });

  describe("resolve", () => {
    describe("getAgentNixDir", () => {
      it("should return path in home directory", () => {
        const dir = getAgentNixDir();
        expect(dir).toBe(join(homedir(), ".local", "share", "agent-nix"));
      });
    });

    describe("getSpecsDir", () => {
      it("should return specs subdirectory", () => {
        const dir = getSpecsDir();
        expect(dir).toBe(
          join(homedir(), ".local", "share", "agent-nix", "specs"),
        );
      });
    });

    describe("getEnvsDir", () => {
      it("should return envs subdirectory", () => {
        const dir = getEnvsDir();
        expect(dir).toBe(
          join(homedir(), ".local", "share", "agent-nix", "envs"),
        );
      });
    });

    describe("getAgentsDir", () => {
      it("should return agents subdirectory", () => {
        const dir = getAgentsDir();
        expect(dir).toBe(
          join(homedir(), ".local", "share", "agent-nix", "agents"),
        );
      });
    });

    describe("validatePackageName", () => {
      it("should accept valid package names", () => {
        expect(validatePackageName("nodejs_24")).toBe(true);
        expect(validatePackageName("python312")).toBe(true);
        expect(validatePackageName("clang_18")).toBe(true);
        expect(validatePackageName("git")).toBe(true);
        expect(validatePackageName("gcc-unwrapped")).toBe(true);
        expect(validatePackageName("lua5_4")).toBe(true);
      });

      it("should reject invalid package names", () => {
        expect(validatePackageName("")).toBe(false);
        expect(validatePackageName("pkg with space")).toBe(false);
        expect(validatePackageName("pkg;rm -rf")).toBe(false);
        expect(validatePackageName("pkg`whoami`")).toBe(false);
        expect(validatePackageName("$(evil)")).toBe(false);
      });
    });

    describe("validateEnvSpec", () => {
      it("should accept valid spec", () => {
        expect(() => {
          validateEnvSpec({
            nixpkgs_pin: "nixos-24.11",
            packages: ["git", "nodejs_24"],
          });
        }).not.toThrow();
      });

      it("should reject invalid nixpkgs_pin", () => {
        expect(() => {
          validateEnvSpec({
            nixpkgs_pin: "invalid-pin",
            packages: ["git"],
          });
        }).toThrow(/Invalid nixpkgs_pin/);
      });

      it("should reject empty packages", () => {
        expect(() => {
          validateEnvSpec({
            nixpkgs_pin: "nixos-24.11",
            packages: [],
          });
        }).toThrow(/non-empty/);
      });

      it("should reject invalid package names", () => {
        expect(() => {
          validateEnvSpec({
            nixpkgs_pin: "nixos-24.11",
            packages: ["git", "evil;rm -rf"],
          });
        }).toThrow(/Invalid package name/);
      });
    });

    describe("normalizeEnvSpec", () => {
      it("should sort packages", () => {
        const spec = normalizeEnvSpec({
          nixpkgs_pin: "nixos-24.11",
          packages: ["zlib", "git", "bash"],
        });
        expect(spec.packages).toEqual(["bash", "git", "zlib"]);
      });

      it("should deduplicate packages", () => {
        const spec = normalizeEnvSpec({
          nixpkgs_pin: "nixos-24.11",
          packages: ["git", "bash", "git", "bash"],
        });
        expect(spec.packages).toEqual(["bash", "git"]);
      });

      it("should use default pin if not specified", () => {
        const spec = normalizeEnvSpec({
          nixpkgs_pin: "",
          packages: ["git"],
        });
        expect(spec.nixpkgs_pin).toBe(DEFAULT_NIXPKGS_PIN);
      });
    });

    describe("computeEnvId", () => {
      it("should return consistent hash for same spec", () => {
        const spec = { nixpkgs_pin: "nixos-24.11", packages: ["git", "bash"] };
        const id1 = computeEnvId(spec);
        const id2 = computeEnvId(spec);
        expect(id1).toBe(id2);
      });

      it("should return different hash for different packages", () => {
        const id1 = computeEnvId({
          nixpkgs_pin: "nixos-24.11",
          packages: ["git"],
        });
        const id2 = computeEnvId({
          nixpkgs_pin: "nixos-24.11",
          packages: ["bash"],
        });
        expect(id1).not.toBe(id2);
      });

      it("should return different hash for different pins", () => {
        const id1 = computeEnvId({
          nixpkgs_pin: "nixos-24.11",
          packages: ["git"],
        });
        const id2 = computeEnvId({
          nixpkgs_pin: "nixos-unstable",
          packages: ["git"],
        });
        expect(id1).not.toBe(id2);
      });

      it("should return 16 character hex string", () => {
        const id = computeEnvId({
          nixpkgs_pin: "nixos-24.11",
          packages: ["git"],
        });
        expect(id).toMatch(/^[a-f0-9]{16}$/);
      });
    });

    describe("resolveEnv", () => {
      it("should return resolved env with paths", () => {
        const resolved = resolveEnv({
          nixpkgs_pin: "nixos-24.11",
          packages: ["git", "bash"],
        });

        expect(resolved.envId).toBeDefined();
        expect(resolved.specPath).toContain(resolved.envId);
        expect(resolved.specPath).toContain(".nix");
        expect(resolved.outLink).toContain(resolved.envId);
      });

      it("should normalize packages in envId", () => {
        const resolved1 = resolveEnv({
          nixpkgs_pin: "nixos-24.11",
          packages: ["bash", "git"],
        });
        const resolved2 = resolveEnv({
          nixpkgs_pin: "nixos-24.11",
          packages: ["git", "bash"],
        });

        // Same packages in different order should get same envId
        expect(resolved1.envId).toBe(resolved2.envId);
      });
    });
  });

  describe("render", () => {
    describe("renderNixExpression", () => {
      it("should generate valid nix expression", () => {
        const expr = renderNixExpression(
          { nixpkgs_pin: "nixos-24.11", packages: ["git", "bash"] },
          "abc123",
        );

        expect(expr).toContain("fetchTarball");
        expect(expr).toContain("nixos-24.11");
        expect(expr).toContain("pkgs.buildEnv");
        expect(expr).toContain("agent-env-abc123");
        expect(expr).toContain("git");
        expect(expr).toContain("bash");
      });

      it("should include allowUnfree config", () => {
        const expr = renderNixExpression(
          { nixpkgs_pin: "nixos-24.11", packages: ["git"] },
          "test",
        );

        expect(expr).toContain("allowUnfree = true");
      });

      it("should include envId in comment", () => {
        const expr = renderNixExpression(
          { nixpkgs_pin: "nixos-24.11", packages: ["git"] },
          "myenvid123",
        );

        expect(expr).toContain("EnvID: myenvid123");
      });
    });
  });
});
