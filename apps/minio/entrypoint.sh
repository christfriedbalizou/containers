#!/usr/bin/env sh
set -eu

: "${MINIO_ROOT_USER:=minioadmin}"
: "${MINIO_ROOT_PASSWORD:=minioadmin}"

if [ "$#" -eq 0 ]; then
  exec /usr/local/bin/minio server /data --console-address ":9001"
else
  exec /usr/local/bin/minio "$@"
fi