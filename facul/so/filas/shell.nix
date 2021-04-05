{pkgs ? import <nixpkgs> {}}:
pkgs.mkShell {
  builtInputs = with pkgs; [
    gcc
    gdb
  ];
  shellHook = ''
    BIN=out
    function b {
      gcc -Wall -g main.c -o $BIN -lrt
    }
    function r {
      ./$BIN "$@"
    }
    function c {
      rm ./$BIN
    }
    function br {
      b && r "$@"
    }
    function debug {
      b && gdb ./$BIN "$@"
    }
  '';
}
