/**
 * Nixpkgs channel pin configuration
 */
export interface NixpkgsPin {
  /** Channel name (e.g., "nixos-24.11") */
  name: string;
  /** Git revision or branch for fetchTarball */
  ref: string;
}

/**
 * Allowed nixpkgs pins - maps friendly names to tarball refs
 */
export const NIXPKGS_PINS: Record<string, NixpkgsPin> = {
  "nixos-24.11": {
    name: "nixos-24.11",
    ref: "nixos-24.11",
  },
  "nixos-unstable": {
    name: "nixos-unstable",
    ref: "nixos-unstable",
  },
  "nixpkgs-unstable": {
    name: "nixpkgs-unstable",
    ref: "nixpkgs-unstable",
  },
};

/**
 * Default nixpkgs pin to use
 */
export const DEFAULT_NIXPKGS_PIN = "nixos-unstable";

/**
 * Default base packages for agent environments (without agent-specific package)
 */
export const DEFAULT_BASE_PACKAGES = [
  "nodejs_24",
  "git",
  "ripgrep",
  "fd",
  "fzf",
  "jq",
  "curl",
  "coreutils",
  "bash",
] as const;

/**
 * Environment specification for building a Nix environment
 */
export interface EnvSpec {
  /** Nixpkgs pin/channel to use */
  nixpkgs_pin: string;
  /** List of nixpkgs attribute names to include */
  packages: string[];
}

/**
 * Result of resolving an environment specification
 */
export interface ResolvedEnv {
  /** Unique environment ID (sha256 hash) */
  envId: string;
  /** Path to the generated .nix spec file */
  specPath: string;
  /** Path to the nix-build output symlink */
  outLink: string;
  /** Whether the environment is already built (cache hit) */
  ready: boolean;
}

/**
 * Result of building a Nix environment
 */
export interface BuildResult {
  /** Whether the build succeeded */
  success: boolean;
  /** Path to the realized store path (if successful) */
  storePath?: string;
  /** Error message (if failed) */
  error?: string;
  /** Build duration in milliseconds */
  duration: number;
}

/**
 * Agent run input configuration
 */
export interface AgentRunInput {
  /** Environment specification */
  env: EnvSpec;
  /** Unique agent instance ID */
  agent_id: string;
  /** Command to execute */
  cmd: string[];
  /** Working directory inside sandbox */
  workdir: string;
  /** Whether to allow network access */
  network: boolean;
}
