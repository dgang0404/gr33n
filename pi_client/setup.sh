#!/usr/bin/env bash
# One-time setup for Raspberry Pi OS
set -e

echo '==> Creating state directory'
sudo mkdir -p /var/lib/gr33n
sudo chown pi:pi /var/lib/gr33n

echo '==> Installing system dependencies'
sudo apt-get update -qq
sudo apt-get install -y python3-pip python3-venv libgpiod2 i2c-tools

echo '==> Creating Python virtual environment'
python3 -m venv venv
source venv/bin/activate

echo '==> Installing Python packages'
pip install --upgrade pip
pip install -r requirements.txt

echo '==> Installing systemd service'
sudo cp gr33n.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable gr33n

echo ''
echo 'Done! Edit config.yaml then: sudo systemctl start gr33n'
echo 'Tail logs: journalctl -u gr33n -f'
