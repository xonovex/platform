{ pkgs }:
{
  packages = [
    pkgs.git # git for the moon-plugin release tasks
    pkgs.graphviz # dot renders the asset-diagrams PNGs from their .dot sources
    pkgs.file # file reads the MIME type the asset-images check asserts
  ];
}
