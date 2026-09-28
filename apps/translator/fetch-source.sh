#!/usr/bin/env bash
set -euo pipefail

version=${1:?Usage: fetch-source.sh VERSION DESTINATION}
destination=${2:?Usage: fetch-source.sh VERSION DESTINATION}
: "${SECRET_DOMAIN:?SECRET_DOMAIN is required}"
: "${FORGEJO_USERNAME:?FORGEJO_USERNAME is required}"
: "${FORGEJO_ACCESS_TOKEN:?FORGEJO_ACCESS_TOKEN is required}"

[[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || { echo 'Expected a stable vMAJOR.MINOR.PATCH tag.' >&2; exit 1; }
[[ "$SECRET_DOMAIN" =~ ^[a-zA-Z0-9]([a-zA-Z0-9.-]*[a-zA-Z0-9])?$ ]] || { echo 'SECRET_DOMAIN must be a hostname without a scheme or path.' >&2; exit 1; }
[[ ! -e "$destination" ]] || { echo 'Source destination already exists.' >&2; exit 1; }

checkout=$(mktemp -d)
trap 'rm -rf "$checkout"' EXIT
umask 077
cat > "$checkout/askpass" <<'EOF'
#!/bin/sh
case "$1" in
    *Username*) printf '%s\n' "$FORGEJO_USERNAME" ;;
    *Password*) printf '%s\n' "$FORGEJO_ACCESS_TOKEN" ;;
    *) exit 1 ;;
esac
EOF
chmod 700 "$checkout/askpass"
export GIT_ASKPASS="$checkout/askpass" GIT_TERMINAL_PROMPT=0
git init --quiet "$checkout/repo"
git -C "$checkout/repo" -c credential.helper= fetch --quiet --depth=1 \
    "https://git.${SECRET_DOMAIN}/christfried.balizou/translator.git" \
    "refs/tags/${version}:refs/tags/${version}"
revision=$(git -C "$checkout/repo" rev-parse "refs/tags/${version}^{commit}")
umask 022
mkdir -p "$destination"
# Export only committed files; credentials and Git metadata never enter Docker.
git -C "$checkout/repo" archive "$revision" | tar -x -C "$destination"
printf '%s\n' "$revision" > "$destination/.source-revision"
if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
    printf 'revision=%s\n' "$revision" >> "$GITHUB_OUTPUT"
fi
printf 'Exported Translator %s at %s\n' "$version" "$revision"
