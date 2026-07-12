#!/usr/bin/env bash
# Cron-friendly wrappers for scripts/guardian-power.sh (Phase 163 WS4).
#
# Use for overnight / solar sites that stop the Ollama *service* (not just Rest now).
# The web UI never runs systemctl — schedule this on the host where Ollama runs.
#
# Usage (repo root or absolute path in crontab):
#   ./scripts/guardian-power-schedule.sh print-crontab
#   ./scripts/guardian-power-schedule.sh cron-sleep    # typical 22:00 cron target
#   ./scripts/guardian-power-schedule.sh cron-wake     # typical 06:00 cron target
#
# Optional log file (default: /var/log/gr33n-guardian-power.log — falls back to $HOME)
#   GUARDIAN_POWER_LOG=/var/log/gr33n-guardian-power.log
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
POWER="${ROOT}/scripts/guardian-power.sh"
LOG="${GUARDIAN_POWER_LOG:-/var/log/gr33n-guardian-power.log}"

usage() {
  cat <<EOF
Usage: scripts/guardian-power-schedule.sh <command>

Commands:
  print-crontab   Show example crontab lines (edit times + user)
  print-sudoers   Show narrow NOPASSWD sudoers snippet for cron (optional)
  cron-sleep      Log + run guardian-power.sh sleep (for cron)
  cron-wake       Log + run guardian-power.sh wake (for cron)
  -h, --help      This message

Patterns:
  Daytime:  GUARDIAN_AUTO_DORMANT_MINUTES in .env (API) — no sudo, Ollama stays up
  Night:    cron-sleep at 22:00, cron-wake at 06:00 on the inference host

After cron-wake, operators still tap Awaken now in Settings (or send chat) to load the model.
EOF
}

log_line() {
  local msg="[$(date -Iseconds)] $*"
  if [[ -w "$(dirname "$LOG")" ]] 2>/dev/null; then
    echo "$msg" >>"$LOG"
  elif [[ -n "${HOME:-}" ]]; then
    echo "$msg" >>"${HOME}/.gr33n-guardian-power.log"
  fi
  echo "$msg"
}

run_cron() {
  local action="$1"
  log_line "guardian-power-schedule: starting $action"
  if ! "$POWER" "$action"; then
    log_line "guardian-power-schedule: $action failed (exit $?)"
    exit 1
  fi
  log_line "guardian-power-schedule: $action ok"
}

print_crontab() {
  local user
  user="$(whoami)"
  cat <<EOF
# gr33n Guardian — Ollama service sleep/wake on the INFERENCE host (not the browser machine).
# Install: crontab -e  (or /etc/cron.d/gr33n-guardian-power on multi-user servers)
#
# Adjust times for your solar / quiet hours. Use absolute paths.

# Stop Ollama service at 22:00 (deep sleep — saves more than Rest now)
0 22 * * * ${user} ${ROOT}/scripts/guardian-power-schedule.sh cron-sleep

# Start Ollama service at 06:00 (model still cold until Awaken now or first chat)
0 6 * * * ${user} ${ROOT}/scripts/guardian-power-schedule.sh cron-wake

# Optional: status ping every morning (no sudo)
# 5 6 * * * ${user} ${ROOT}/scripts/guardian-power.sh status >> ${LOG} 2>&1

# Daytime RAM saving without stopping Ollama: set in API .env instead of cron:
#   GUARDIAN_AUTO_DORMANT_MINUTES=45
EOF
}

print_sudoers() {
  local user
  user="$(whoami)"
  cat <<EOF
# Optional — lets cron run sleep/wake without a password prompt.
# Edit with: sudo visudo -f /etc/sudoers.d/gr33n-guardian-power
# Replace ${user} with the cron user (often the same user that runs Ollama).

${user} ALL=(root) NOPASSWD: /bin/systemctl start ollama, /bin/systemctl stop ollama, /bin/systemctl is-active ollama

# Narrow scope only — do NOT grant general sudo or a web-facing password form.
EOF
}

cmd="${1:-}"
case "$cmd" in
  cron-sleep) run_cron sleep ;;
  cron-wake)  run_cron wake ;;
  print-crontab) print_crontab ;;
  print-sudoers) print_sudoers ;;
  -h|--help|help|"") usage; [[ -n "$cmd" ]] || exit 0 ;;
  *)
    echo "unknown command: $cmd" >&2
    usage >&2
    exit 1
    ;;
esac
