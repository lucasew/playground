#!/usr/bin/env bash
set -euo pipefail

# Validation harness for semantic level A:
# - Deterministic generation by seed
# - Cross-compiler behavioral consistency (gcc vs clang)
# - Runtime UB signal via sanitizers (default: UBSan)

COUNT="${COUNT:-20}"
SEED_START="${SEED_START:-1}"
TIMEOUT_SECS="${TIMEOUT_SECS:-3}"
WORKDIR="${WORKDIR:-/tmp/csmith-go-validate}"
RUNTIME_INCLUDE="${RUNTIME_INCLUDE:-/nix/store/hrf9nixgjz33q1563l9bxx155py477qv-csmith-2.3.0/include/csmith-2.3.0}"
CANDIDATE_CMD="${CANDIDATE_CMD:-GOCACHE=/tmp/go-cache go run ./cmd/csmith-go}"
STRICT_SAN="${STRICT_SAN:-0}"
SAN_FLAGS="${SAN_FLAGS:--fsanitize=undefined -fno-sanitize-recover=all}"

usage() {
  cat <<USAGE
Usage: scripts/validate-a.sh [--count N] [--seed-start N] [--timeout N] [--workdir PATH]

Env overrides:
  COUNT, SEED_START, TIMEOUT_SECS, WORKDIR, RUNTIME_INCLUDE, CANDIDATE_CMD
  SAN_FLAGS='-fsanitize=undefined -fno-sanitize-recover=all'
  STRICT_SAN=1   fail if sanitizer toolchain is unavailable

Examples:
  scripts/validate-a.sh --count 50 --seed-start 1000
  CANDIDATE_CMD='./bin/csmith-go' scripts/validate-a.sh
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --count)
      COUNT="$2"; shift 2 ;;
    --seed-start)
      SEED_START="$2"; shift 2 ;;
    --timeout)
      TIMEOUT_SECS="$2"; shift 2 ;;
    --workdir)
      WORKDIR="$2"; shift 2 ;;
    -h|--help)
      usage; exit 0 ;;
    *)
      echo "unknown argument: $1" >&2
      usage
      exit 2 ;;
  esac
done

if ! command -v gcc >/dev/null 2>&1; then
  echo "gcc not found" >&2
  exit 1
fi
if ! command -v clang >/dev/null 2>&1; then
  echo "clang not found" >&2
  exit 1
fi
if ! command -v timeout >/dev/null 2>&1; then
  echo "timeout command not found" >&2
  exit 1
fi

mkdir -p "$WORKDIR"

pass=0
fail=0

read -r -a SAN_ARR <<< "$SAN_FLAGS"

probe_src="${WORKDIR}/_san_probe.c"
echo 'int main(void){return 0;}' > "$probe_src"

gcc_san_enabled=1
clang_san_enabled=1

if ! gcc -std=c11 -O1 "${SAN_ARR[@]}" "$probe_src" -o "${WORKDIR}/_san_probe.gcc" >/dev/null 2>&1; then
  gcc_san_enabled=0
fi
if ! clang -std=c11 -O1 "${SAN_ARR[@]}" "$probe_src" -o "${WORKDIR}/_san_probe.clang" >/dev/null 2>&1; then
  clang_san_enabled=0
fi

if [[ "$gcc_san_enabled" -eq 0 ]]; then
  echo "[WARN] gcc sanitizer toolchain unavailable for SAN_FLAGS='$SAN_FLAGS'; skipping gcc sanitizer checks"
fi
if [[ "$clang_san_enabled" -eq 0 ]]; then
  echo "[WARN] clang sanitizer toolchain unavailable for SAN_FLAGS='$SAN_FLAGS'; skipping clang sanitizer checks"
fi

if [[ "$STRICT_SAN" == "1" && ( "$gcc_san_enabled" -eq 0 || "$clang_san_enabled" -eq 0 ) ]]; then
  echo "[FAIL] strict sanitizer mode requires both gcc and clang sanitizer toolchains"
  exit 1
fi

for ((i=0; i<COUNT; i++)); do
  seed=$((SEED_START + i))
  base="$WORKDIR/seed_${seed}"
  src="${base}.c"

  eval "$CANDIDATE_CMD --seed ${seed}" > "$src"

  gcc_o0="${base}.gcc.o0"
  clang_o0="${base}.clang.o0"
  gcc_san="${base}.gcc.san"
  clang_san="${base}.clang.san"

  gcc -std=c11 -O0 "$src" -I"$RUNTIME_INCLUDE" -o "$gcc_o0"
  clang -std=c11 -O0 "$src" -I"$RUNTIME_INCLUDE" -o "$clang_o0"

  out_gcc="${base}.gcc.out"
  out_clang="${base}.clang.out"

  timeout "$TIMEOUT_SECS" "$gcc_o0" > "$out_gcc"
  timeout "$TIMEOUT_SECS" "$clang_o0" > "$out_clang"

  if ! cmp -s "$out_gcc" "$out_clang"; then
    echo "[FAIL] seed=${seed}: gcc/clang output mismatch"
    fail=$((fail + 1))
    continue
  fi

  if [[ "$gcc_san_enabled" -eq 1 ]]; then
    if ! gcc -std=c11 -O1 "${SAN_ARR[@]}" "$src" -I"$RUNTIME_INCLUDE" -o "$gcc_san"; then
      echo "[FAIL] seed=${seed}: gcc sanitizer build failed"
      fail=$((fail + 1))
      continue
    fi
    if ! timeout "$TIMEOUT_SECS" "$gcc_san" > /dev/null; then
      echo "[FAIL] seed=${seed}: gcc sanitizer runtime failure"
      fail=$((fail + 1))
      continue
    fi
  fi

  if [[ "$clang_san_enabled" -eq 1 ]]; then
    if ! clang -std=c11 -O1 "${SAN_ARR[@]}" "$src" -I"$RUNTIME_INCLUDE" -o "$clang_san"; then
      echo "[FAIL] seed=${seed}: clang sanitizer build failed"
      fail=$((fail + 1))
      continue
    fi
    if ! timeout "$TIMEOUT_SECS" "$clang_san" > /dev/null; then
      echo "[FAIL] seed=${seed}: clang sanitizer runtime failure"
      fail=$((fail + 1))
      continue
    fi
  fi

  echo "[PASS] seed=${seed}"
  pass=$((pass + 1))
done

echo ""
echo "Validation summary: pass=${pass} fail=${fail} total=${COUNT}"

if [[ "$fail" -ne 0 ]]; then
  exit 1
fi
