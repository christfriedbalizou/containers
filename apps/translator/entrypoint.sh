#!/bin/sh
set -eu

if [ "$#" -eq 0 ]; then set -- api; fi
case "$1" in
    api|worker|migrate)
        # Serialize schema upgrades across API and worker containers/replicas.
        python - <<'PY'
import asyncio
import fcntl
import os
import subprocess
import sys

from sqlalchemy import text
from sqlalchemy.engine import make_url
from sqlalchemy.ext.asyncio import create_async_engine
from sqlalchemy.pool import NullPool

url = make_url(os.environ["DATABASE_URL"])
command = [sys.executable, "-m", "translator.operations", "migrate"]

async def migrate_postgres():
    engine = create_async_engine(url, poolclass=NullPool)
    try:
        async with engine.connect() as connection:
            await connection.execute(text("SELECT pg_advisory_lock(824637190421)"))
            try:
                process = await asyncio.create_subprocess_exec(*command)
                return await process.wait()
            finally:
                await connection.execute(text("SELECT pg_advisory_unlock(824637190421)"))
    finally:
        await engine.dispose()

if url.get_backend_name() == "postgresql":
    sys.exit(asyncio.run(migrate_postgres()))
if url.get_backend_name() == "sqlite" and url.database not in (None, "", ":memory:"):
    with open(url.database + ".migrate.lock", "a") as lock:
        fcntl.flock(lock, fcntl.LOCK_EX)
        sys.exit(subprocess.call(command))
raise RuntimeError("startup_migrations_require_postgresql_or_file_sqlite")
PY
        ;;
esac
case "$1" in
    api) shift; exec python -m translator.server "$@" ;;
    worker) shift; exec python -m translator.jobs.command "$@" ;;
    migrate) exit 0 ;;
    retention) shift; exec python -m translator.retention "$@" ;;
    *) exec "$@" ;;
esac
