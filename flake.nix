{
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
      devShells = forAllSystems (system:
        let
          pkgs = import nixpkgs {
            inherit system;
            overlays = [ rust-overlay.overlays.default ];
          };
          rustToolchain = pkgs.rust-bin.stable.latest.default.override {
            targets = [ "wasm32-wasip1" ];
          };

          # Tool groups, composed into the full `default` shell and lean
          # per-purpose shells selected per project via the nix toolchain's
          # `shellByTag` setting (see .moon/toolchains.yml).
          general = [ pkgs.git ]; # git for the moon-plugin release tasks
          node = [ pkgs.nodejs_24 ];
          go = [ pkgs.go_1_26 pkgs.golangci-lint ];
          k8s = [ pkgs.kind pkgs.kubectl pkgs.setup-envtest pkgs.kubernetes-controller-tools ];
          shell = [ pkgs.shellcheck pkgs.shfmt ];
          rust = [ rustToolchain pkgs.binaryen pkgs.wabt ]; # cargo/clippy/rustfmt + wasm-opt + wasm tools
          release = [ pkgs.gh ]; # gh for the moon-plugin GitHub releases
          ci = [ pkgs.zizmor ]; # GitHub Actions workflow linter
        in
        {
          # Full shell — every task that does not select a named shell uses this.
          default = pkgs.mkShell {
            packages = node ++ go ++ k8s ++ shell ++ rust ++ release ++ ci ++ general;
          };

          # Lean per-purpose shells, selected via the nix toolchain `shellByTag` setting.
          go = pkgs.mkShell { packages = go ++ general; };
          shell = pkgs.mkShell { packages = shell ++ general; };
          rust = pkgs.mkShell { packages = rust ++ release ++ general; };
        }
      );
    };
}
