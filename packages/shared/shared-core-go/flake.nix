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
  outputs = { nixShells, ... }: {
    devShells.x86_64-linux.default = nixShells.devShells.x86_64-linux.go;
  };
}
