#!/usr/bin/env bash

cd ./cloned

function clone {
  git clone --depth 1 "$@"
}

clone https://github.com/anomalyco/opencode
clone https://github.com/anthropics/claude-code
clone https://github.com/charmbracelet/crush
