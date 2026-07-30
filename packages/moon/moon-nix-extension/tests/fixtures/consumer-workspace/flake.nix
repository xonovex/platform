{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/b3c092d3c36d91e2f61f3dfb39a159f180a56659";

  outputs = { nixpkgs, ... }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
      componentsFor = system: import ./nix/components.nix {
        pkgs = nixpkgs.legacyPackages.${system};
      };
      mkMoonShell = system: names:
        let
          pkgs = nixpkgs.legacyPackages.${system};
          components = componentsFor system;
          missing = builtins.filter (name: !(builtins.hasAttr name components)) names;
        in
        if missing != [ ] then
          throw "unknown fixture component(s): ${builtins.concatStringsSep ", " missing}"
        else
          pkgs.mkShell {
            inputsFrom = map (name: components.${name}) names;
          };
    in
    {
      lib.mkMoonShell = mkMoonShell;
      devShells = forAllSystems (system: {
        default = mkMoonShell system [ "general" ];
      });
    };
}
