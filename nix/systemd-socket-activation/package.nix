{ path, stdenv, python3 }:
let

in

import "${path}/nixos/tests/make-test-python.nix" ({pkgs, ...}: {
  name = "socket-activation";

  nodes = {
    server = { pkgs, ... }: {
      networking.firewall.allowedTCPPorts = [ 80 ];
      systemd.sockets.demo-server = {
        socketConfig = {
          ListenStream = 80;
          Service = "demo-server.service";
        };
        wantedBy = [ "sockets.target" ];
      };
      systemd.services.demo-server = {
        path = [ python3 ];
        serviceConfig = {
          StateDirectory = "demo_server";
        };
        script = ''
          python ${./payload.py} --out-dir $STATE_DIRECTORY
        '';
      };
    };

    client = { pkgs, ... }: {
      environment.systemPackages = [ pkgs.curl ];
    };
  };

  testScript = ''
    start_all()
    server.succeed("[ ! -f /var/lib/demo_server/init ]")
    server.succeed("[ ! -f /var/lib/demo_server/request0 ]")
    client.sleep(2)
    client.succeed("curl server")
    client.sleep(2)
    server.succeed("[ -f /var/lib/demo_server/init ]")
    server.succeed("[ -f /var/lib/demo_server/request0 ]")
    client.succeed("curl server")
    server.succeed("[ -f /var/lib/demo_server/request1 ]")
  '';
})
