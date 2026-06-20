{ pkgs }:
{
  # cargo/clippy/rustfmt + wasm-opt + wasm tools. Requires the rust-overlay
  # (applied in nix/flake.nix) so pkgs.rust-bin is available.
  packages = [
    (pkgs.rust-bin.stable.latest.default.override {
      targets = [ "wasm32-wasip1" ];
    })
    pkgs.binaryen
    pkgs.wabt
  ];
}
