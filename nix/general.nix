{ pkgs }:
let
  spaceGrotesk = pkgs.runCommand "xonovex-space-grotesk" { } ''
    install -Dm644 ${pkgs.fetchurl {
      url = "https://fonts.gstatic.com/s/spacegrotesk/v22/V8mQoQDjQSkFtoMM3T6r8E7mF71Q-gOoraIAEj42Vksj.ttf";
      hash = "sha256-bANGuNKX69wiWDKDPgPohKJq2Z0mXs05JNRuG6KF6oc=";
    }} "$out/share/fonts/truetype/SpaceGrotesk-SemiBold.ttf"
  '';
  diagramFonts = pkgs.writeText "xonovex-diagram-fonts.conf" ''
    <?xml version="1.0"?>
    <!DOCTYPE fontconfig SYSTEM "urn:fontconfig:fonts.dtd">
    <fontconfig>
      <dir>${pkgs.inter}/share/fonts</dir>
      <dir>${spaceGrotesk}/share/fonts</dir>
      <cachedir prefix="xdg">fontconfig</cachedir>
    </fontconfig>
  '';
in
{
  packages = [
    pkgs.git # git for the moon-plugin release tasks
    pkgs.graphviz # dot renders the asset-diagrams PNGs from their .dot sources
    pkgs.fontconfig # fc-match verifies diagram fonts before rendering
    pkgs.file # file reads the MIME type the asset-images check asserts
    pkgs.uv # uv runs the shipped portable auditor the parity test compares against
    # uv resolves the interpreter the portable auditor asks for from PATH rather
    # than downloading one, which ci-check has no network to do.
    pkgs.python3
  ];
  shellHook = ''
    export XONOVEX_DIAGRAM_FONTCONFIG_FILE="${diagramFonts}"
  '';
}
