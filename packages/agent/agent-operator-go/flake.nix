{
  description = "agent-operator-go - Kubernetes operator for agent orchestration";

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
        ];
      };
    };
}
