{pkgs ? import <nixpkgs> {}}:
pkgs.mkShell {
  builtInputs = with pkgs; [
    gcc
    gdb
  ];
  shellHook = ''
    BIN=out
    function b {
      gcc -Wall -g main.c -o $BIN -lrt -lpthread -O0 -D_FORTIFY_SOURCE=0
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
