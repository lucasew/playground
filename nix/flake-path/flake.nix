{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };
  outputs = self: 
  let
    pkgs = import self.nixpkgs {};
    inherit (builtins) concatStringsSep attrValues mapAttrs;
  in {
    devShells.default = pkgs.mkShell {
      shellHook = ''
        ${concatStringsSep ":" (
          attrValues (
            mapAttrs (k: v: "${k}=${v}") self
          )
        )}
      '';
    };
  };
}
