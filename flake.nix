{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    nixShells = {
      url = "path:./nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
    # Agent packaging (claude-code, codex, …). Packaging ONLY — isolation is out
    # of scope upstream. Pinned by flake.lock to a rev; consumers pin via the
    # lock alone. Do NOT set `inputs.nixpkgs.follows` against it — that breaks its
    # binary-cache hits. Substituting prebuilt agents from cache.numtide.com is a
    # deliberate TRUST EXPANSION (key niks3.numtide.com-1:DTx8wZduET09hRmMtKdQDxNNthLQETkc/yaX7M4qK0g=);
    # we do not enable it here, so agents build from source.
    llm-agents.url = "github:numtide/llm-agents.nix";
  };

  outputs = { nixpkgs, nixShells, llm-agents, ... }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      mkMoonShell = system: names:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          shells = nixShells.devShells.${system};
          missingNames = builtins.filter (name: !(builtins.hasAttr name shells)) names;
        in
        if missingNames != [ ] then
          throw "unknown Moon Nix environment component(s): ${builtins.concatStringsSep ", " missingNames}"
        else
          pkgs.mkShell {
            inputsFrom = map (name: shells.${name}) names;
            # mkMoonShell disables the shared evaluation cache because Moon may
            # enter several dev shells concurrently.
            NIX_CONFIG = "eval-cache = false";
          };

      # The agent image (dockerTools) is Linux-only; scope the smoke outputs that
      # realize it to Linux systems.
      linuxSystems = [ "x86_64-linux" "aarch64-linux" ];
      forLinux = nixpkgs.lib.genAttrs linuxSystems;

      # A lightweight agent env (hello stands in for an llm-agents agent) that
      # exercises agent-env.nix + mkAgentImage.nix end to end without a heavy
      # build — the smoke fixture for both the devShell and the image.
      agentTestEnv = system:
        let pkgs = nixpkgs.legacyPackages.${system};
        in (import ./nix/agent-env.nix { inherit pkgs; }) {
          agent = pkgs.hello;
          packages = [ pkgs.coreutils ];
          name = "agent-test";
        };
    in
    {
      lib.mkMoonShell = mkMoonShell;

      devShells = forAllSystems (system: {
        # Full shell — composed from the shared per-tool devShells in nix/.
        # The docker component supplies hadolint, which `docker-lint` (folded into
        # ci-check via the docker tag's `lint` alias) needs on PATH.
        default = mkMoonShell system [ "node" "go" "k8s" "shell" "rust" "release" "ci" "docker" "general" ];

        # Lean per-purpose shells, selected via the nix toolchain `shellByTag` setting.
        ci = mkMoonShell system [ "ci" "general" ];
        docker = mkMoonShell system [ "docker" "general" ];
        go = mkMoonShell system [ "go" "general" ];
        shell = mkMoonShell system [ "shell" "general" ];
        rust = mkMoonShell system [ "rust" "release" "general" ];

        # Smoke: agent-env.nix resolves a devShell from one closure.
        agentEnvTest = (agentTestEnv system).devShell;
      });

      packages = forLinux (system: {
        # Smoke: mkAgentImage.nix builds a streamLayeredImage from the closure.
        agentImageTest = (agentTestEnv system).image;
      });

      checks = forLinux (system:
        let pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          # `nix flake check` builds the test image.
          agentImageTest = (agentTestEnv system).image;

          # Locks llm-agents.nix into flake.lock and eval-checks the
          # packages.<system>.<name> attr path WITHOUT building a heavy agent
          # (reads `.name` only).
          llmAgentsPinned = pkgs.runCommand "llm-agents-pinned"
            {
              rev = llm-agents.rev or "unlocked";
              agentName = llm-agents.packages.${system}.claude-code.name;
            } ''
            printf '%s @ %s\n' "$agentName" "$rev" > "$out"
          '';
        }
      );

      # The real agent image the operator references (claude-code + base
      # toolchain), built from the SAME pinned flake.lock + agent-env.nix as the
      # CLI provisions — the same content-addressed store-path closure (compare
      # `nix path-info -r`, not layer bytes). It lives under legacyPackages so
      # `nix flake check` does not build the heavy agent closure; CI builds it on
      # demand (`nix build .#legacyPackages.<system>.agentImage`) and pushes it
      # digest-pinned (`created` defaults to the epoch — never `now`).
      legacyPackages = forLinux (system:
        let pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          agentImage = ((import ./nix/agent-env.nix { inherit pkgs; }) {
            agent = llm-agents.packages.${system}.claude-code;
            name = "agent";
          }).image;
        }
      );
    };
}
