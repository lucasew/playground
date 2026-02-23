#!/usr/bin/env bash
set -euo pipefail

SEED="${SEED:-2}"
MAX_ITERS="${MAX_ITERS:-20}"
WORKDIR="${WORKDIR:-/tmp/csmith-parity}"
CONTEXT="${CONTEXT:-8}"
STRICT_RAW="${STRICT_RAW:-0}"
PROMPT_FILE="${PROMPT_FILE:-PROMPT.md}"
AGENT="${AGENT:-ralph}"
# Claude CLI command (can include flags).
CLAUDE_CMD="${CLAUDE_CMD:-claude}"
# If RALPH_CMD contains "{prompt}", placeholder is replaced with generated prompt path.
# Otherwise, generated prompt path is appended as last argument.
RALPH_CMD="${RALPH_CMD:-ralph run}"
DRY_RUN="${DRY_RUN:-0}"
MEMORY_FILE="${MEMORY_FILE:-}"
CHECKPOINT_COMMITS="${CHECKPOINT_COMMITS:-1}"

usage() {
  cat <<'USAGE'
Usage: scripts/ralph-loop.sh [options]

Iterative loop:
  1) run RNG divergence checker
  2) if mismatch, call Ralph with base prompt + current divergence
  3) repeat until match or max iterations

Options:
  --seed N
  --max-iters N
  --workdir DIR
  --context N
  --strict-raw
  --prompt-file PATH
  --agent NAME            (ralph|claude)
  --claude-cmd "CMD"
  --ralph-cmd "CMD"
  --memory-file PATH
  --no-checkpoint-commits
  --checkpoint-commits
  --dry-run
  -h|--help

Env overrides:
  SEED, MAX_ITERS, WORKDIR, CONTEXT, STRICT_RAW, PROMPT_FILE, AGENT, CLAUDE_CMD, RALPH_CMD, MEMORY_FILE, CHECKPOINT_COMMITS, DRY_RUN

Examples:
  scripts/ralph-loop.sh --seed 2 --max-iters 30
  scripts/ralph-loop.sh --agent claude --claude-cmd "claude"
  scripts/ralph-loop.sh --ralph-cmd "ralph run --prompt-file {prompt}"
  scripts/ralph-loop.sh --dry-run
USAGE
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --seed) SEED="$2"; shift 2 ;;
    --max-iters) MAX_ITERS="$2"; shift 2 ;;
    --workdir) WORKDIR="$2"; shift 2 ;;
    --context) CONTEXT="$2"; shift 2 ;;
    --strict-raw) STRICT_RAW=1; shift ;;
    --prompt-file) PROMPT_FILE="$2"; shift 2 ;;
    --agent) AGENT="$2"; shift 2 ;;
    --claude-cmd) CLAUDE_CMD="$2"; shift 2 ;;
    --ralph-cmd) RALPH_CMD="$2"; shift 2 ;;
    --memory-file) MEMORY_FILE="$2"; shift 2 ;;
    --checkpoint-commits) CHECKPOINT_COMMITS=1; shift ;;
    --no-checkpoint-commits) CHECKPOINT_COMMITS=0; shift ;;
    --dry-run) DRY_RUN=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *)
      echo "unknown argument: $1" >&2
      usage
      exit 2
      ;;
  esac
done

if [[ "$AGENT" != "ralph" && "$AGENT" != "claude" ]]; then
  echo "invalid --agent: $AGENT (use ralph or claude)" >&2
  exit 2
fi

if [[ ! -f "$PROMPT_FILE" ]]; then
  echo "prompt file not found: $PROMPT_FILE" >&2
  exit 1
fi

if [[ ! -x scripts/find-rng-divergence.sh ]]; then
  chmod +x scripts/find-rng-divergence.sh
fi

mkdir -p "$WORKDIR"

if [[ -z "$MEMORY_FILE" ]]; then
  MEMORY_FILE="$WORKDIR/seed_${SEED}.memory.md"
fi

if [[ ! -f "$MEMORY_FILE" ]]; then
  cat > "$MEMORY_FILE" <<EOF
# Ralph Loop Memory

- seed: $SEED
- started_at: $(date -u +"%Y-%m-%dT%H:%M:%SZ")
- prompt_file: $PROMPT_FILE
- agent: $AGENT

## Iterations
EOF
fi

run_agent() {
  local prompt_path="$1"
  local log_path="$2"
  local cmd=""
  if [[ "$AGENT" == "claude" ]]; then
    cmd="$CLAUDE_CMD -p \"\$(cat '$prompt_path')\""
  else
    if [[ "$RALPH_CMD" == *"{prompt}"* ]]; then
      cmd="${RALPH_CMD//\{prompt\}/$prompt_path}"
    else
      cmd="$RALPH_CMD \"$prompt_path\""
    fi
  fi

  echo "[loop] agent=$AGENT cmd: $cmd"
  if [[ "$DRY_RUN" == "1" ]]; then
    : > "$log_path"
    return 0
  fi
  bash -lc "$cmd" | tee "$log_path"
}

for ((iter=1; iter<=MAX_ITERS; iter++)); do
  echo ""
  echo "========== ITERATION $iter/$MAX_ITERS =========="

  report_file="$WORKDIR/seed_${SEED}.iter_${iter}.report.txt"
  if scripts/find-rng-divergence.sh --seed "$SEED" --workdir "$WORKDIR" --context "$CONTEXT" $([[ "$STRICT_RAW" == "1" ]] && echo "--strict-raw") | tee "$report_file"; then
    :
  fi

  if grep -q "^result=match" "$report_file"; then
    echo "[loop] parity achieved at iteration $iter"
    exit 0
  fi

  mismatch_event="$(awk -F= '/^first_divergence_event=/{print $2}' "$report_file" | tail -n1)"
  up_event="$(awk -F': ' '/^upstream_event:/{print $2}' "$report_file" | tail -n1)"
  go_event="$(awk -F': ' '/^go_event:/{print $2}' "$report_file" | tail -n1)"
  reason="$(awk -F= '/^reason=/{print $2}' "$report_file" | tail -n1)"

  iter_prompt="$WORKDIR/seed_${SEED}.iter_${iter}.prompt.md"
  iter_log="$WORKDIR/seed_${SEED}.iter_${iter}.agent.log"
  {
    cat "$PROMPT_FILE"
    cat <<EOF

---
## Iteração Atual
- seed: $SEED
- iteration: $iter/$MAX_ITERS
- first_divergence_event: ${mismatch_event:-unknown}
- reason: ${reason:-unknown}
- upstream_event: ${up_event:-<none>}
- go_event: ${go_event:-<none>}
- report_file: $report_file

Instrução:
- Faça as mudanças de código necessárias para empurrar esse desvio para frente.
- Depois finalize para que o loop rode a próxima comparação.
EOF
  } > "$iter_prompt"

  {
    echo ""
    echo "### Iteration $iter"
    echo "- timestamp: $(date -u +"%Y-%m-%dT%H:%M:%SZ")"
    echo "- mismatch_event: ${mismatch_event:-unknown}"
    echo "- reason: ${reason:-unknown}"
    echo "- upstream_event: ${up_event:-<none>}"
    echo "- go_event: ${go_event:-<none>}"
    echo "- report_file: $report_file"
    echo "- prompt_file: $iter_prompt"
    echo "- agent_log: $iter_log"
  } >> "$MEMORY_FILE"

  run_agent "$iter_prompt" "$iter_log"

  if [[ -s "$iter_log" ]]; then
    mem_notes="$(grep -E '^MEMORY:' "$iter_log" || true)"
    if [[ -n "$mem_notes" ]]; then
      {
        echo "- agent_memory:"
        echo '```text'
        echo "$mem_notes"
        echo '```'
      } >> "$MEMORY_FILE"
    fi
  fi

  if [[ "$CHECKPOINT_COMMITS" == "1" && "$DRY_RUN" != "1" ]]; then
    if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
      if ! git diff --quiet || ! git diff --cached --quiet; then
        git add -A
        if ! git diff --cached --quiet; then
          commit_msg="checkpoint: seed ${SEED} iter ${iter}"
          if [[ -n "${mismatch_event:-}" ]]; then
            commit_msg="${commit_msg} div=${mismatch_event}"
          fi
          git commit -m "$commit_msg" >/dev/null
          echo "[loop] checkpoint commit created: $commit_msg"
          echo "- checkpoint_commit: $commit_msg" >> "$MEMORY_FILE"
        fi
      fi
    else
      echo "[loop] skipping checkpoint commit (not a git repository)"
    fi
  fi
done

echo "[loop] max iterations reached without parity match"
exit 1
