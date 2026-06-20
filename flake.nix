{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    nixShells = {
      url = "path:./nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = { nixpkgs, nixShells, ... }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
    in
    {
      devShells = forAllSystems (system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          g = nixShells.devShells.${system};
        in
        {
          # Full shell — composed from the shared per-tool devShells in nix/.
          default = pkgs.mkShell {
            inputsFrom = [ g.node g.go g.k8s g.shell g.rust g.release g.ci g.general ];
          };

          # Lean per-purpose shells, selected via the nix toolchain `shellByTag` setting.
          go = pkgs.mkShell { inputsFrom = [ g.go g.general ]; };
          shell = pkgs.mkShell { inputsFrom = [ g.shell g.general ]; };
          rust = pkgs.mkShell { inputsFrom = [ g.rust g.release g.general ]; };
        }
      );
    };
}
