{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = with pkgs; [
    python3Packages.grpcio
    python3Packages.grpcio-tools
  ];
}
