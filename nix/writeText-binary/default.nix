{ pkgs ? import <nixpkgs> {} }:

rec {
  writeText = pkgs.callPackage ./package.nix { };

  tests = {
    basic = writeText "text" "teste";
    withDrv = writeText "text" "teste-${pkgs.bash.outPath}";
  };
}
