{ pkgs }:
{
  # gh creates the GitHub release of the agent CLI.
  packages = [
    pkgs.gh
  ];
}
