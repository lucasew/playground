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
        };
        unitConfig = {
          OnFailure = [ "demo-server.socket" ];
        };
        wantedBy = [ "sockets.target" "multi-user.target" ];
      };
      systemd.services.demo-server = {
        script = ''
          echo systemd service start >&2
          ${python3}/bin/python ${./payload.py} --out-dir $STATE_DIRECTORY
          echo systemd service stop >&2
        '';
        wantedBy = [ "multi-user.target" ];
        unitConfig = {
          After = [ "network.target"  ];
        };
        serviceConfig = {
          StateDirectory = "demo_server";
        };
      };
    };

    client = { pkgs, ... }: {
      environment.systemPackages = [ pkgs.curl ];
    };
  };

  testScript = ''
    start_all()

    log.info("Showing units content")
    server.succeed("systemctl cat demo-server.service >&2")
    server.succeed("systemctl cat demo-server.socket >&2")

    log.info("Check if the service is not running at first")
    server.succeed("[ ! -f /var/lib/demo_server/init ]")
    server.succeed("[ ! -f /var/lib/demo_server/request0 ]")
    client.sleep(2)

    log.info("Wake the service up")
    client.succeed("curl server")
    client.sleep(2)

    log.info("Check if the unit actually started and achieved the route handler")
    server.succeed("[ -f /var/lib/demo_server/init ]")
    server.succeed("[ -f /var/lib/demo_server/request0 ]")

    log.info("Send request to the service again to check if systemd is not spawning another instance")
    client.succeed("curl server")

    log.info("Check if systemd hasn't started another instance")
    server.succeed("[ -f /var/lib/demo_server/request1 ]")

    log.info("Wait for the auto shutdown logic of the service to stop it")
    client.sleep(10)

    log.info("Confirm that the service actually stopped")
    server.succeed("[ -f /var/lib/demo_server/stop ]")

    log.info("Check if the service can wake up again from that exact same socket")
    client.succeed("curl server")
  '';
})
