{
  inputs.nixpkgs.url = "nixpkgs";
  outputs = { nixpkgs, ... }: let
    system = "x86_64-linux";
    pkgs = import nixpkgs { inherit system; };
  in  {
      packages.${system}.default = pkgs.stdenvNoCC.mkDerivation {
        name = "test";
        dontUnpack = true;
        installPhase = ''
          mkdir -p $out
          ln -s /etc/passwd $out/demo
        '';
      };
  };
}
