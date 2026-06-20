{
  description = "shared-core-go - Shared Go core library";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    nixShells = {
      url = "path:../../../nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  # Shared Go tooling (golangci-lint) from nix/, plus this project's go.mod-matched
  # Go toolchain.
  outputs = { nixpkgs, nixShells, ... }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        inputsFrom = [ nixShells.devShells.${system}.go ];
        packages = [ pkgs.go_1_25 ];
      };
    };
}
