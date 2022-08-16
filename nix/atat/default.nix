{ stdenv, ... }:
stdenv.mkDerivation {
  name = "atat-demo";
  personName = "lucas";
  dontUnpack = true;
  installPhase = ''
    substitute ${./teste.txt} $out --subst-var personName
  '';
}
