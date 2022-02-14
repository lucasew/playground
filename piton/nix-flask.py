#!/usr/bin/env nix-shell
#! nix-shell -i python -p python3Packages.flask python3Packages.pythonix

from nix import eval as nix_eval

print(nix_eval("with import <nixpkgs> {}; {hello = ''${hello}''; python3 = ''${python3}'';}"))
print(nix_eval("(import <nixpkgs> {}).pkgsCross"))
