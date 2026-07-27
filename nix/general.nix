{ pkgs }:
{
  packages = [
    pkgs.git # git for the moon-plugin release tasks
    pkgs.graphviz # dot renders the diagram-sandbox-isolation PNGs from their .dot sources
  ];
}
