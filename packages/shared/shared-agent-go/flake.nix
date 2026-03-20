{
  description = "shared-agent-go - Shared Go agent library";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = { nixpkgs, ... }:
    let
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        packages = [
          # Go
          pkgs.go_1_25
          pkgs.golangci-lint
        ];
      };
    };
}
