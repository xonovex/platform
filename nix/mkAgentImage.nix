# mkAgentImage builds a non-root agent OCI image with streamLayeredImage.
#
# Adapted from nothingnesses/agent-images (lib/mkAgentImage.nix), which is
# bus-factor 1 and uses buildLayeredImage with a named user. This vendored fork
# reworks the layout onto streamLayeredImage (which never realizes a store
# tarball) and a hand-written passwd/group keyed on NUMERIC uid 1000, so the
# image works under `User = "1000:1000"` without relying on a named user.
#
# The operator builds this image; the CLI resolves the same closure as a
# devShell (see agent-env.nix). Both surfaces share one content-addressed
# store-path closure — compare with `nix path-info -r`, not layer bytes.
{ pkgs }:
{ agent
, packages ? [ ]
, extraPackages ? [ ]
, name ? "agent"
, tag ? "latest"
}:
let
  # /etc/passwd + /etc/group for uid/gid 1000. Owned by root (read-only at
  # runtime); the agent runs numerically so the entry is convenience, not a
  # dependency.
  passwdFile = pkgs.writeText "passwd" ''
    root:x:0:0:root:/root:/bin/sh
    agent:x:1000:1000:agent:/home/agent:/bin/sh
    nobody:x:65534:65534:nobody:/nonexistent:/bin/sh
  '';
  groupFile = pkgs.writeText "group" ''
    root:x:0:
    agent:x:1000:
    nobody:x:65534:
  '';
  etcLayer = pkgs.runCommand "agent-etc" { } ''
    mkdir -p $out/etc
    cp ${passwdFile} $out/etc/passwd
    cp ${groupFile} $out/etc/group
  '';
in
pkgs.dockerTools.streamLayeredImage {
  inherit name tag;
  # 128 sits at overlay2's modern ceiling (zero headroom); 100 is the safe
  # conventional choice that still leaves room for the runtime's own layers.
  maxLayers = 100;
  contents = [ agent etcLayer ] ++ packages ++ extraPackages;

  # Pre-create the XDG dirs + /workspace owned by uid 1000. fakeRootCommands
  # runs under fakeroot with the assembled image root as the working directory,
  # so chown records 1000:1000 ownership in the layer.
  fakeRootCommands = ''
    mkdir -p \
      ./home/agent/.config \
      ./home/agent/.cache \
      ./home/agent/.local/state \
      ./home/agent/.local/share \
      ./workspace
    chown -R 1000:1000 ./home/agent ./workspace
  '';

  # NEVER set created = "now": dockerTools defaults to the epoch, which is the
  # reproducibility invariant — two builds of the same closure stream an
  # identical image digest.
  config = {
    User = "1000:1000";
    WorkingDir = "/workspace";
    Env = [
      "HOME=/home/agent"
      "PATH=/bin:/usr/bin"
      "XDG_CONFIG_HOME=/home/agent/.config"
      "XDG_CACHE_HOME=/home/agent/.cache"
      "XDG_STATE_HOME=/home/agent/.local/state"
      "XDG_DATA_HOME=/home/agent/.local/share"
    ];
  };
}
