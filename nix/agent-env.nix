# agent-env is the one declarative core both sandbox surfaces resolve from.
#
# Given an `agent` (sourced from llm-agents.packages.<system>.<name>) plus a base
# package set, it produces two outputs from the SAME closure:
#   - devShell : a mkShell the CLI resolves (`nix print-dev-env`) at run time and
#                bind-mounts read-only into the sandbox;
#   - image    : a streamLayeredImage the operator builds — the same closure baked
#                into TAR layers (verify parity with `nix path-info -r`, not bytes).
#
# Source -> output mapping (the NixSource kinds the shared Go types describe):
#   NixSourcePackages     -> a rev-pinned devShell/image, `agent` drawn from a
#                            rev-pinned package set (this function).
#   NixSourceProjectFlake -> the project's own flake; CLI only — the operator
#                            always builds the image from the synthesized env.
{ pkgs }:
let
  # The default base toolchain (mirrors shared-agent-go nix.DefaultPackages).
  defaultBasePackages = with pkgs; [
    nodejs_24
    git
    ripgrep
    fd
    fzf
    jq
    curl
    coreutils
    bash
  ];
  mkAgentImage = import ./mkAgentImage.nix { inherit pkgs; };
in
{ agent
, packages ? defaultBasePackages
, extraPackages ? [ ]
, name ? "agent"
}:
{
  # The CLI resolves this devShell to a content-pinned closure.
  devShell = pkgs.mkShell {
    packages = [ agent ] ++ packages ++ extraPackages;
  };

  # The operator builds this image from the same closure.
  image = mkAgentImage {
    inherit agent packages extraPackages name;
  };
}
