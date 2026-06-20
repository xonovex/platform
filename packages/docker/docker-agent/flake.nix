{
  description = "docker-agent - Agent container image and compose setup";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    nixShells = {
      url = "path:../../../nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  # Container linting/scanning tooling comes from the shared nix/ `docker` devShell.
  outputs = { nixShells, ... }: {
    devShells.x86_64-linux.default = nixShells.devShells.x86_64-linux.docker;
  };
}
