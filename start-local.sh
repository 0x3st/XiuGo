#!/bin/zsh
set -e

cd "${0:A:h}"

if [[ ! -f manifest/config/config.yaml ]]; then
  echo "Missing manifest/config/config.yaml — copy from config.example.yaml first."
  exit 1
fi

if ! /opt/homebrew/bin/mysqladmin ping -h 127.0.0.1 --silent >/dev/null 2>&1; then
  /opt/homebrew/bin/brew services start mysql >/dev/null 2>&1 || true
fi

echo "XiuGo 5.0.0  http://127.0.0.1:8081"
echo "Admin         http://127.0.0.1:8081/admin"
echo "Ctrl+C to stop."
exec /opt/homebrew/bin/go run .
