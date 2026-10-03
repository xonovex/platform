{ pkgs }:
let
  marker = component: executable:
    pkgs.writeShellScriptBin executable ''
      printf '%s\n' 'component:${component}'
    '';
  shell = component: executable: pkgs.mkShell {
    packages = [ (marker component executable) ];
  };
  nodeCommand = pkgs.writeShellScriptBin "node" ''
    moon-general-marker
    moon-node-marker
    for unexpected in moon-go-marker moon-node20-marker moon-node24-marker moon-postgresql-marker moon-protobuf-marker; do
      if command -v "$unexpected" >/dev/null 2>&1; then
        printf 'unexpected component command: %s\n' "$unexpected" >&2
        exit 1
      fi
    done
  '';
in
{
  general = shell "general" "moon-general-marker";
  node = pkgs.mkShell {
    packages = [ (marker "node" "moon-node-marker") nodeCommand ];
  };
  go = pkgs.mkShell {
    packages = [ (marker "go" "moon-go-marker") (marker "go" "go") ];
  };
  node20 = shell "node20" "moon-node20-marker";
  node24 = shell "node24" "moon-node24-marker";
  postgresql = shell "postgresql" "moon-postgresql-marker";
  protobuf = shell "protobuf" "moon-protobuf-marker";
  broken = throw "deliberate fixture Nix evaluation error for component broken";
}
