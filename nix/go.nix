{ pkgs }:
{
  packages = [
    pkgs.go_1_27
    pkgs.golangci-lint
    pkgs.git
    pkgs.tmux
  ];
}
