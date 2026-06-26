{
  description = "shared-agent-go - Shared Go agent library";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    nixShells = {
      url = "path:../../../nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  # The Go toolchain (go + golangci-lint) comes from the shared nix/ `go` devShell.
  # `default` and the named `go` devShell are the same shell, so the nix toolchain
  # plugin's shellByTag `go` routing resolves to it (and not just the bare default).
  outputs = { nixShells, ... }:
    let go = nixShells.devShells.x86_64-linux.go;
    in {
      devShells.x86_64-linux.default = go;
      devShells.x86_64-linux.go = go;
    };
}
