#!/usr/bin/env sh
set -eu

if [ "$#" -eq 0 ]; then
  set -- server /data --console-address :9001
fi

# Accept both the image's existing argument convention and upstream's.
if [ "$1" = minio ] || [ "$1" = silo ]; then
  shift
fi

if [ "${1:-}" = server ]; then
  # Resolve Docker/Kubernetes secrets without printing their contents.
  if [ -n "${MINIO_ROOT_USER_FILE:-}" ]; then
    [ -z "${MINIO_ROOT_USER:-}" ] || { echo 'Set MINIO_ROOT_USER or MINIO_ROOT_USER_FILE, not both.' >&2; exit 1; }
    MINIO_ROOT_USER=$(cat "$MINIO_ROOT_USER_FILE")
    export MINIO_ROOT_USER
    unset MINIO_ROOT_USER_FILE
  fi
  if [ -n "${MINIO_ROOT_PASSWORD_FILE:-}" ]; then
    [ -z "${MINIO_ROOT_PASSWORD:-}" ] || { echo 'Set MINIO_ROOT_PASSWORD or MINIO_ROOT_PASSWORD_FILE, not both.' >&2; exit 1; }
    MINIO_ROOT_PASSWORD=$(cat "$MINIO_ROOT_PASSWORD_FILE")
    export MINIO_ROOT_PASSWORD
    unset MINIO_ROOT_PASSWORD_FILE
  fi
  if [ -z "${MINIO_ROOT_USER:-}" ] || [ -z "${MINIO_ROOT_PASSWORD:-}" ] || [ "$MINIO_ROOT_PASSWORD" = minioadmin ]; then
    echo 'Configure explicit, non-default MINIO_ROOT_USER and MINIO_ROOT_PASSWORD credentials (or their _FILE variants).' >&2
    exit 1
  fi
  umask 077
fi

exec /usr/local/bin/minio "$@"
