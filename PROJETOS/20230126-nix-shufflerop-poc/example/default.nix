{ stdenv }:

stdenv.mkDerivation {
  name = "ropshuffle-example";
  src = ./.;
  buildPhase = ''
    cc -c *.c
  '';
  installPhase = ''
    mkdir $out/obj/{lib,bin}/summer -p
    install *.o $out/obj/lib/summer
    install *.o $out/obj/bin/summer
    rm $out/obj/lib/summer/main.o
    touch $out/teste
    mkdir -p $out/testepasta
  '';
}
