#!/usr/bin/env bash
# Phase 33 WS5 — apply an enterprise site manifest via the public API.
# Creates a farm (optionally under an org), its zones, imports a commons recipe
# pack, and prints Pi/edge wiring hints + smoke commands. Idempotent-ish:
# existing zones (by name) are skipped on a real run.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
API_URL="${GR33N_API_URL:-http://localhost:8080}"
LOGIN_USER="${GR33N_LOGIN_USER:-dev@gr33n.local}"
LOGIN_PASS="${GR33N_LOGIN_PASS:-devpassword}"
MANIFEST="${SITE_MANIFEST:-$ROOT/scripts/enterprise/site-manifest.example.yaml}"
DRY_RUN=0

usage() {
  cat <<EOF
Usage: $(basename "$0") [--dry-run] [--manifest PATH]

Stand up a warehouse site from a YAML manifest (farm + zones + recipe pack).

  --dry-run       Print planned API actions only (no JWT required)
  --manifest      Path to site manifest YAML (default: scripts/enterprise/site-manifest.example.yaml)

Real run requires: API up, farm-admin JWT, python3 + PyYAML (pip install pyyaml).
Recipe pack import reuses POST /farms/{id}/commons/catalog-imports — the pack
slug must already be published in the commons catalog.

Env: GR33N_API_URL, GR33N_LOGIN_USER, GR33N_LOGIN_PASS, SITE_MANIFEST

See scripts/enterprise/README.md and docs/hypothetical-enterprise-topology.md
EOF
}

load_env() {
  if [[ -f "$ROOT/.env" ]]; then
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

run_dry() {
  MANIFEST="$MANIFEST" python3 - <<'PY'
import os, sys
try:
    import yaml
except ImportError:
    print("PyYAML not installed — `pip install pyyaml` for dry-run/apply.", file=sys.stderr)
    sys.exit(1)

with open(os.environ["MANIFEST"], encoding="utf-8") as f:
    m = yaml.safe_load(f)

org = m.get("org_slug")
farm = m["farm_name"]
zones = m.get("zones") or []
pack = m.get("recipe_pack_slug")
hints = m.get("pi_device_hints") or []

print(f"Site manifest: {os.environ['MANIFEST']}")
print("")
if org:
    print(f"[org]  ensure organization slug={org!r} (POST /organizations if missing)")
print(f"[farm] POST /farms  name={farm!r}" + (f"  org_slug={org!r}" if org else ""))
for z in zones:
    print(f"[zone] POST /farms/{{farm_id}}/zones  name={z['name']!r} zone_type={z.get('type')!r}")
if pack:
    print(f"[pack] POST /farms/{{farm_id}}/commons/catalog-imports  slug={pack!r}")
print("")
if hints:
    print("Pi/edge wiring hints (informational — provision on-site):")
    for h in hints:
        print(f"  - {h.get('device_uid')}: {h.get('role')}  (zone {h.get('zone')!r})")
    print("  See docs/pi-integration-guide.md and the Phase 37 guided wiring procedures.")
print("")
print("Dry-run only — no HTTP calls made.")
PY
}

run_apply() {
  local jwt
  jwt=$(login_jwt)
  GR33N_API_URL="$API_URL" MANIFEST="$MANIFEST" JWT="$jwt" python3 - <<'PY'
import json, os, sys, urllib.error, urllib.request
try:
    import yaml
except ImportError:
    print("PyYAML not installed — `pip install pyyaml`.", file=sys.stderr)
    sys.exit(1)

api = os.environ["GR33N_API_URL"].rstrip("/")
token = os.environ["JWT"]
with open(os.environ["MANIFEST"], encoding="utf-8") as f:
    m = yaml.safe_load(f)

def req(method, path, data=None):
    headers = {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}
    body = json.dumps(data).encode() if data is not None else None
    r = urllib.request.Request(api + path, data=body, headers=headers, method=method)
    try:
        with urllib.request.urlopen(r, timeout=30) as resp:
            raw = resp.read().decode()
            return resp.status, (json.loads(raw) if raw else {})
    except urllib.error.HTTPError as e:
        print(f"HTTP {e.code} {method} {path}: {e.read().decode()[:500]}", file=sys.stderr)
        sys.exit(1)

farm_payload = {"name": m["farm_name"]}
if m.get("org_slug"):
    farm_payload["org_slug"] = m["org_slug"]
status, farm = req("POST", "/farms", farm_payload)
farm_id = farm.get("id")
print(f"CREATED farm id={farm_id} name={m['farm_name']!r} (HTTP {status})")

status, existing = req("GET", f"/farms/{farm_id}/zones")
have = {z.get("name") for z in existing if isinstance(z, dict)}
for z in (m.get("zones") or []):
    if z["name"] in have:
        print(f"  SKIP zone {z['name']!r} (already exists)")
        continue
    status, created = req("POST", f"/farms/{farm_id}/zones",
                          {"name": z["name"], "zone_type": z.get("type")})
    print(f"  CREATED zone {z['name']!r} id={created.get('id')} (HTTP {status})")

if m.get("recipe_pack_slug"):
    status, _ = req("POST", f"/farms/{farm_id}/commons/catalog-imports",
                    {"slug": m["recipe_pack_slug"], "note": "apply-site-manifest.sh"})
    print(f"  recipe pack {m['recipe_pack_slug']!r} import recorded (HTTP {status})")

for h in (m.get("pi_device_hints") or []):
    print(f"  Pi hint: {h.get('device_uid')} — {h.get('role')} (zone {h.get('zone')!r}) — provision on-site")

print(f"Done. Smoke: curl {api}/farms/{farm_id}/zones")
PY
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --dry-run) DRY_RUN=1; shift ;;
    --manifest) MANIFEST="$2"; shift 2 ;;
    -h|--help) usage; exit 0 ;;
    *) echo "Unknown option: $1" >&2; usage >&2; exit 1 ;;
  esac
done

load_env

if [[ ! -f "$MANIFEST" ]]; then
  echo "Missing manifest: $MANIFEST" >&2
  exit 1
fi

if [[ "$DRY_RUN" -eq 1 ]]; then
  run_dry
else
  run_apply
fi
