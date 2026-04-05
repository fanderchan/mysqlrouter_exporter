#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${script_dir}"

if [[ ! -x build/mysqlrouter_exporter ]]; then
  sh ./build.sh
fi

install -d /etc/mysqlrouter_exporter
install -d /usr/local/bin
install -m 0755 build/mysqlrouter_exporter /usr/local/bin/mysqlrouter_exporter

if [[ ! -f /etc/mysqlrouter_exporter/config.yml ]]; then
  install -m 0640 config.example.yml /etc/mysqlrouter_exporter/config.yml
fi

install -d /etc/systemd/system
install -m 0644 contrib/systemd/mysqlrouter_exporter.service /etc/systemd/system/mysqlrouter_exporter.service

echo "installed /usr/local/bin/mysqlrouter_exporter"
echo "review /etc/mysqlrouter_exporter/config.yml before starting the service"

