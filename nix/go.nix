{ pkgs }:
{
  packages = [
    pkgs.go_1_26
    pkgs.golangci-lint
    pkgs.git
  ];
}
