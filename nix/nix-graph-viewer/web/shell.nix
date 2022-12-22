{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  shellHook = ''
    PATH="$PATH:$(pwd)/node_modules/.bin"
  '';
  buildInputs = with pkgs; [
    nodejs
  ];
}
