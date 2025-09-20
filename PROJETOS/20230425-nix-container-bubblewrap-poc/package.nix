{ stdenvNoCC, dockerTools, lib, jq }:

stdenvNoCC.mkDerivation {
  pname = "audiobookshelf";
  version = "22.18";

  dontUnpack = true;

  src = dockerTools.pullImage {
    imageName = "ghcr.io/advplyr/audiobookshelf";
    imageDigest = "sha256:bc9c6819b74fcf93193e8674175e03004cb7bbf6309cd7e8a251f8eebfffe058";
    sha256 = "sha256-Argf/cV6OvPjAoY7Q4aNAON55/LlCcw8AyUAmoZ2T9U=";
    finalImageName = "audiobookshelf";
    finalImageTag = "22.18";
  };

  nativeBuildInputs = [ jq ];

  installPhase = ''
    tar -xvf $src
    mkdir $out -p
    cp $(cat manifest.json | jq '.[0].Config' -r) $out/oci.json
    for tarFile in $(cat manifest.json  | jq '.[0].Layers[]' -r); do
      tar -xvf $tarFile -C $out
    done
  '';
}

