{ pkgs ? import <nixpkgs> {
}}:

let

  inherit (pkgs) lib;
  packages = with pkgs.pkgsCross.mingwW64; [
    threadsCross.package
  ];
  builder = let
    ldflags = map (p: "-L${p}/lib") packages;
    cflags = map (p: "-I${p.dev or p}/include") packages;

  in pkgs.runCommand "wingo" { buildInputs = [ pkgs.makeWrapper ];  } ''
    mkdir -p $out/bin
    makeWrapper ${lib.getExe pkgs.go} $out/bin/wingo \
      --set CC ${pkgs.lib.getExe' pkgs.pkgsCross.mingwW64.stdenv.cc "x86_64-w64-mingw32-gcc"} \
      --set CXX ${pkgs.lib.getExe' pkgs.pkgsCross.mingwW64.stdenv.cc "x86_64-w64-mingw32-g++"} \
      --set CGO_CFLAGS "${lib.escapeShellArgs cflags}" \
      --set CGO_CXXFLAGS "${lib.escapeShellArgs cflags}" \
      --set CGO_LDFLAGS "${lib.escapeShellArgs ldflags}" \
      --set CGO_ENABLED 1 \
      --set GOOS windows \
      --set GOARCH amd64 \
  '';
in

pkgs.mkShell {
  packages = with pkgs; [
    builder
    go
    gopls
  ];
}
