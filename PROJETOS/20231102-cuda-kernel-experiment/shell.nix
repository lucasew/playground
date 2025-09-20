{ pkgs ? import <nixpkgs> { config.allowUnfree = true; }}:

pkgs.mkShell {
  buildInputs = with pkgs; [
    cudatoolkit
    rocmPackages.clr
    rocmPackages.hipify
    clang
  ];
}
