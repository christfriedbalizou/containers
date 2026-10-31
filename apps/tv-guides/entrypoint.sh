#!/bin/sh
set -eu
umask 077
exec /opt/tv-guides/.venv/bin/tv-guides "$@"
