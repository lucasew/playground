{ pkgs ? import <nixpkgs> {}}:

{
  darwin = pkgs.callPackage ./darwin.nix {};
}
