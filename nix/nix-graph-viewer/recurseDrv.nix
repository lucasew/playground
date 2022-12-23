# BROKEN
let
  pkgs = import <nixpkgs> {};
in pkgs.callPackage (
  { lib
  }:
  let
    inherit (builtins) concatStringsSep mapAttrs tryEval length match trace;
    inherit (lib) isAttrs isDerivation head reverseList filter;
    collectDrvs = key: attrs:
      if  length key > 2 then
        null
      else if isDerivation attrs then
        (
          let
            eval = tryEval attrs.drvPath;
          in if eval.success then
            (trace "${concatStringsSep "." key} ${eval.value}" null)
          else null
          )
      else if isAttrs attrs then
        (mapAttrs (k: v: if (length (filter (fk: fk == k || ((match ".*[dD]arwin.*" (trace fk fk)) != null)) key) == 0) then (collectDrvs (key ++ [k]) v) else null) attrs)
      else null;
  in collectDrvs [ ]  pkgs
) {}
