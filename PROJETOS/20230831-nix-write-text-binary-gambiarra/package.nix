{ stdenv
, lib
, coreutils
}:

let
  builder = stdenv.mkDerivation {
    name = "writeTest-builder";

    dontUnpack = true;

    buildPhase = ''
      cc ${./builder.c} -o builder
    '';

    installPhase = ''
      install -m 755 builder $out
    '';
  };
in

name: text:

builtins.derivation {
  text = builtins.unsafeDiscardStringContext text;
  inherit name;
  inherit (stdenv) system;

  inherit builder;
  # builder = stdenv.shell;
  # args = [ "-c" "${coreutils}/bin/mv $textPath $out; ${coreutils}/bin/cat $out" ];
  # builder = "${coreutils}/bin/mv";
  # builder = "${coreutils}/bin/ls";
  # args = [ "-lha" "." "${builtins.replaceStrings [ "/" ] [""] (placeholder "textPath")}" "${placeholder "out"}" ];
  # args = [ "/build/.attr-${builtins.replaceStrings [ "/" ] [""] (placeholder "textPath")}" "${placeholder "out"}" ];


  passAsFile = [ "text" ];

  outputHashMode = "flat";
  outputHashAlgo = "sha256";
  # outputHash = "sha256-KHTXgA/sriCET2gmlpILKFNFPNMsR+E2KndiA9AimxQ=";
  outputHash = builtins.hashString "sha256" text;
}
