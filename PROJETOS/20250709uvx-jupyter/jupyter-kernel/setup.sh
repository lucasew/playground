#!/usr/bin/env bash
for version in 3.13 3.12; do
  KERNEL_DIR=~/.local/share/jupyter/kernels/uv-$version

  mkdir -p "$KERNEL_DIR"
  UV="$(which uv)"
  UV_DIR="$(dirname "$UV")"

  cat <<EOF > "$KERNEL_DIR"/kernel.json 
  {
    "env": {
      "PATH": "\${PATH}:$UV_DIR"
    },
    "argv": [
      "$UV",
      "run",
      "--python", "$version",
      "--with", "ipykernel",
      "--no-project",
      "--isolated",
      "--refresh",
      "python", "-m", "ipykernel_launcher",
      "-f", "{connection_file}"
    ],
    "display_name": "uv-$version",
    "language": "python",
    "metadata": {
      "debugger": true
    }
  }
EOF
done
