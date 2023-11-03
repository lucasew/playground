{ pkgs ? import <nixpkgs> { config.allowUnfree = true; }}:

pkgs.mkShell {
  buildInputs = with pkgs; [ cudatoolkit ];
}
