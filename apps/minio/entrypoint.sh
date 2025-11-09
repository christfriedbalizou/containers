#!/usr/bin/env sh
set -eu

: "${MINIO_ROOT_USER:=minioadmin}"
: "${MINIO_ROOT_PASSWORD:=minioadmin}"

setup_mc() {
  local retries=30
  while [ $retries -gt 0 ] && ! /usr/local/bin/mc ping local >/dev/null 2>&1; do
    sleep 1
    retries=$((retries - 1))
  done
  
  /usr/local/bin/mc alias set local http://localhost:9000 "${MINIO_ROOT_USER}" "${MINIO_ROOT_PASSWORD}" >/dev/null 2>&1 || true
}

if [ "$#" -eq 0 ]; then
  /usr/local/bin/minio server /data --console-address ":9001" &
else
  /usr/local/bin/minio "$@" &
fi

MINIO_PID=$!

setup_mc

wait $MINIO_PID