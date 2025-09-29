{pkgs ? import <nixpkgs> {}}:
pkgs.mkShell {
  buildInputs = with pkgs; [
    java-language-server
    openjdk11-bootstrap
    go
    gopls
  ];
}
