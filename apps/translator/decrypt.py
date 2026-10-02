"""Unlock the image payload into a private, per-container tmpfs."""

import fcntl
import os
from pathlib import Path
import shutil
import subprocess
import tarfile
import tempfile


ROOT = Path("/opt/translator")
PAYLOAD = Path("/usr/share/translator/payload.tar.gz.gpg")


def extract_filter(member, destination):
    # uv's interpreter links intentionally point outside the virtual environment.
    if (
        member.name in {".venv/bin/python", ".venv/bin/python3", ".venv/bin/python3.12"}
        and member.issym()
        and member.linkname in {"python", "python3", "python3.12", "/usr/local/bin/python", "/usr/local/bin/python3", "/usr/local/bin/python3.12"}
    ):
        return tarfile.tar_filter(member, destination)
    return tarfile.data_filter(member, destination)


def unlock():
    os.umask(0o077)
    mounts = [line.split() for line in Path("/proc/mounts").read_text().splitlines()]
    if not any(m[1] == str(ROOT) and m[2] == "tmpfs" and "noexec" not in m[3].split(",") for m in mounts):
        raise RuntimeError("mount an executable, UID 1000-owned tmpfs at /opt/translator")

    key = Path(os.environ.get("TRANSLATOR_GPG_PRIVATE_KEY_FILE", "/run/secrets/translator-private-key"))
    if not key.is_file():
        raise RuntimeError(f"missing Translator decryption key: {key}")

    # Multiple entrypoint invocations in one container share a single extraction.
    with (ROOT / ".decrypt.lock").open("a") as lock:
        fcntl.flock(lock, fcntl.LOCK_EX)
        if (ROOT / ".decrypted").exists():
            return
        with tempfile.TemporaryDirectory(prefix=".decrypt-", dir=ROOT) as work:
            work = Path(work)
            home = work / "gnupg"
            home.mkdir(mode=0o700)
            gpg = ["gpg", "--no-options", "--homedir", str(home), "--batch", "--no-tty"]
            try:
                subprocess.run([*gpg, "--import", str(key)], check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
                passphrase = os.environ.get("TRANSLATOR_GPG_PASSPHRASE_FILE")
                password_args = ["--passphrase-file", passphrase] if passphrase else ["--passphrase", ""]
                archive = work / "payload.tar.gz"
                # Verify GPG's integrity check before extracting any application files.
                subprocess.run(
                    [*gpg, "--pinentry-mode", "loopback", *password_args,
                     "--output", str(archive), "--decrypt", str(PAYLOAD)],
                    check=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL,
                )
                staging = work / "app"
                staging.mkdir()
                with tarfile.open(archive, "r:gz") as bundle:
                    bundle.extractall(staging, filter=extract_filter)
                for name in (".venv", "migrations", "alembic.ini"):
                    target = ROOT / name
                    if target.is_dir():
                        shutil.rmtree(target)
                    elif target.exists():
                        target.unlink()
                    (staging / name).rename(target)
                (ROOT / "dist").unlink(missing_ok=True)
                (ROOT / "dist").symlink_to("/usr/share/translator/dist", target_is_directory=True)
                (ROOT / ".decrypted").touch(mode=0o600)
            finally:
                subprocess.run(["gpgconf", "--homedir", str(home), "--kill", "gpg-agent"], check=False)


if __name__ == "__main__":
    try:
        unlock()
    except (OSError, RuntimeError, subprocess.SubprocessError, tarfile.TarError) as error:
        raise SystemExit(f"Translator decryption failed: {error}") from None
