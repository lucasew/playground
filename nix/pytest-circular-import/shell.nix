{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = [
    (pkgs.python3Packages.callPackage ./package.nix { })
  ];
}
