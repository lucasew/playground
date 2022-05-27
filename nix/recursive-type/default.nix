let
  pkgs = import <nixpkgs> {};
  inherit (pkgs) lib;
  inherit (lib) types;
  T = types.either (types.attrsOf T) types.str;
in T.check
