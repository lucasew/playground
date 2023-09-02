{
  description = "Tool to do dataset annotation for semantic segmentation datsets";

  inputs = {
    nixpkgs.url = "nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
  flake-utils.lib.eachDefaultSystem (system: let
    pkgs = import nixpkgs {
      inherit system;
      config.allowUnfree = true;
    };
  in {
      overlay = import ./nix/overlay.nix;
      shellHook = ''
        PYTHONPATH="$PYTHONPATH:$(pwd)"
      '';
      packages = {
        default = pkgs.python3Packages.callPackage ./package.nix { };
      };
      devShells.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          gnumake
          # dev
          python3Packages.pylsp-mypy
          python3Packages.isort
          python3Packages.black
          python3Packages.mypy
          python3Packages.flake8
          python3Packages.pytest
          # runtime
          python3Packages.feedparser
        ];
      };
    });
}
