#!/usr/bin/env bash
set -euo pipefail

COUNT="${COUNT:-10}"
SEED_START="${SEED_START:-1}"
TIMEOUT_SECS="${TIMEOUT_SECS:-5}"
WORKDIR="${WORKDIR:-/tmp/csmith-parity}"
UPSTREAM_GEN_CMD="${UPSTREAM_GEN_CMD:-csmith}"
GO_GEN_CMD="${GO_GEN_CMD:-GOCACHE=/tmp/go-cache go run ./cmd/csmith-go}"
UPSTREAM_INCLUDE="${UPSTREAM_INCLUDE:-/nix/store/hrf9nixgjz33q1563l9bxx155py477qv-csmith-2.3.0/include/csmith-2.3.0}"
GO_INCLUDE="${GO_INCLUDE:-$UPSTREAM_INCLUDE}"
CC_BIN="${CC_BIN:-cc}"

usage() {
  cat <<USAGE
Usage: scripts/compare-upstream.sh [--count N] [--seed-start N] [--timeout N] [--workdir PATH]

Env overrides:
  COUNT, SEED_START, TIMEOUT_SECS, WORKDIR
  UPSTREAM_GEN_CMD (default: csmith)
  GO_GEN_CMD (default: GOCACHE=/tmp/go-cache go run ./cmd/csmith-go)
  UPSTREAM_INCLUDE
  GO_INCLUDE (default: UPSTREAM_INCLUDE)
  CC_BIN

Examples:
  scripts/compare-upstream.sh --count 20 --seed-start 100
  UPSTREAM_GEN_CMD='/tmp/csmith/src/csmith' scripts/compare-upstream.sh
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

if ! command -v "$CC_BIN" >/dev/null 2>&1; then
  echo "compiler not found: $CC_BIN" >&2
  exit 1
fi
if ! command -v timeout >/dev/null 2>&1; then
  echo "timeout command not found" >&2
  exit 1
fi

mkdir -p "$WORKDIR"

extract_checksum_token() {
  local file="$1"
  local line
  line=$(grep -E "checksum = " "$file" | tail -n 1 || true)
  if [[ -z "$line" ]]; then
    return 1
  fi
  line=${line#*checksum = }
  echo "$line" | tr -d ' \t\r\n'
}

normalize_checksum_u32() {
  local raw="$1"
  if [[ "$raw" =~ ^0[xX][0-9A-Fa-f]+$ ]]; then
    printf "%u" "$((raw))"
    return
  fi

  if [[ "$raw" =~ ^[0-9]+$ ]]; then
    printf "%u" "$raw"
    return
  fi

  if [[ "$raw" =~ ^[0-9A-Fa-f]+$ ]]; then
    printf "%u" "$((16#$raw))"
    return
  fi

  return 1
}

pass=0
fail=0

for ((i=0; i<COUNT; i++)); do
  seed=$((SEED_START + i))
  base="$WORKDIR/seed_${seed}"
  up_c="${base}.up.c"
  go_c="${base}.go.c"
  up_bin="${base}.up.bin"
  go_bin="${base}.go.bin"
  up_out="${base}.up.out"
  go_out="${base}.go.out"

  if ! eval "$UPSTREAM_GEN_CMD --seed ${seed}" > "$up_c"; then
    echo "[FAIL] seed=${seed}: upstream generation failed"
    fail=$((fail + 1))
    continue
  fi

  if ! eval "$GO_GEN_CMD --seed ${seed}" > "$go_c"; then
    echo "[FAIL] seed=${seed}: go generation failed"
    fail=$((fail + 1))
    continue
  fi

  if ! "$CC_BIN" -std=c11 -O0 "$up_c" -I"$UPSTREAM_INCLUDE" -o "$up_bin" >/dev/null 2>&1; then
    echo "[FAIL] seed=${seed}: upstream compile failed"
    fail=$((fail + 1))
    continue
  fi

  if ! "$CC_BIN" -std=c11 -O0 "$go_c" -I"$GO_INCLUDE" -o "$go_bin" >/dev/null 2>&1; then
    echo "[FAIL] seed=${seed}: go compile failed"
    fail=$((fail + 1))
    continue
  fi

  if ! timeout "$TIMEOUT_SECS" "$up_bin" > "$up_out"; then
    echo "[FAIL] seed=${seed}: upstream run failed/hanged"
    fail=$((fail + 1))
    continue
  fi

  if ! timeout "$TIMEOUT_SECS" "$go_bin" > "$go_out"; then
    echo "[FAIL] seed=${seed}: go run failed/hanged"
    fail=$((fail + 1))
    continue
  fi

  up_raw=$(extract_checksum_token "$up_out" || true)
  go_raw=$(extract_checksum_token "$go_out" || true)

  if [[ -z "$up_raw" || -z "$go_raw" ]]; then
    echo "[FAIL] seed=${seed}: checksum line missing (up='$up_raw' go='$go_raw')"
    fail=$((fail + 1))
    continue
  fi

  up_norm=$(normalize_checksum_u32 "$up_raw" || true)
  go_norm=$(normalize_checksum_u32 "$go_raw" || true)

  if [[ -z "$up_norm" || -z "$go_norm" ]]; then
    echo "[FAIL] seed=${seed}: checksum parse failed (up='$up_raw' go='$go_raw')"
    fail=$((fail + 1))
    continue
  fi

  if [[ "$up_norm" == "$go_norm" ]]; then
    echo "[PASS] seed=${seed}: checksum=${up_norm}"
    pass=$((pass + 1))
  else
    echo "[FAIL] seed=${seed}: upstream=${up_raw}(${up_norm}) go=${go_raw}(${go_norm})"
    fail=$((fail + 1))
  fi
done

echo ""
echo "Parity summary: pass=${pass} fail=${fail} total=${COUNT}"

if [[ "$fail" -ne 0 ]]; then
  exit 1
fi
