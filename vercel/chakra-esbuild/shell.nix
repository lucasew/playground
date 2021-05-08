let
  sources = import ./nix/sources.nix;
  overlay = _: pkgs:
  {
    niv = (import sources.niv {}).niv;
  };
  pkgs = import sources.nixpkgs {
    overlays = [
      overlay
    ];
    config = {};
  };
in pkgs.mkShell {
  buildInputs = with pkgs; [
    nodejs
    yarn
  ];
  shellHook = ''
  '';
}
