#!/usr/bin/env bash
# Phase 83 WS1 — promote gr33n-cultivator-seed-pack-v1 to farm(s) via commons catalog.
# Records import audit + verifies platform catalog_version / row counts in Postgres (WS-F).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
API_URL="${GR33N_API_URL:-http://localhost:8080}"
LOGIN_USER="${GR33N_LOGIN_USER:-dev@gr33n.local}"
LOGIN_PASS="${GR33N_LOGIN_PASS:-devpassword}"
CATALOG_SLUG="${CATALOG_SLUG:-gr33n-cultivator-seed-pack-v1}"
FARM_IDS="${GR33N_FARM_IDS:-1}"
PACK_JSON="${AGRONOMY_PACK_JSON:-$ROOT/scripts/enterprise/sample-cultivator-seed-pack-v1.body.json}"
DRY_RUN=0
VERIFY_ONLY=0

usage() {
  cat <<EOF
Usage: $(basename "$0") [--dry-run] [--verify-only] [--farm-ids 1,2] [--slug SLUG]

Record agronomy seed pack import per farm and verify Postgres catalog matches pack contract.

  --dry-run       Print planned API actions + DB checks (no JWT / no writes)
  --verify-only   Skip catalog import POST; run DB verification only
  --farm-ids      Comma-separated farm_id list (default: 1)
  --slug          Commons catalog slug (default: gr33n-cultivator-seed-pack-v1)

Real run requires: API up (unless --verify-only), migration 20260618_phase83_cultivator_seed_pack_v1.sql,
Phase 84 catalog seed applied.

Env: GR33N_API_URL, GR33N_LOGIN_USER, GR33N_LOGIN_PASS, GR33N_FARM_IDS, DATABASE_URL

Next step: ./scripts/enterprise/guardian-bootstrap-farm.sh --farm-id N
EOF
}

load_env() {
  if [[ -f "$ROOT/.env" ]]; then
    set -a
    # shellcheck disable=1091
    source "$ROOT/.env"
    set +a
    while IFS= read -r line || [[ -n "$line" ]]; do
      [[ "$line" =~ ^[[:space:]]*# ]] && continue
      [[ "$line" =~ ^PORT= ]] && API_URL="http://localhost:${line#PORT=}"
    done < "$ROOT/.env" || true
  fi
}

login_jwt() {
  local resp token
  resp=$(curl -sf -X POST "${API_URL}/auth/login" \
    -H 'Content-Type: application/json' \
    -d "{\"username\":\"${LOGIN_USER}\",\"password\":\"${LOGIN_PASS}\"}") \
    || { echo "Login failed (${LOGIN_USER}). Set GR33N_LOGIN_USER/PASS." >&2; exit 1; }
  token=$(python3 -c "import json,sys; print(json.load(sys.stdin).get('token',''))" <<<"$resp")
  if [[ -z "$token" ]]; then
    echo "No JWT in login response" >&2
    exit 1
  fi
  echo "$token"
}

verify_db_contract() {
  PACK_JSON="$PACK_JSON" python3 - <<'PY'
import json, os, subprocess, sys

pack_path = os.environ["PACK_JSON"]
with open(pack_path, encoding="utf-8") as f:
    pack = json.load(f)

db_url = os.environ.get("DATABASE_URL", "")
if not db_url:
    print("error: DATABASE_URL required for verification", file=sys.stderr)
    sys.exit(1)

def psql(q):
    r = subprocess.run(
        ["psql", db_url, "-v", "ON_ERROR_STOP=1", "-tAc", q],
        capture_output=True, text=True,
    )
    if r.returncode != 0:
        print(r.stderr, file=sys.stderr)
        sys.exit(1)
    return (r.stdout or "").strip()

want_ver = int(pack.get("platform_catalog_version", 0))
got_ver = int(psql("SELECT COALESCE(MAX(catalog_version), 0) FROM gr33ncrops.crop_catalog_entries") or "0")
if got_ver < want_ver:
    print(f"FAIL platform_catalog_version: DB={got_ver} pack expects >={want_ver}", file=sys.stderr)
    sys.exit(1)
print(f"OK   platform_catalog_version: DB max={got_ver} (pack>={want_ver})")

expected = pack.get("expected_counts") or {}
checks = [
    ("crop_catalog_entries", "SELECT count(*) FROM gr33ncrops.crop_catalog_entries"),
    ("supported_crops", "SELECT count(*) FROM gr33ncrops.crop_catalog_entries WHERE supported"),
    ("unsupported_crops", "SELECT count(*) FROM gr33ncrops.crop_catalog_entries WHERE NOT supported"),
    ("field_guides_published", "SELECT count(*) FROM gr33ncrops.agronomy_field_guides WHERE published"),
    ("builtin_profiles", "SELECT count(*) FROM gr33ncrops.crop_profiles WHERE farm_id IS NULL AND is_builtin"),
]
fail = 0
for key, sql in checks:
    got = int(psql(sql) or "0")
    want = int(expected.get(key, 0))
    if want and got < want:
        print(f"FAIL {key}: got {got} want >={want}", file=sys.stderr)
        fail = 1
    else:
        label = f">={want}" if want else ""
        print(f"OK   {key}: {got} {label}".rstrip())

if fail:
    sys.exit(1)
print("agronomy seed pack DB contract OK")
PY
}

run_dry() {
  CATALOG_SLUG="$CATALOG_SLUG" FARM_IDS="$FARM_IDS" PACK_JSON="$PACK_JSON" VERIFY_ONLY="$VERIFY_ONLY" python3 - <<'PY'
import json, os

farm_ids = [int(x) for x in os.environ["FARM_IDS"].split(",") if x.strip()]
slug = os.environ["CATALOG_SLUG"]
with open(os.environ["PACK_JSON"], encoding="utf-8") as f:
    body = json.load(f)

print(f"Agronomy pack: slug={slug} platform_catalog_version={body.get('platform_catalog_version')}")
print(f"Source JSON: {os.environ['PACK_JSON']}")
print("")
if os.environ.get("VERIFY_ONLY") != "1":
    for fid in farm_ids:
        print(f"[farm {fid}] POST /farms/{fid}/commons/catalog-imports")
        print(f'         body: {{"slug": "{slug}", "note": "Phase 83 import-agronomy-seed-pack"}}')
        print("")
print("[db] verify platform_catalog_version + expected_counts (see verify_db_contract)")
print("")
print("Next: ./scripts/enterprise/guardian-bootstrap-farm.sh --farm-id", farm_ids[0] if farm_ids else "N")
print("Dry-run only — no HTTP/DB verification executed.")
PY
}

run_apply() {
  verify_db_contract
  if [[ "$VERIFY_ONLY" -eq 1 ]]; then
    echo "Verify-only — skipping catalog import POST."
    return 0
  fi
  local jwt
  jwt=$(login_jwt)
  GR33N_API_URL="$API_URL" CATALOG_SLUG="$CATALOG_SLUG" FARM_IDS="$FARM_IDS" JWT="$jwt" python3 - <<'PY'
import json, os, sys, urllib.error, urllib.request

api = os.environ["GR33N_API_URL"].rstrip("/")
slug = os.environ["CATALOG_SLUG"]
farm_ids = [int(x) for x in os.environ["FARM_IDS"].split(",") if x.strip()]
token = os.environ["JWT"]

def req(method, path, data=None):
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    body = json.dumps(data).encode() if data is not None else None
    r = urllib.request.Request(api + path, data=body, headers=headers, method=method)
    try:
        with urllib.request.urlopen(r, timeout=30) as resp:
            raw = resp.read().decode()
            return resp.status, json.loads(raw) if raw else {}
    except urllib.error.HTTPError as e:
        err = e.read().decode()
        print(f"HTTP {e.code} {method} {path}: {err[:500]}", file=sys.stderr)
        sys.exit(1)

status, cat = req("GET", f"/commons/catalog/{slug}")
body = cat.get("body") or {}
print(f"Loaded catalog slug={slug} platform_catalog_version={body.get('platform_catalog_version')}")

for fid in farm_ids:
    print(f"--- farm_id={fid} ---")
    status, _ = req("POST", f"/farms/{fid}/commons/catalog-imports", {
        "slug": slug,
        "note": "Phase 83 import-agronomy-seed-pack.sh",
    })
    print(f"  catalog import recorded (HTTP {status})")

print("Done. Run guardian-bootstrap-farm.sh next.")
PY
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=1; shift ;;
    --verify-only) VERIFY_ONLY=1; shift ;;
    --farm-ids) FARM_IDS="$2"; shift 2 ;;
    --slug) CATALOG_SLUG="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage >&2; exit 1 ;;
  esac
done

load_env

if [[ ! -f "$PACK_JSON" ]]; then
  echo "Missing pack JSON: $PACK_JSON" >&2
  exit 1
fi

if [[ "$DRY_RUN" -eq 1 ]]; then
  run_dry
else
  run_apply
fi
