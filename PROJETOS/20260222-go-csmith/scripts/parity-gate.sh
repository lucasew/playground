#!/usr/bin/env bash
set -euo pipefail

# Opinionated parity gate: compares upstream Csmith and csmith-go over a seed range.
COUNT="${COUNT:-20}"
SEED_START="${SEED_START:-2}"
TIMEOUT_SECS="${TIMEOUT_SECS:-5}"

UPSTREAM_GEN_CMD="${UPSTREAM_GEN_CMD:-/nix/store/hrf9nixgjz33q1563l9bxx155py477qv-csmith-2.3.0/bin/csmith}"
UPSTREAM_INCLUDE="${UPSTREAM_INCLUDE:-/nix/store/hrf9nixgjz33q1563l9bxx155py477qv-csmith-2.3.0/include/csmith-2.3.0}"
GO_GEN_CMD="${GO_GEN_CMD:-GOCACHE=/tmp/go-cache go run ./cmd/csmith-go}"

UPSTREAM_GEN_CMD="$UPSTREAM_GEN_CMD" \
UPSTREAM_INCLUDE="$UPSTREAM_INCLUDE" \
GO_INCLUDE="$UPSTREAM_INCLUDE" \
GO_GEN_CMD="$GO_GEN_CMD" \
scripts/compare-upstream.sh \
  --count "$COUNT" \
  --seed-start "$SEED_START" \
  --timeout "$TIMEOUT_SECS"
