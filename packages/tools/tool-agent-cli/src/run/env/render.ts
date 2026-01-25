import { NIXPKGS_PINS, type EnvSpec } from "./types.js";

/**
 * Get the nixpkgs tarball URL for a pin
 */
function getNixpkgsTarballUrl(pin: string): string {
  const pinConfig = NIXPKGS_PINS[pin];
  if (!pinConfig) {
    throw new Error(`Unknown nixpkgs pin: ${pin}`);
  }
  return `https://github.com/NixOS/nixpkgs/archive/${pinConfig.ref}.tar.gz`;
}

/**
 * Render a Nix expression for building an agent environment
 *
 * Generates a pkgs.buildEnv expression with:
 * - fetchTarball for pinned nixpkgs
 * - allowUnfree = true (required for claude-code)
 * - All specified packages
 */
export function renderNixExpression(spec: EnvSpec, envId: string): string {
  const tarballUrl = getNixpkgsTarballUrl(spec.nixpkgs_pin);

  // Format packages list with proper indentation
  const packagesLines = spec.packages.map((pkg) => `    ${pkg}`).join("\n");

  return `# Auto-generated agent environment - do not edit
# EnvID: ${envId}
# Pin: ${spec.nixpkgs_pin}

{ pkgs ? import (fetchTarball "${tarballUrl}") {
    config.allowUnfree = true;
  }
}:

pkgs.buildEnv {
  name = "agent-env-${envId}";
  paths = with pkgs; [
${packagesLines}
  ];
}
`;
}
