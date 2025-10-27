#!/usr/bin/env sh

set -e

PUID=${PUID:-0}
PGID=${PGID:-0}

files_ownership () {
    # -h Changes the ownership of an encountered symbolic link and not that of the file or directory pointed to by the symbolic link.
    # -R Recursively descends the specified directories
    # -c Like verbose but report only when a change is made
    # The /app/data directory is where Uptime Kuma stores its configuration and database.
    echo "==> Changing ownership of /app/data to $PUID:$PGID"
    chown -hRc "$PUID":"$PGID" /app/data
}

echo "==> Performing startup jobs and maintenance tasks"
files_ownership

echo "==> Starting application with user $PUID group $PGID"

exec setpriv --reuid "$PUID" --regid "$PGID" --clear-groups /usr/bin/dumb-init -- "$@"