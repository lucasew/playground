#!/usr/bin/env bash
set -euo pipefail

SEED="${SEED:-2}"
WORKDIR="${WORKDIR:-/tmp/csmith-parity}"
CONTEXT="${CONTEXT:-8}"
UPSTREAM_CMD="${UPSTREAM_CMD:-/tmp/csmith/build-instrumented/src/csmith}"
GO_CMD="${GO_CMD:-GOCACHE=/tmp/go-cache go run ./cmd/csmith-go}"
STRICT_RAW="${STRICT_RAW:-0}"

usage() {
  cat <<'USAGE'
Usage: scripts/find-rng-divergence.sh [--seed N] [--workdir DIR] [--context N] [--strict-raw]

Runs upstream Csmith and csmith-go with the same seed, captures RNG traces,
normalizes them, and reports the first divergence event.

Env overrides:
  SEED        (default: 2)
  WORKDIR     (default: /tmp/csmith-parity)
  CONTEXT     (default: 8)
  UPSTREAM_CMD (default: /tmp/csmith/build-instrumented/src/csmith)
  GO_CMD      (default: GOCACHE=/tmp/go-cache go run ./cmd/csmith-go)
  STRICT_RAW  (default: 0) compare tries/raw too (1 = enabled)

Examples:
  scripts/find-rng-divergence.sh --seed 2
  UPSTREAM_CMD='/tmp/csmith/build-instrumented/src/csmith' scripts/find-rng-divergence.sh --seed 2 --strict-raw
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --seed)
      SEED="$2"; shift 2 ;;
    --workdir)
      WORKDIR="$2"; shift 2 ;;
    --context)
      CONTEXT="$2"; shift 2 ;;
    --strict-raw)
      STRICT_RAW=1; shift ;;
    -h|--help)
      usage; exit 0 ;;
    *)
      echo "unknown argument: $1" >&2
      usage
      exit 2 ;;
  esac
done

mkdir -p "$WORKDIR"
BASE="${WORKDIR}/seed_${SEED}"
UP_RNG="${BASE}.up.rng"
GO_RNG="${BASE}.go.rng"
UP_C="${BASE}.up.c"
GO_C="${BASE}.go.c"
UP_NORM="${BASE}.up.norm"
GO_NORM="${BASE}.go.norm"
CMP_OUT="${BASE}.cmp.txt"

if [[ ! -x "$(command -v timeout || true)" ]]; then
  echo "timeout command not found" >&2
  exit 1
fi

echo "[1/4] Generating upstream trace..."
if ! CSMITH_TRACE_RNG=1 timeout 60s bash -lc "$UPSTREAM_CMD --seed $SEED > '$UP_C' 2> '$UP_RNG'"; then
  echo "upstream generation/trace failed" >&2
  exit 1
fi

echo "[2/4] Generating go trace..."
if ! CSMITH_TRACE_RNG=1 CSMITH_TRACE_RNG_RAW=1 CSMITH_TRACE_RNG_FILE="$GO_RNG" timeout 60s bash -lc "$GO_CMD --seed $SEED > '$GO_C'"; then
  echo "go generation/trace failed" >&2
  exit 1
fi

echo "[3/4] Normalizing traces..."
awk '
  BEGIN { idx=0 }
  /^U depth=/ {
    idx++
    n=""; v=""; tries="-"; raw="-"
    for (i=1; i<=NF; i++) {
      if ($i ~ /^n=/) { split($i,a,"="); n=a[2] }
      else if ($i ~ /^v=/) { split($i,a,"="); v=a[2] }
      else if ($i ~ /^tries=/) { split($i,a,"="); tries=a[2] }
      else if ($i ~ /^raw=/) { split($i,a,"="); raw=a[2] }
    }
    print idx, "U", n, v, tries, raw
    next
  }
  /^F depth=/ {
    idx++
    p=""; v=""
    for (i=1; i<=NF; i++) {
      if ($i ~ /^p=/) { split($i,a,"="); p=a[2] }
      else if ($i ~ /^v=/) { split($i,a,"="); v=a[2] }
    }
    print idx, "F", p, v, "-", "-"
    next
  }
' "$UP_RNG" > "$UP_NORM"

awk '
  $2=="U" {
    tries="-"; raw="-"
    for (i=6; i<=NF; i++) {
      if ($i ~ /^tries=/) { split($i,a,"="); tries=a[2] }
      else if ($i ~ /^raw=/) { split($i,a,"="); raw=a[2] }
    }
    print $1, "U", $3, $5, tries, raw
    next
  }
  $2=="F" {
    print $1, "F", $3, $5, "-", "-"
    next
  }
' "$GO_RNG" > "$GO_NORM"

up_count=$(wc -l < "$UP_NORM" | tr -d ' ')
go_count=$(wc -l < "$GO_NORM" | tr -d ' ')

echo "[4/4] Comparing events..."
awk -v strict="$STRICT_RAW" '
  NR==FNR {
    up_kind[NR]=$2
    up_arg[NR]=$3
    up_val[NR]=$4
    up_try[NR]=$5
    up_raw[NR]=$6
    up_line[NR]=$0
    up_n=NR
    next
  }
  {
    go_line[FNR]=$0
    go_n=FNR
    if (!mismatch) {
      if (FNR > up_n) {
        mismatch=1
        why="go_has_extra_events"
      } else {
        bad = (up_kind[FNR] != $2 || up_arg[FNR] != $3 || up_val[FNR] != $4)
        if (!bad && strict == 1 && $2 == "U") {
          if (up_try[FNR] != $5 || up_raw[FNR] != $6) bad=1
        }
        if (bad) {
          mismatch=1
          why="event_diff"
        }
      }
      if (mismatch) first=FNR
    }
  }
  END {
    if (!mismatch && up_n != go_n) {
      mismatch=1
      first=(up_n<go_n?up_n+1:go_n+1)
      why=(up_n>go_n?"go_ended_early":"go_has_extra_events")
    }
    if (!mismatch) {
      print "MATCH"
      exit
    }
    print "MISMATCH", first, why
  }
' "$UP_NORM" "$GO_NORM" > "$CMP_OUT"

read -r status mismatch_at why < "$CMP_OUT" || true
echo "seed=$SEED"
echo "upstream_events=$up_count"
echo "go_events=$go_count"

if [[ "$status" == "MATCH" ]]; then
  echo "result=match (no divergence in compared events)"
  echo "files:"
  echo "  $UP_RNG"
  echo "  $GO_RNG"
  echo "  $UP_NORM"
  echo "  $GO_NORM"
  exit 0
fi

echo "result=mismatch"
echo "first_divergence_event=$mismatch_at"
echo "reason=$why"

up_line=$(sed -n "${mismatch_at}p" "$UP_NORM" || true)
go_line=$(sed -n "${mismatch_at}p" "$GO_NORM" || true)
echo "upstream_event: ${up_line:-<none>}"
echo "go_event:       ${go_line:-<none>}"

start=$((mismatch_at - CONTEXT))
if (( start < 1 )); then start=1; fi
end=$((mismatch_at + CONTEXT))

echo ""
echo "upstream_context (${start}..${end}):"
nl -ba "$UP_NORM" | sed -n "${start},${end}p"
echo ""
echo "go_context (${start}..${end}):"
nl -ba "$GO_NORM" | sed -n "${start},${end}p"
echo ""
echo "artifacts:"
echo "  $UP_RNG"
echo "  $GO_RNG"
echo "  $UP_NORM"
echo "  $GO_NORM"

