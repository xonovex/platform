{ pkgs }:
{
  packages = [
    pkgs.kind
    pkgs.kubectl
    pkgs.setup-envtest
    pkgs.kubernetes-controller-tools
  ];
}
