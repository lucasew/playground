{pkgs ? import <nixpkgs> {}}:
pkgs.stdenv.mkDerivation {
  name = "poctpoctpoct";
  phases = ["installPhase"];
  outputHash = "sha256-W2GwwgMrSqlRnWXMmMZBbBJBXgLH+7qhvlEh3HUWLts=";
  buildInputs = with pkgs; [
    curl
    cacert
  ];
  installPhase = "curl https://google.com > $out";
}
