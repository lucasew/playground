{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = with pkgs; [
    (python3.withPackages (p: with p; [
      feedparser
    ]))
    python3Packages.pylsp-mypy
  ];
  shellHook = ''
    PYTHONPATH=$PYTHONPATH:$(pwd)
  '';
}
