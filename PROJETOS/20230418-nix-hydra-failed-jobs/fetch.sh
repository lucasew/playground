#!/usr/bin/env bash
curl -H 'Accept: text/json' https://hydra.nixos.org/jobset/nixpkgs/trunk/jobs-tab?filter=% > /tmp/jobs.html
