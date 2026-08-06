{ pkgs }:
{
  packages = [
    pkgs.git # git for the moon-plugin release tasks
    pkgs.graphviz # dot renders the asset-diagrams PNGs from their .dot sources
    pkgs.file # file reads the MIME type the asset-images check asserts
    pkgs.uv # uv runs the shipped portable auditor the parity test compares against
    # uv resolves the interpreter the portable auditor asks for from PATH rather
    # than downloading one, which ci-check has no network to do.
    pkgs.python3
  ];
}
