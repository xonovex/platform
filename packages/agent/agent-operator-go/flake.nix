{
  description = "agent-operator-go - Kubernetes operator for agent orchestration";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    nixShells = {
      url = "path:../../../nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  # Shared Go + shell + Kubernetes tooling from nix/.
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
          # `default` and the named `go` devShell are the same composite shell, so the
          # nix toolchain plugin's shellByTag `go` routing resolves to it.
          devShell = pkgs.mkShell {
            inputsFrom = [ g.go g.shell g.k8s ];
          };
        in {
          default = devShell;
          go = devShell;
        });
    };
}
