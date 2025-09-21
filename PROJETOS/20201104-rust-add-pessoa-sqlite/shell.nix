# https://marketplace.visualstudio.com/items?itemName=arrterian.nix-env-selector

{
  pkgs ? import <nixpkgs> {}
}: pkgs.stdenv.mkDerivation {
  name = "rust-workspace";
  buildInputs = [
    pkgs.cargo
    pkgs.rls
    pkgs.rustc
  ];
}
