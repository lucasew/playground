{pkgs ? import <nixpkgs> {}}:
rec {
  hello = pkgs.writeShellScriptBin "hello" ''
    echo hello
  '';
  symlink = pkgs.symlinkJoin {
    name = "hello-0.1";
    version = "0.1";
    paths = [
      hello
    ];
  };
}
