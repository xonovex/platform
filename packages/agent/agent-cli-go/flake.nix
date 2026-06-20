{
  description = "agent-cli-go - CLI tool for agent orchestration";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    nixShells = {
      url = "path:../../../nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  # Shared Go + shell tooling from nix/, plus this project's go.mod-matched Go toolchain.
  outputs = { nixpkgs, nixShells, ... }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
      g = nixShells.devShells.${system};
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        inputsFrom = [ g.go g.shell ];
        packages = [ pkgs.go_1_25 ];
      };
    };
}
