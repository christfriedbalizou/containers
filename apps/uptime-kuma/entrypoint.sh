#!/usr/bin/env sh

set -e

# Environment defaults (PUID/PGID can be passed via -e when running the container)
PUID=${PUID:-0}
PGID=${PGID:-0}

log() {
    # Simple logger used for consistent output formatting
    printf "%s\n" "==> $*"
}

files_ownership() {
    # -h : change the ownership of symbolic links themselves
    # -R : recursive
    # -c : like verbose but report only when a change is made
    # Uptime Kuma stores its data in /app/data
    log "Changing ownership of /app/data to ${PUID}:${PGID}"
    chown -hRc "${PUID}":"${PGID}" /app/data
}

log "Performing startup jobs and maintenance tasks"
files_ownership

log "Starting application with user ${PUID} group ${PGID}"

if command -v setpriv >/dev/null 2>&1; then
    exec setpriv --reuid "${PUID}" --regid "${PGID}" --clear-groups "$@"
fi

if [ -x /usr/bin/dumb-init ]; then
    exec /usr/bin/dumb-init -- "$@"
fi