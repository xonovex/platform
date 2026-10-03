{
  description = "shared-core-go - Shared Go core library";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    nixShells = {
      url = "path:../../../nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  # The Go toolchain (go + golangci-lint) comes from the shared nix/ `go` devShell.
  # `default` and the named `go` devShell are the same shell, so the nix toolchain
  # plugin's shellByTag `go` routing resolves to it (and not just the bare default).
  outputs = { nixpkgs, nixShells, ... }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    in {
      devShells = forAllSystems (system:
        let go = nixShells.devShells.${system}.go;
        in {
          default = go;
          go = go;
        });
    };
}
