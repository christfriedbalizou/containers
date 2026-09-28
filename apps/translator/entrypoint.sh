#!/bin/sh
set -eu

if [ "$#" -eq 0 ]; then set -- api; fi
case "$1" in
    api) shift; exec python -m translator.server "$@" ;;
    worker) shift; exec python -m translator.jobs.command "$@" ;;
    migrate) shift; exec python -m translator.operations migrate "$@" ;;
    retention) shift; exec python -m translator.retention "$@" ;;
    *) exec "$@" ;;
esac
