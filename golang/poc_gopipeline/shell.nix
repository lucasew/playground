with import <nixpkgs> {};
stdenv.mkDerivation {
  name = "environment";
  shellHook = ''
    export PATH=$(pwd)/demo:$PATH
  '';
  buildInputs = [
    go
    gopls
  ];
}
