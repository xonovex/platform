# =============================================================================
# Devcontainer Package Groups — Centrally Approved CLI Tools
#
# Single source of truth for which packages developers can install.
# Developers request new packages via PR. They activate groups locally
# via PACKAGE_GROUPS in .env (gitignored).
#
# The "base" group is always installed. Optional groups are installed
# only when listed in PACKAGE_GROUPS (comma-separated).
#
# All versions are pinned via flake.lock (committed). Run `nix flake
# update` to bump versions, which updates the lock file for review.
# =============================================================================
{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = { nixpkgs, ... }:
    let
      # Devcontainer always runs on Linux x86_64.
      # Add aarch64-linux here if ARM containers are needed.
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};

      # Helper: create a buildEnv package from a list of packages.
      mkGroup = name: paths: pkgs.buildEnv {
        name = "devcontainer-${name}";
        inherit paths;
        # Ignore collisions between groups (e.g., both base and python
        # might pull in python3). The last writer wins, which is fine
        # since they're the same package from the same nixpkgs.
        ignoreCollisions = true;
      };
    in
    {
      packages.${system} = {
        # =================================================================
        # Base — always installed, not selectable
        # =================================================================
        base = mkGroup "base" (with pkgs; [
          # Core utilities
          git
          curl
          wget
          unzip
          jq
          vim
          nano

          # Build tools
          gnumake
          gcc

          # Process management
          procps

          # Node.js (matches the version used in the project root flake)
          nodejs_20

          # Shell
          bash
        ]);

        # =================================================================
        # Android — ADB client for emulator access via proxy relay
        # Requires: COMPOSE_PROFILES=android (for the port relay)
        # =================================================================
        android = mkGroup "android" (with pkgs; [
          android-tools  # adb, fastboot
        ]);

        # =================================================================
        # Python — pip and venv for Python workflows
        # =================================================================
        python = mkGroup "python" (with pkgs; [
          python3
          python3Packages.pip
          python3Packages.virtualenv
        ]);

        # =================================================================
        # Database — CLI clients for database access
        # =================================================================
        database = mkGroup "database" (with pkgs; [
          postgresql
          mysql-client
          redis
        ]);

        # =================================================================
        # Kubernetes — cluster management and development
        # =================================================================
        kubernetes = mkGroup "kubernetes" (with pkgs; [
          kubectl
          kind
          kubernetes-helm
        ]);

        # =================================================================
        # Go — Go development toolchain
        # =================================================================
        go = mkGroup "go" (with pkgs; [
          go_1_24
          golangci-lint
          gopls
        ]);
      };
    };
}
