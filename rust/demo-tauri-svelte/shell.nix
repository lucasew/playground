{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  buildInputs = with pkgs; [
    nodejs cargo rustc
    rust-analyzer
    pkg-config glib glib-networking gtk3 libsoup webkitgtk
    (pkgs.stdenvNoCC.mkDerivation {
      name = "gwrap";
      dontUnpack = true;
      preferLocalBuild = true;
      nativeBuildInputs = [ pkgs.wrapGAppsHook ]; 
      installPhase = ''
        mkdir $out/bin -p
        makeWrapper ${pkgs.writeShellScript "run" ''"$@"''} \
          $out/bin/gwrap \
          ${"$"}{gappsWrapperArgs[@]}
      '';
    })
  ];
}
