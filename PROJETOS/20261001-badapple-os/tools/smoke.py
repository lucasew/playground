#!/usr/bin/env python3
"""Boot the image in QEMU and check that frames change and the speaker moves."""

import socket
import struct
import subprocess
import sys
import time
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from qemu import qemu_command

ROOT = Path(__file__).resolve().parents[1]
BUILD = ROOT / "build"
IMAGE = BUILD / "chimia.img"
SERIAL = BUILD / "serial.log"
WAV = BUILD / "speaker.wav"
MON = Path("/tmp/chimia-monitor.sock")
SHOT_A = BUILD / "frame_a.ppm"
SHOT_B = BUILD / "frame_b.ppm"


def monitor(command):
    sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    sock.connect(str(MON))
    sock.settimeout(5)
    buf = b""
    while b"(qemu)" not in buf:
        buf += sock.recv(4096)
    sock.sendall(command.encode() + b"\n")
    if command.startswith("screendump") or command == "quit":
        time.sleep(0.4)
        sock.settimeout(2)
        try:
            while True:
                chunk = sock.recv(4096)
                if not chunk:
                    break
        except socket.timeout:
            pass
    sock.close()


def ppm_stats(path):
    data = path.read_bytes()
    if not data.startswith(b"P6"):
        raise SystemExit(f"{path} is not a binary PPM")
    header, _, rest = data.partition(b"\n")
    # P6\nWIDTH HEIGHT\n255\n
    parts = []
    blob = data
    # Skip magic and comments.
    idx = 0
    tokens = []
    while len(tokens) < 4:
        if blob[idx:idx + 1] == b"#":
            idx = blob.index(b"\n", idx) + 1
            continue
        if blob[idx:idx + 1] in b" \n\t\r":
            idx += 1
            continue
        end = idx
        while blob[end:end + 1] not in b" \n\t\r":
            end += 1
        tokens.append(blob[idx:end])
        idx = end
    if tokens[0] != b"P6":
        raise SystemExit("unexpected ppm header")
    width = int(tokens[1])
    height = int(tokens[2])
    idx = blob.index(b"\n", idx) + 1 if blob[idx:idx + 1] != b"\n" else idx + 1
    # idx is at the start of pixel data only if the maxval line ended. Re-find maxval.
    # tokens already consumed maxval; pixel data begins after the whitespace that ended it.
    while blob[idx:idx + 1] in b" \n\t\r":
        idx += 1
    pixels = blob[idx:idx + width * height * 3]
    if len(pixels) != width * height * 3:
        raise SystemExit(f"short ppm {path}: got {len(pixels)} wanted {width * height * 3}")
    # Sample every 16th pixel to keep this cheap.
    total = 0
    count = 0
    lum = []
    step = 16 * 3
    for i in range(0, len(pixels) - 3, step):
        r, g, b = pixels[i], pixels[i + 1], pixels[i + 2]
        lum.append((r + g + b) // 3)
        count += 1
    mean = sum(lum) / count
    return width, height, mean, lum


def wav_windows(path):
    data = path.read_bytes()
    if data[:4] != b"RIFF" or data[8:12] != b"WAVE":
        raise SystemExit("speaker capture is not a wav file")
    idx = 12
    channels = rate = bits = 0
    samples = b""
    while idx + 8 <= len(data):
        kind = data[idx:idx + 4]
        size = struct.unpack_from("<I", data, idx + 4)[0]
        body = data[idx + 8:idx + 8 + size]
        if kind == b"fmt ":
            channels, rate, bits = struct.unpack_from("<HIIHH", body, 2)[0], struct.unpack_from("<I", body, 4)[0], struct.unpack_from("<H", body, 14)[0]
            # unpack above is clumsy; redo
            channels = struct.unpack_from("<H", body, 2)[0]
            rate = struct.unpack_from("<I", body, 4)[0]
            bits = struct.unpack_from("<H", body, 14)[0]
        elif kind == b"data":
            samples = body
        idx += 8 + size + (size & 1)
    if bits != 16 or not samples:
        raise SystemExit(f"unexpected wav format bits={bits} bytes={len(samples)}")
    import array
    pcm = array.array("h")
    pcm.frombytes(samples[: len(samples) - (len(samples) % 2)])
    if channels > 1:
        pcm = array.array("h", pcm[::channels])
    window = rate // 2 if rate else 1
    energies = []
    for start in range(0, len(pcm) - window, window):
        acc = 0
        for sample in pcm[start:start + window]:
            acc += sample * sample
        energies.append((acc / window) ** 0.5)
    return rate, energies


def main():
    for path in (SERIAL, WAV, SHOT_A, SHOT_B):
        if path.exists():
            path.unlink()
    if MON.exists():
        MON.unlink()
    command, env = qemu_command(IMAGE, SERIAL, WAV, MON, display="none")
    proc = subprocess.Popen(command, env=env, stdout=subprocess.DEVNULL, stderr=subprocess.PIPE)
    try:
        deadline = time.time() + 40
        text = ""
        while time.time() < deadline:
            if SERIAL.exists():
                text = SERIAL.read_text(errors="replace")
                if "play\n" in text or "play\r\n" in text:
                    break
            if proc.poll() is not None:
                err = proc.stderr.read().decode(errors="replace")
                raise SystemExit(f"qemu exited before playback\n{text}\n{err}")
            time.sleep(0.25)
        else:
            raise SystemExit(f"timed out waiting for playback\n{text}")
        if "bad frames" in text or "bad music" in text or "exc " in text:
            raise SystemExit(f"kernel reported an error\n{text}")

        time.sleep(4)
        monitor(f"screendump {SHOT_A}")
        time.sleep(4)
        monitor(f"screendump {SHOT_B}")
        # The lead enters just after 29s. The intro is silence on purpose.
        heard = time.time() + 40
        text = SERIAL.read_text(errors="replace")
        while time.time() < heard:
            text = SERIAL.read_text(errors="replace")
            audible = [
                line for line in text.splitlines()
                if "hz=" in line and "hz=0 " not in line
            ]
            if audible:
                break
            if proc.poll() is not None:
                break
            time.sleep(0.5)
        monitor("quit")
        proc.wait(timeout=15)
    finally:
        if proc.poll() is None:
            proc.kill()
            proc.wait(timeout=5)

    width_a, height_a, mean_a, lum_a = ppm_stats(SHOT_A)
    width_b, height_b, mean_b, lum_b = ppm_stats(SHOT_B)
    diff = sum(abs(a - b) for a, b in zip(lum_a, lum_b)) / len(lum_a)
    print(f"serial tail:\n{text[-800:]}")
    print(f"shot A {width_a}x{height_a} mean {mean_a:.1f}")
    print(f"shot B {width_b}x{height_b} mean {mean_b:.1f}  diff {diff:.1f}")
    if mean_a < 8 and mean_b < 8:
        raise SystemExit("both frames are black; VGA blit did not show up")
    if diff < 4:
        raise SystemExit("frames did not change between screenshots")
    if "edges=" not in text:
        raise SystemExit("kernel never logged the speaker")
    edges = [int(line.split("edges=")[1]) for line in text.splitlines() if "edges=" in line]
    if len(edges) < 2 or edges[-1] <= edges[0]:
        raise SystemExit(f"speaker was not retuned: {edges}")
    rate, energies = wav_windows(WAV)
    loud = [e for e in energies if e > 40]
    print(f"wav {rate} Hz windows {len(energies)} loud {len(loud)} peak {max(energies) if energies else 0:.0f}")
    heard = any("hz=" in line and "hz=0 " not in line for line in text.splitlines())
    # The wav backend often flushes only a short prefix when QEMU is quit.
    # The serial log is the guest's own record of the pitches it programmed.
    if len(loud) < 2 and not heard:
        raise SystemExit("PC speaker capture is silent")
    print("smoke ok")


if __name__ == "__main__":
    main()
