{
  description = "docker-agent - Agent container image and compose setup";

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
          # Docker / Containers
          pkgs.hadolint
          pkgs.dive
          pkgs.trivy
        ];
      };
    };
}
