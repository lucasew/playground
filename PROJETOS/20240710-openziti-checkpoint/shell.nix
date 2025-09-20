{ pkgs ? import <nixpkgs> {}}:

let
  package = pkgs.callPackage ./. {};
in

pkgs.mkShell {
  buildInputs = [ package ];
}
