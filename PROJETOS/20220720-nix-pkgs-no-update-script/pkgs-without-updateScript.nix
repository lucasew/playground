{ pkgs ? import <nixpkgs> {
    allowUnfree = true;
    allowBroken = true;
    allowInsecure = true;
    acceptAndroidNdkLicense = true;
    overlays = [
      (self: super: {
        AAAAAASomeThingsFailToEvaluate = null;
      })
    ];
  }
}:
let
  inherit (pkgs) lib;
  recur = path: node: 
    let
      isDrv = lib.isDerivation node;
      isAttrs = lib.isAttrs node;
      trace = (node.updateScript or node.passthru.updateScript or null) != null;
      tryEvalTrace = builtins.tryEval trace;
      retTrace = if tryEvalTrace.success then tryEvalTrace.value else null;

      mappedAttrs = builtins.mapAttrs (k: v: recur "${path}.${k}" v) node;
      evaluated = if isDrv then retTrace else if isAttrs then mappedAttrs else null;
    in evaluated;
in recur "" pkgs
