#!/usr/bin/env sh

set -e

PUID=${PUID:-0}
PGID=${PGID:-0}

log() {
    printf "%s\n" "==> $*"
}

files_ownership() {
    log "Changing ownership of /app/data to ${PUID}:${PGID}"
    chown -hRc "${PUID}":"${PGID}" /app/data
}

log "Performing startup jobs and maintenance tasks"
files_ownership

log "Starting application with user ${PUID} group ${PGID}"

exec setpriv --reuid "${PUID}" --regid "${PGID}" --clear-groups "$@"