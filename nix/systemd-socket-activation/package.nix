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
          # Service = "demo-server.service";
        };
        unitConfig = {
          BindsTo = [ "demo-server.service" ];
        };
        wantedBy = [ "sockets.target" ];
      };
      systemd.services.demo-server = {
        script = "${python3}/bin/python ${./payload.py} --out-dir $STATE_DIRECTORY";
        unitConfig = {
          Requires = [ "demo-server.socket" ];
          After = [ "network.target" "demo-server.socket" ];
        };
        serviceConfig = {
          StateDirectory = "demo_server";
          Restart = "always";
        };
      };
    };

    client = { pkgs, ... }: {
      environment.systemPackages = [ pkgs.curl ];
    };
  };

  testScript = ''
    start_all()

    # not running at first
    server.succeed("[ ! -f /var/lib/demo_server/init ]")
    server.succeed("[ ! -f /var/lib/demo_server/request0 ]")
    client.sleep(2)

    # wake up
    client.succeed("curl server")
    client.sleep(2)

    # confirm it's started
    server.succeed("[ -f /var/lib/demo_server/init ]")
    server.succeed("[ -f /var/lib/demo_server/request0 ]")
    client.succeed("curl server")

    # confirm that's not starting another instance
    server.succeed("[ -f /var/lib/demo_server/request1 ]")
    client.sleep(10)

    # confirm that's stopped for inactivity
    server.succeed("[ -f /var/lib/demo_server/stop ]")

    # check if the service can be restarted
    client.succeed("curl server")
  '';
})
