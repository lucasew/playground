{ pkgs ? import <nixpkgs> {} }:

pkgs.nixosTest {
  name = "qual-a-senha-do-wifi";

  nodes = {
    vizinho = {};
  };

  testScript = ''
    start_all()
    vizinho.wait_for_unit('network.target')
    vizinho.sleep(3)
    vizinho.succeed("ping -c 4 google.com >&2")
  '';
}
