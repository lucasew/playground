{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = with pkgs; [
    texlive.combined.scheme-basic
    cmake
    ninja
    imagemagick
  ];
}
