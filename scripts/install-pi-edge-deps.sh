#!/usr/bin/env bash
# Minimal OS packages on Raspberry Pi OS for the edge daemon (sensor client / MQTT tooling).
# Raspberry Pi OS is Debian-based; uses sudo apt — you may be prompted for your password.
#
# Does NOT install Postgres, Node, Go, or the dashboard — only what a headless Pi typically needs
# before running pi_client/setup.sh from a clone of this repo.
#
# Usage:
#   ./scripts/install-pi-edge-deps.sh
#   ./scripts/install-pi-edge-deps.sh --with-docker    # optional: Docker Engine + Compose v2 (full stack on Pi)
#
set -euo pipefail

WITH_DOCKER=0
while [[ $# -gt 0 ]]; do
  case "$1" in
    --with-docker) WITH_DOCKER=1 ;;
    -h|--help)
      grep -E '^#(|  )' "$0" | sed 's/^# \{0,1\}//'
      exit 0
      ;;
    *)
      echo "Unknown option: $1" >&2
      exit 1
      ;;
  esac
  shift || true
done

if [[ "${EUID:-}" -eq 0 ]]; then
  echo "error: run as a normal user; this script invokes sudo." >&2
  exit 1
fi

if [[ ! -f /etc/debian_version ]]; then
  echo "error: intended for Debian-based Pi OS — see docs/raspberry-pi-and-deployment-topology.md." >&2
  exit 1
fi

echo "==> apt: Python venv + GPIO helpers + git (clone/updates)"
sudo apt-get update -qq
sudo apt-get install -y ca-certificates curl git python3 python3-pip python3-venv libgpiod2 i2c-tools

if [[ "$WITH_DOCKER" -eq 1 ]]; then
  echo "==> apt: Docker Engine + Compose plugin (optional full stack — see topology doc)"
  sudo apt-get install -y docker.io docker-compose-v2
  sudo systemctl enable --now docker || true
  echo "Tip: add your user to group 'docker' if you run compose without sudo: sudo usermod -aG docker \"\$USER\" (then re-login)."
fi

echo ""
echo "Done."
echo "  • Clone this repo on the Pi (or rsync pi_client/), then:"
echo "      cd pi_client && ./setup.sh"
echo "  • Configure pi_client/config.yaml — base_url must reach your API (LAN IP or hostname)."
echo "  • Full layouts and scaling: docs/raspberry-pi-and-deployment-topology.md"
