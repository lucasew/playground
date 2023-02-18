{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = with pkgs; [
    (python3.withPackages (p: with p; [
      feedparser
    ]))
  ];
  shellHook = ''
    PYTHONPATH=$PYTHONPATH:$(pwd)
  '';
}
