#!/usr/bin/env sh
set -eu

: "${CODE_SERVER_BIND_ADDR:=0.0.0.0:8080}"
: "${CODE_SERVER_AUTH:=password}"
: "${CODE_SERVER_DEFAULTS_DIR:=/opt/code-server-defaults}"
: "${CODE_SERVER_USER_DATA_DIR:=/config/data}"
: "${CODE_SERVER_EXTENSIONS_DIR:=/config/extensions}"
: "${HOME:=/home/coder}"
: "${CODEX_AUTO_LOGIN:=false}"
: "${CODEX_HOME:=${HOME}/.codex}"
: "${CODEX_UNSET_API_KEY_AFTER_LOGIN:=true}"
: "${GIT_USER_NAME:=}"
: "${GIT_USER_EMAIL:=}"

export CODEX_HOME

if [ "${CODE_SERVER_AUTH}" = "password" ] \
  && [ -z "${PASSWORD:-}" ] \
  && [ -z "${HASHED_PASSWORD:-}" ]; then
  echo "Set PASSWORD or HASHED_PASSWORD when CODE_SERVER_AUTH=password." >&2
  exit 1
fi

mkdir -p \
  "${CODE_SERVER_USER_DATA_DIR}/User" \
  "${CODE_SERVER_EXTENSIONS_DIR}" \
  "${CODEX_HOME}" \
  "${HOME}/.cache" \
  "${HOME}/.config" \
  "${HOME}/.local/share" \
  /workspace

touch "${HOME}/.bashrc" "${HOME}/.bash_profile"

if ! grep -Fq 'mise activate bash' "${HOME}/.bashrc" 2>/dev/null; then
  printf '%s\n' 'eval "$(mise activate bash)"' >> "${HOME}/.bashrc"
fi

if ! grep -Fq '. ~/.bashrc' "${HOME}/.bash_profile" 2>/dev/null; then
  printf '%s\n' '[ -f ~/.bashrc ] && . ~/.bashrc' >> "${HOME}/.bash_profile"
fi

if command -v git >/dev/null 2>&1; then
  if [ -n "${GIT_USER_NAME}" ]; then
    git config --global user.name "${GIT_USER_NAME}"
  fi

  if [ -n "${GIT_USER_EMAIL}" ]; then
    git config --global user.email "${GIT_USER_EMAIL}"
  fi
fi

if [ -f "${CODE_SERVER_DEFAULTS_DIR}/settings.json" ]; then
  cp "${CODE_SERVER_DEFAULTS_DIR}/settings.json" "${CODE_SERVER_USER_DATA_DIR}/User/settings.json"
fi

if [ -f "${CODE_SERVER_DEFAULTS_DIR}/extensions.txt" ]; then
  while read -r extension; do
    [ -z "${extension}" ] && continue
    case "${extension}" in
      \#*) continue ;;
    esac

    if ! code-server \
      --user-data-dir "${CODE_SERVER_USER_DATA_DIR}" \
      --extensions-dir "${CODE_SERVER_EXTENSIONS_DIR}" \
      --install-extension "${extension}"; then
      echo "Failed to install extension ${extension}; continuing." >&2
    fi
  done < "${CODE_SERVER_DEFAULTS_DIR}/extensions.txt"
fi

if [ "${CODEX_AUTO_LOGIN}" = "true" ] && [ -n "${OPENAI_API_KEY:-}" ]; then
  if command -v codex >/dev/null 2>&1; then
    if printf '%s\n' "${OPENAI_API_KEY}" | codex login --with-api-key >/dev/null 2>&1; then
      if [ "${CODEX_UNSET_API_KEY_AFTER_LOGIN}" = "true" ]; then
        unset OPENAI_API_KEY
      fi
    else
      echo "Failed to log Codex in with OPENAI_API_KEY; continuing." >&2
    fi
  else
    echo "CODEX_AUTO_LOGIN=true but codex is not installed; continuing." >&2
  fi
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
