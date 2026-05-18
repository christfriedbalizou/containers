#!/usr/bin/env sh
set -eu

: "${CODE_SERVER_BIND_ADDR:=0.0.0.0:8080}"
: "${CODE_SERVER_AUTH:=password}"
: "${CODE_SERVER_DEFAULTS_DIR:=/opt/code-server-defaults}"
: "${CODE_SERVER_USER_DATA_DIR:=/config/data}"
: "${CODE_SERVER_EXTENSIONS_DIR:=/config/extensions}"

if [ "${CODE_SERVER_AUTH}" = "password" ] \
  && [ -z "${PASSWORD:-}" ] \
  && [ -z "${HASHED_PASSWORD:-}" ]; then
  echo "Set PASSWORD or HASHED_PASSWORD when CODE_SERVER_AUTH=password." >&2
  exit 1
fi

mkdir -p "${CODE_SERVER_USER_DATA_DIR}/User" "${CODE_SERVER_EXTENSIONS_DIR}" /workspace

if [ -f "${CODE_SERVER_DEFAULTS_DIR}/settings.json" ]; then
  cp "${CODE_SERVER_DEFAULTS_DIR}/settings.json" "${CODE_SERVER_USER_DATA_DIR}/User/settings.json"
fi

if [ -f "${CODE_SERVER_DEFAULTS_DIR}/extensions.txt" ]; then
  while read -r extension; do
    [ -z "${extension}" ] && continue
    code-server \
      --user-data-dir "${CODE_SERVER_USER_DATA_DIR}" \
      --extensions-dir "${CODE_SERVER_EXTENSIONS_DIR}" \
      --install-extension "${extension}"
  done < "${CODE_SERVER_DEFAULTS_DIR}/extensions.txt"
fi

if [ "$#" -eq 0 ]; then
  exec code-server \
    --bind-addr "${CODE_SERVER_BIND_ADDR}" \
    --auth "${CODE_SERVER_AUTH}" \
    --user-data-dir "${CODE_SERVER_USER_DATA_DIR}" \
    --extensions-dir "${CODE_SERVER_EXTENSIONS_DIR}" \
    /workspace
else
  exec code-server \
    --user-data-dir "${CODE_SERVER_USER_DATA_DIR}" \
    --extensions-dir "${CODE_SERVER_EXTENSIONS_DIR}" \
    "$@"
fi
