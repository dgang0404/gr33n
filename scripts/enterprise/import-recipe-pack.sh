#!/usr/bin/env bash
# Phase 31 WS5 / Phase 108 — promote a commons Recipe Pack to one or more farms via public API.
# Records catalog import audit + creates fertigation programs (idempotent by name) + crop_key metadata.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
API_URL="${GR33N_API_URL:-http://localhost:8080}"
LOGIN_USER="${GR33N_LOGIN_USER:-dev@gr33n.local}"
LOGIN_PASS="${GR33N_LOGIN_PASS:-devpassword}"
CATALOG_SLUG="${CATALOG_SLUG:-gr33n-recipe-pack-v7-lettuce-veg}"
FARM_IDS="${GR33N_FARM_IDS:-1}"
PACK_JSON="${RECIPE_PACK_JSON:-$ROOT/scripts/enterprise/sample-recipe-pack-v7.body.json}"
DRY_RUN=0

usage() {
  cat <<EOF
Usage: $(basename "$0") [--dry-run] [--farm-ids 1,2] [--slug SLUG]

Promote Recipe Pack v7 (demo) to multiple farms without a core broadcast feature.

  --dry-run       Print planned API actions only (no JWT required)
  --farm-ids      Comma-separated farm_id list (default: 1)
  --slug          Commons catalog slug (default: gr33n-recipe-pack-v7-lettuce-veg)

Real run requires: API up, farm admin JWT, migrations through phase 108 applied.

Env: GR33N_API_URL, GR33N_LOGIN_USER, GR33N_LOGIN_PASS, GR33N_FARM_IDS, CATALOG_SLUG

See scripts/enterprise/README.md and docs/commons-catalog-operator-playbook.md
EOF
}

load_env() {
  if [[ -f "$ROOT/.env" ]]; then
    # shellcheck disable=SC1091
    while IFS= read -r line || [[ -n "$line" ]]; do
      [[ "$line" =~ ^[[:space:]]*# ]] && continue
      [[ "$line" =~ ^PORT= ]] && port="${line#PORT=}" && API_URL="http://localhost:${port}"
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

run_dry() {
  CATALOG_SLUG="$CATALOG_SLUG" FARM_IDS="$FARM_IDS" PACK_JSON="$PACK_JSON" python3 - <<'PY'
import json, os, sys

farm_ids = [int(x) for x in os.environ["FARM_IDS"].split(",") if x.strip()]
slug = os.environ["CATALOG_SLUG"]
pack_path = os.environ["PACK_JSON"]
with open(pack_path, encoding="utf-8") as f:
    body = json.load(f)
programs = body.get("programs") or []

def validate_programs(programs):
    for p in programs:
        keys = p.get("recommended_crop_keys") or []
        for ck in keys:
            if not isinstance(ck, str) or not ck.strip():
                print(f"Invalid recommended_crop_keys on {p.get('name')!r}", file=sys.stderr)
                sys.exit(1)

validate_programs(programs)
print(f"Recipe pack: slug={slug} pack_version={body.get('pack_version')} programs={len(programs)}")
print(f"Source JSON: {pack_path}")
print("(dry-run: crop_key validation against live catalog skipped)")
print("")
for fid in farm_ids:
    print(f"[farm {fid}] POST /farms/{fid}/commons/catalog-imports")
    print(f'         body: {{"slug": "{slug}", "note": "Phase 108 import-recipe-pack"}}')
    for p in programs:
        name = p["name"]
        meta = {k: p[k] for k in (
            "recommended_crop_keys", "recommended_stages", "profile_ec_source", "ec_band_mscm"
        ) if p.get(k)}
        print(f"[farm {fid}] POST /farms/{fid}/fertigation/programs  (skip if name exists)")
        print(f"         name={name!r} ec_low={p.get('ec_trigger_low')} is_active={p.get('is_active', False)}")
        if meta:
            print(f"         PATCH /fertigation/programs/{{id}}/metadata  {meta!r}")
    print("")
print("Dry-run only — no HTTP calls made.")
PY
}

run_apply() {
  local jwt
  jwt=$(login_jwt)
  GR33N_API_URL="$API_URL" CATALOG_SLUG="$CATALOG_SLUG" FARM_IDS="$FARM_IDS" \
    PACK_JSON="$PACK_JSON" JWT="$jwt" python3 - <<'PY'
import json, os, sys, urllib.error, urllib.request

api = os.environ["GR33N_API_URL"].rstrip("/")
slug = os.environ["CATALOG_SLUG"]
farm_ids = [int(x) for x in os.environ["FARM_IDS"].split(",") if x.strip()]
token = os.environ["JWT"]
pack_path = os.environ["PACK_JSON"]

with open(pack_path, encoding="utf-8") as f:
    local_body = json.load(f)

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

def catalog_crop_keys():
    status, cat = req("GET", "/commons/crop-catalog")
    entries = cat.get("entries") or []
    return {str(e.get("crop_key", "")).strip().lower() for e in entries if e.get("crop_key")}

def program_metadata_payload(p):
    payload = {}
    for key in ("recommended_crop_keys", "recommended_stages", "profile_ec_source", "ec_band_mscm"):
        if p.get(key):
            payload[key] = p[key]
    return payload

def validate_crop_keys(programs, valid_keys):
    for p in programs:
        for ck in p.get("recommended_crop_keys") or []:
            norm = str(ck).strip().lower()
            if norm not in valid_keys:
                print(
                    f"Unknown crop_key {ck!r} in program {p.get('name')!r} "
                    f"(not in /commons/crop-catalog)",
                    file=sys.stderr,
                )
                sys.exit(1)
        src = p.get("profile_ec_source") or {}
        src_key = str(src.get("crop_key", "")).strip().lower()
        if src_key and src_key not in valid_keys:
            print(
                f"Unknown profile_ec_source.crop_key {src.get('crop_key')!r} "
                f"in program {p.get('name')!r}",
                file=sys.stderr,
            )
            sys.exit(1)

def apply_metadata(pid, p):
    meta = program_metadata_payload(p)
    if not meta:
        return
    status, _ = req("PATCH", f"/fertigation/programs/{pid}/metadata", meta)
    print(f"    metadata applied (HTTP {status}) keys={meta.get('recommended_crop_keys')}")

valid_keys = catalog_crop_keys()
if not valid_keys:
    print("No crop keys from /commons/crop-catalog", file=sys.stderr)
    sys.exit(1)

status, cat = req("GET", f"/commons/catalog/{slug}")
body = cat.get("body") or local_body
programs = body.get("programs") or []
validate_crop_keys(programs, valid_keys)
print(f"Loaded catalog slug={slug} programs={len(programs)}")

for fid in farm_ids:
    print(f"--- farm_id={fid} ---")
    status, _ = req("POST", f"/farms/{fid}/commons/catalog-imports", {
        "slug": slug,
        "note": "Phase 108 import-recipe-pack.sh",
    })
    print(f"  catalog import recorded (HTTP {status})")

    status, existing = req("GET", f"/farms/{fid}/fertigation/programs")
    by_name = {p.get("name"): p for p in existing if isinstance(p, dict)}

    for p in programs:
        name = p["name"]
        meta = program_metadata_payload(p)
        if name in by_name:
            pid = by_name[name]["id"]
            print(f"  SKIP program {name!r} id={pid} (already exists)")
            apply_metadata(pid, p)
            continue
        payload = {
            "name": name,
            "description": p.get("description"),
            "total_volume_liters": p["total_volume_liters"],
            "ec_trigger_low": p["ec_trigger_low"],
            "ph_trigger_low": p["ph_trigger_low"],
            "ph_trigger_high": p["ph_trigger_high"],
            "is_active": bool(p.get("is_active", False)),
        }
        payload.update(meta)
        status, created = req("POST", f"/farms/{fid}/fertigation/programs", payload)
        pid = created.get("id")
        print(f"  CREATED program {name!r} id={pid} (HTTP {status})")
        by_name[name] = created

print("Done.")
PY
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=1; shift ;;
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
