{ pkgs ? import <nixpkgs> {config.allowUnfree = true;} }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    python3Packages.bentoml
    python3Packages.opencv4
    python3Packages.numpy
    python3Packages.fastapi
    python3Packages.pillow
  ];
}
