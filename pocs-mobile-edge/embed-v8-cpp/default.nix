{ pkgs ? import <nixpkgs> {}
, ...}:
pkgs.stdenv.mkDerivation {
  name = "v8-test";
  dontUnpack = true;
  nativeBuildInputs = with pkgs; [ v8 ];
  installPhase = ''
    mkdir -p $out/bin
    g++ ${./code.cc} -o $out/bin/teste -lv8 -pthread -std=c++14 -DV8_COMPRESS_POINTERS
  '';
}
