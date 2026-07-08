#!/usr/bin/env bash
# Phase 146 WS3 — export thumbs-down feedback as eval fixture candidates.
#
# Usage (repo root, API running, admin JWT):
#   source scripts/source-local-env.sh --refresh-eval-token
#   ./scripts/guardian-feedback-to-fixture.sh
#   ./scripts/guardian-feedback-to-fixture.sh --farm-id 1 --since 30d --out data/guardian_feedback_fixtures.json
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
FARM_ID="${FARM_ID:-1}"
SINCE="${SINCE:-30d}"
OUT="${OUT:-$ROOT/data/guardian_feedback_fixtures.json}"
API="${GR33N_API_URL:-http://127.0.0.1:8080}"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --farm-id) FARM_ID="$2"; shift 2 ;;
    --since) SINCE="$2"; shift 2 ;;
    --out) OUT="$2"; shift 2 ;;
    -h|--help)
      sed -n '2,8p' "$0"
      exit 0
      ;;
    *) echo "unknown option: $1" >&2; exit 1 ;;
  esac
done

if [[ -f "$ROOT/.env" ]]; then set -a && . "$ROOT/.env" && set +a; fi
if [[ -f "$ROOT/.env.local" ]]; then set -a && . "$ROOT/.env.local" && set +a; fi

if [[ -z "${GUARDIAN_EVAL_TOKEN:-}" ]]; then
  echo "error: GUARDIAN_EVAL_TOKEN not set — run: source scripts/source-local-env.sh --refresh-eval-token" >&2
  exit 1
fi

mkdir -p "$(dirname "$OUT")"
curl -sf -H "Authorization: Bearer $GUARDIAN_EVAL_TOKEN" \
  "${API%/}/v1/chat/feedback/export?farm_id=${FARM_ID}&since=${SINCE}&rating=down" \
  | python3 - "$OUT" <<'PY'
import json, sys, re, pathlib

out_path = pathlib.Path(sys.argv[1])
payload = json.load(sys.stdin)
rows = payload.get("rows") or []
candidates = []
for i, row in enumerate(rows, start=1):
    q = (row.get("question") or "").strip()
    if not q:
        continue
    slug = re.sub(r"[^a-z0-9]+", "-", q.lower()).strip("-")[:40] or "row"
    candidates.append({
        "id": f"feedback-{slug}-{i}",
        "category": "field_guide" if row.get("grounded") else "ungrounded",
        "prompt": q,
        "answer_excerpt": row.get("answer_excerpt") or "",
        "rating": row.get("rating") or "down",
        "reason": row.get("reason"),
        "grounded": bool(row.get("grounded")),
        "model": row.get("model") or "",
        "session_id": row.get("session_id"),
        "turn_index": row.get("turn_index"),
        "promote_hint": "Add archived answer to eval/score_*_test.go after human triage",
    })

doc = {
    "farm_id": payload.get("farm_id"),
    "since": payload.get("since"),
    "candidate_count": len(candidates),
    "candidates": candidates,
}
out_path.write_text(json.dumps(doc, indent=2) + "\n")
print(f"wrote {len(candidates)} candidate(s) to {out_path}")
PY
