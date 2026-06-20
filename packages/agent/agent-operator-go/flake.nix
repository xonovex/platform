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
      system = "x86_64-linux";
      pkgs = nixpkgs.legacyPackages.${system};
      g = nixShells.devShells.${system};
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        inputsFrom = [ g.go g.shell g.k8s ];
      };
    };
}
