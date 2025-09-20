{ stdenv
, lib
}:

stdenv.mkDerivation {
  name = "outputtag";

  dontUnpack = true;

  buildPhase = ''
    runHook preBuild
    cc ${./O.c} -o O
    runHook postBuild
  '';

  installPhase = ''
    runHook preInstall
    install -D -m755 O $out/bin/O
    runHook postInstall
  '';

  meta = {
    description = "Simple utility that runs a binary with arguments but adds a prefix tag for each line";
    maintainers = [ lib.maintainers.lucasew ];
    license = [ lib.licenses.mit ];
  };
}
