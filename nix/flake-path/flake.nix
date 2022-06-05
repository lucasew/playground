{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };
  outputs = {nixpkgs, ...} @ self: 
  let
    pkgs = import nixpkgs { inherit system; };
    inherit (builtins) concatStringsSep attrValues mapAttrs;
    system = "x86_64-linux";
  in {
    devShells.${system}.default = pkgs.mkShell {
      shellHook = ''
        NIX_PATH=${concatStringsSep ":" (
          attrValues (
            mapAttrs (k: v: "${k}=${v}") self
          )
        )}
      '';
    };
  };
}
