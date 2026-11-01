#!/bin/sh
set -eu
umask 077
if [ "${TV_GUIDES_WORK_DIR:-/data/state}" = /data/state ]; then
    mkdir -p /data/state
    prepare_state() {
        set -- /data/config.json /data/guide-mappings.sqlite3* \
            /data/discovery /data/reports /data/generations \
            /data/current /data/.refresh.lock
        for entry in "$@"; do
            if [ ! -e "$entry" ] && [ ! -L "$entry" ]; then
                continue
            fi
            target="/data/state/${entry##*/}"
            if [ -e "$target" ] || [ -L "$target" ]; then
                echo "State conflict; refusing to overwrite guide data" >&2
                return 1
            fi
        done
        for entry in "$@"; do
            if [ -e "$entry" ] || [ -L "$entry" ]; then
                mv "$entry" /data/state/
            fi
        done
    }
    prepare_state
fi
exec /opt/tv-guides/.venv/bin/tv-guides "$@"
