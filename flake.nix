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
        in
        {
          default = pkgs.mkShell {
            packages = [
              # JavaScript / TypeScript
              pkgs.nodejs_24

              # Go
              pkgs.go_1_26
              pkgs.golangci-lint

              # Kubernetes
              pkgs.kind
              pkgs.kubectl
              pkgs.setup-envtest
              pkgs.kubernetes-controller-tools

              # Shell
              pkgs.shellcheck
              pkgs.shfmt

              # Rust → WASM (moon plugins)
              rustToolchain
              pkgs.binaryen
              pkgs.wabt

              # Release tooling
              pkgs.gh
            ];
          };
        }
      );
    };
}
