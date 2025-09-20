#!/usr/bin/env python3

from pathlib import Path
import tempfile
from urllib.request import urlopen, urlretrieve
from argparse import ArgumentParser
import tarfile
import shutil
import subprocess
import sys

parser =ArgumentParser(description="Roda o eval de um flake no nix e entrega a closure de todas as derivations.")
parser.add_argument('workdir', type=Path)
parser.add_argument('ref', type=str)
args = parser.parse_args()

try:
    ref_dir = Path(args.ref.split("#")[0])
    if ref_dir.exists():
        ref_dir = ref_dir.resolve()
    else:
        ref_dir = None
except Exception:
    ref_dir = None

args.workdir.mkdir(parents=True, exist_ok=True)
args.workdir = args.workdir.resolve()

NIX_TAR = args.workdir / "nix.tar.xz"
if not (args.workdir / "nix").exists():
    installer = urlopen("https://nixos.org/nix/install").read().decode('utf-8')
    release_url = None
    for line in installer.split('\n'):
        line = line.strip()
        if line.startswith('url='):
            release_url = line[4:]
    release_url = release_url.replace('$system', 'x86_64-linux')

    print(release_url, file=sys.stderr)
    urlretrieve(release_url, NIX_TAR)

    assert tarfile.is_tarfile(NIX_TAR)
    with tarfile.open(NIX_TAR, 'r|*') as tar:
        tar.extractall(TEMP_STORE_DIR)

    NIX_DIR = args.workdir / "nix"
    NIX_DIR.mkdir(parents=True, exist_ok=True)
    (subdir[0] / "store").rename(NIX_DIR / "store")

TEMP_STORE_DIR = args.workdir / "store"
subdir = list(TEMP_STORE_DIR.iterdir())
assert len(subdir) == 1

NIX_BIN = None
NIX_CACERT = None
for line in (subdir[0] / "install").open('r'):
    line = line.strip()
    if line.startswith('nix="'):
        NIX_BIN = Path(line[5:].replace('"', '')) / "bin" / "nix"
    if line.startswith('cacert="'):
        NIX_CACERT = Path(line[8:].replace('"', ''))

print(NIX_BIN, file=sys.stderr)
print(NIX_CACERT, file=sys.stderr)

systemd_run_flags = [
    "--user",
    "-tP",
    "-p", f"RootDirectory={args.workdir}",
    "-p", "BindReadOnlyPaths=/etc/resolv.conf",
    *(["-p", f"BindReadOnlyPaths={ref_dir}" ]if ref_dir is not None else []),
    "-E", f"NIX_SSL_CERT_FILE={NIX_CACERT}/etc/ssl/certs/ca-bundle.crt",
    NIX_BIN,
    "--extra-experimental-features",
    "nix-command flakes",
]

cmd = [
    "systemd-run",
    *[str(i) for i in systemd_run_flags],
    "eval",
    "--raw",
    args.ref
]
print(cmd, file=sys.stderr)
drvPath = subprocess.run(cmd, stdout=subprocess.PIPE).stdout.decode('utf-8')
print(drvPath, file=sys.stderr)
print(len(drvPath), file=sys.stderr)
print('drvPath', drvPath, file=sys.stderr)

cmd = [
    "systemd-run",
    *[str(i) for i in systemd_run_flags],
    "derivation",
    "show",
    "-r",
    drvPath
]
subprocess.run(cmd)

if NIX_TAR.exists():
    NIX_TAR.unlink()


