{ pkgs ? import <nixpkgs> {}}:

pkgs.mkShell {
  buildInputs = with pkgs; [
    cargo
    rustc
    cargo-audit
    cargo-edit
    cargo-outdated
    rust-analyzer
    rustfmt
  ];
}

