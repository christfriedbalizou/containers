# TV guides

Builds `christfried.balizou/tv-guides` from the stable Forgejo release tag in
`docker-bake.hcl`. The image is ordinary, unencrypted Python code, published as
`ghcr.io/christfriedbalizou/tv-guides` with version tags and `rolling`.

## Source and release updates

The shared Renovate runner watches Forgejo tags. It updates `VERSION`, and a
merged packaging change triggers the existing image build. TV guides updates,
including major and digest updates, may auto-merge after checks pass.

Checkout uses the existing `SECRET_DOMAIN`, `FORGEJO_USERNAME`, and
`FORGEJO_ACCESS_TOKEN` repository secrets. The read token must cover TV guides.
The shared Renovate runner also needs its existing Forgejo token to read this
repository's tags. No encryption key is needed.

Source is exported from the exact tag into a named build context. Git metadata
and checkout credentials stay outside Docker. The source commit and tag are
recorded under `/usr/share/tv-guides`. Registry cache export is disabled for
private source builds, matching Translator's packaging boundary.

## Local build

From this directory, provide the checkout credentials through the environment:

```sh
bash fetch-source.sh v0.1.0 source
docker buildx bake image-local
```

For a clean local tag export, set `SOURCE_DIR` to that directory and include a
`.source-revision` file containing the full source commit ID. `VERSION` must
match the package version; the build checks it.

## Persistent state

The image runs as UID/GID 1000. Mount a private writable directory at `/data`.
It holds SQLite, `config.json`, inventory reports and generated M3U/XMLTV files.
Set `TV_GUIDES_WORK_DIR` to another mounted path when needed. The default
command is `run`; use `refresh` for a single run.

```sh
mkdir -m 700 state
docker run --rm \
  --mount type=bind,src="$PWD/state",dst=/data \
  --mount type=bind,src="$PWD/playlist-url",dst=/run/secrets/playlist-url,readonly \
  -e TV_GUIDES_WORK_DIR=/data \
  -e TV_GUIDES_PLAYLIST_URL_FILE=/run/secrets/playlist-url \
  tv-guides:v0.1.0 refresh
```

Create `playlist-url` with the provider's full M3U URL and make it readable by
UID 1000. `TV_GUIDES_PLAYLIST_URL` is available when a URL file cannot be used.
`TV_GUIDES_INTERVAL_SECONDS` sets the refresh interval, defaulting to 43200.
The upstream README documents profiles, mapping rules and backup/restore.

Matched output is published through `/data/current/playlist.m3u` and
`/data/current/guide.xml`. The M3U contains private stream URLs. Keep the mount
private and back up the database before replacing state. No Jellyfin or homelab
configuration is installed by the image.

Run the container contract tests from the packaging repository root:

```sh
TEST_IMAGE=tv-guides:v0.1.0 go test -v ./apps/tv-guides/...
```
