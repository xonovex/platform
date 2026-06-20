{ pkgs }:
{
  # Version-agnostic Go tooling. The `go` compiler is per-project (go.mod-matched),
  # so each consumer adds its own pkgs.go_1_2x.
  packages = [
    pkgs.golangci-lint
    pkgs.git
  ];
}
