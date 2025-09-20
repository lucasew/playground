{pkgs ? import <nixpkgs> {}}:
pkgs.mkShell {
  buildInputs = with pkgs; [
    pandoc
  ];
  shellHook = ''
    function run {
      pandoc demo.md  -f markdown -t html --lua-filter filter.lua "$@"
    }
  '';
}
