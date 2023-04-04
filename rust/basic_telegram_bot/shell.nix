{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = with pkgs; [
    cargo rustc rust-analyzer
    openssl openssl.dev pkg-config
  ];
}
