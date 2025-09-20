{ fetchurl
, stdenv
, lib
}:

stdenv.mkDerivation {
  pname = "darwin";
  version = "0.013";

  dontUnpack = true;

  src = fetchurl {
    url = "https://github.com/OUIsolutions/Darwin/releases/download/0.013/darwin013.c";
    hash = "sha256-XONTYFDpaWIZIf65CrtChWQd1rqatKsfZV5wqFBSZ4k=";
  };

  buildPhase = ''
    runHook preBuild
    cc $src -o darwin
    runHook postBuild
  '';

  installPhase = ''
    runHook preInstall
    install -m 755 -D darwin $out/bin/darwin
    runHook postInstall
  '';
}
