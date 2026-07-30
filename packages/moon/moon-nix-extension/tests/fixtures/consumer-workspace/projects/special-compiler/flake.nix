{
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/b3c092d3c36d91e2f61f3dfb39a159f180a56659";

  outputs = { nixpkgs, ... }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      devShells = forAllSystems (system:
        import ./nix/shells.nix {
          pkgs = nixpkgs.legacyPackages.${system};
        });
    };
}
