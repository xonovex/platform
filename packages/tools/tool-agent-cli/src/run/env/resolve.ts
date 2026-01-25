import { createHash } from "node:crypto";
import { existsSync, lstatSync, realpathSync } from "node:fs";
import { homedir } from "node:os";
import { join } from "node:path";
import {
  DEFAULT_NIXPKGS_PIN,
  NIXPKGS_PINS,
  type EnvSpec,
  type ResolvedEnv,
} from "./types.js";

/**
 * Base directory for agent-nix data
 */
export function getAgentNixDir(): string {
  return join(homedir(), ".local", "share", "agent-nix");
}

/**
 * Directory for storing generated .nix spec files
 */
export function getSpecsDir(): string {
  return join(getAgentNixDir(), "specs");
}

/**
 * Directory for storing nix-build output symlinks
 */
export function getEnvsDir(): string {
  return join(getAgentNixDir(), "envs");
}

/**
 * Directory for per-agent runtime data
 */
export function getAgentsDir(): string {
  return join(getAgentNixDir(), "agents");
}

/**
 * Regex for validating package names
 * Allows alphanumeric, underscore, plus, hyphen, and dot
 */
const PACKAGE_NAME_REGEX = /^[\w+\-.]+$/;

/**
 * Validate a single package name
 */
export function validatePackageName(name: string): boolean {
  return PACKAGE_NAME_REGEX.test(name);
}

/**
 * Validate an EnvSpec
 * @throws Error if validation fails
 */
export function validateEnvSpec(spec: EnvSpec): void {
  // Validate nixpkgs_pin
  if (!spec.nixpkgs_pin) {
    throw new Error("nixpkgs_pin is required");
  }
  if (!(spec.nixpkgs_pin in NIXPKGS_PINS)) {
    const allowed = Object.keys(NIXPKGS_PINS).join(", ");
    throw new Error(
      `Invalid nixpkgs_pin "${spec.nixpkgs_pin}". Allowed: ${allowed}`,
    );
  }

  // Validate packages
  if (spec.packages.length === 0) {
    throw new Error("packages must be a non-empty array");
  }

  for (const pkg of spec.packages) {
    if (!validatePackageName(pkg)) {
      throw new Error(
        `Invalid package name "${pkg}". Must match ${PACKAGE_NAME_REGEX.source}`,
      );
    }
  }
}

/**
 * Normalize an EnvSpec by sorting and deduplicating packages
 */
export function normalizeEnvSpec(spec: EnvSpec): EnvSpec {
  const packages = [...new Set(spec.packages)].toSorted();
  return {
    nixpkgs_pin: spec.nixpkgs_pin || DEFAULT_NIXPKGS_PIN,
    packages,
  };
}

/**
 * Compute the environment ID from a normalized EnvSpec
 * Uses sha256 of "nixpkgs_pin\npackage1\npackage2\n..."
 */
export function computeEnvId(spec: EnvSpec): string {
  const content = spec.nixpkgs_pin + "\n" + spec.packages.join("\n");
  return createHash("sha256").update(content).digest("hex").slice(0, 16);
}

/**
 * Check if an outLink symlink points to a valid store path
 */
function isValidOutLink(outLink: string): boolean {
  if (!existsSync(outLink)) {
    return false;
  }

  try {
    const stat = lstatSync(outLink);
    if (!stat.isSymbolicLink()) {
      return false;
    }

    const target = realpathSync(outLink);
    return target.startsWith("/nix/store/") && existsSync(target);
  } catch {
    return false;
  }
}

/**
 * Resolve an EnvSpec to paths and check cache status
 */
export function resolveEnv(spec: EnvSpec): ResolvedEnv {
  validateEnvSpec(spec);
  const normalized = normalizeEnvSpec(spec);
  const envId = computeEnvId(normalized);

  const specPath = join(getSpecsDir(), `${envId}.nix`);
  const outLink = join(getEnvsDir(), envId);
  const ready = isValidOutLink(outLink);

  return {
    envId,
    specPath,
    outLink,
    ready,
  };
}
