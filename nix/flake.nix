{
  description = "Composable per-tool devShells for the xonovex workspace and project flakes.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    rust-overlay = {
      url = "github:oxalica/rust-overlay";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { nixpkgs, rust-overlay, ... }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    in
    {
      # One devShell per tool group, each built from a small nix/<group>.nix module.
      # Workspace and project flakes compose these via `inputsFrom`, so every tool
      # group is defined exactly once.
      devShells = forAllSystems (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [ rust-overlay.overlays.default ];
          };
          group = file: pkgs.mkShell (import file { inherit pkgs; });
        in
        {
          general = group ./general.nix;
          node = group ./node.nix;
          go = group ./go.nix;
          shell = group ./shell.nix;
          k8s = group ./k8s.nix;
          rust = group ./rust.nix;
          release = group ./release.nix;
          ci = group ./ci.nix;
          docker = group ./docker.nix;
        }
      );
    };
}
