{ path, stdenv, python3 }:
let
  queue = "/tmp/queue";
in

import "${path}/nixos/tests/make-test-python.nix" ({pkgs, ...}: {
  name = "limit-template-paralelism";

  nodes = {
    actor = { pkgs, ... }: {
      systemd.tmpfiles.rules = [
        "D ${queue} 777 root root"
      ];
      systemd.services = {
        "demo-unit-runner" = {
          script = ''
            function getNextTask {
              echo "${queue}/$(ls -1 ${queue} | sort -R | head -n 1)"
            }
            while [ ! -d "$(getNextTask)" ]; do
              taskFile="$(getNextTask)"

              echo "runner start $(ls ${queue})" >&2
              if [ -d "$taskFile" ]; then
                echo "runner no tasks" >&2
                exit 0
              fi
              UNIT_TEMPLATE="$(cat "$taskFile")"

              # ========= begin payload

              echo "unit template '$UNIT_TEMPLATE'" >&2
              echo > "$STATE_DIRECTORY/$UNIT_TEMPLATE"
              ls $STATE_DIRECTORY >&2

              sleep 1

              # ======== end payload

              rm "$taskFile"
            done
            echo "finish runner" >&2
          '';
          wantedBy = [ "multi-user.target" ];
          serviceConfig = {
            Restart = "on-failure";
            RestartMaxDelaySec = 120;
            RestartSteps = 3;
            StateDirectory = "demo-unit";
          };
        };
        "demo-unit@" = {
          environment.UNIT_TEMPLATE="%I";

          script = ''
            echo "$UNIT_TEMPLATE" >> ${queue}/$INVOCATION_ID
            systemctl start demo-unit-runner
          '';
        };
      };
    };
  };
  testScript = ''
    # import time

    start_all()
    actor.wait_for_unit("multi-user.target")

    for i in range(5):
      actor.succeed(f"systemctl start demo-unit@{i}")

    last_value = 0
    for i in range(5):
      actor.sleep(1)
      this_value = int(actor.succeed('ls /var/lib/demo-unit -1 | wc -l'))
      if last_value == 0:
        this_value = last_value
        continue
      assert this_value - last_value == 1
      this_value = last_value

    actor.sleep(3)

    for i in range(5, 10):
      actor.succeed(f"systemctl start demo-unit@{i}")

    last_value = 0
    for i in range(5, 10):
      actor.sleep(1)
      this_value = int(actor.succeed('ls /var/lib/demo-unit -1 | wc -l'))
      if last_value == 0:
        this_value = last_value
        continue
      assert this_value - last_value == 1
      this_value = last_value

  '';
})
