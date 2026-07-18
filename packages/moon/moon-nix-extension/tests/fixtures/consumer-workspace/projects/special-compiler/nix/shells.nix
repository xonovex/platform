{ pkgs }:
let
  marker = component: executable:
    pkgs.writeShellScriptBin executable ''
      printf '%s\n' 'component:${component}'
    '';
in
{
  moon = pkgs.mkShell {
    packages = [ (marker "special-moon" "special-moon-marker") ];
  };
  bootstrap = pkgs.mkShell {
    packages = [ (marker "special-bootstrap" "special-bootstrap-marker") ];
  };
}
