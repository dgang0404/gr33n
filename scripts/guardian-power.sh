#!/usr/bin/env bash
# Phase 163 WS4 — admin-only Ollama service power control (not exposed via JWT API).
#
# Guardian "Rest now" / auto-dormant only unload models from RAM — the Ollama
# daemon keeps running. For solar/battery sites that want to cut the full
# process draw between sessions, an admin can stop/start the systemd unit.
#
# Usage (repo root):
#   ./scripts/guardian-power.sh sleep   # sudo systemctl stop ollama
#   ./scripts/guardian-power.sh wake    # sudo systemctl start ollama
#   ./scripts/guardian-power.sh status  # systemctl is-active ollama
#
# Optional sudoers (document only — edit at your own risk):
#   deploy ALL=(root) NOPASSWD: /bin/systemctl start ollama, /bin/systemctl stop ollama, /bin/systemctl is-active ollama
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: scripts/guardian-power.sh {sleep|wake|status}

  sleep   Stop the Ollama systemd service (admin — saves more power than Rest now)
  wake    Start the Ollama systemd service
  status  Print whether ollama.service is active

Guardian UI "Rest now" unloads the chat model only — use this script when you
also want the Ollama process stopped (e.g. overnight on solar power).
EOF
}

cmd="${1:-}"
case "$cmd" in
  sleep)
    sudo systemctl stop ollama
    echo "ollama.service stopped"
    ;;
  wake)
    sudo systemctl start ollama
    echo "ollama.service started — open Guardian and tap Awaken now"
    ;;
  status)
    systemctl is-active ollama || true
    ;;
  -h|--help|help|"")
    usage
    [[ -n "$cmd" ]] || exit 0
    ;;
  *)
    echo "unknown command: $cmd" >&2
    usage >&2
    exit 1
    ;;
esac
