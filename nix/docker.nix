{ pkgs }:
{
  packages = [
    pkgs.hadolint
    pkgs.dive
    pkgs.trivy
  ];
}
