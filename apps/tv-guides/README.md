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
bash fetch-source.sh v0.3.0 source
docker buildx bake image-local
```

For a clean local tag export, set `SOURCE_DIR` to that directory and include a
`.source-revision` file containing the full source commit ID. `VERSION` must
match the package version; the build checks it.

## Persistent state

The image runs as UID/GID 1000. Mount a writable directory at `/data`. SQLite,
`config.json`, inventory reports and generated M3U/XMLTV files live directly
there. The entrypoint sets a private file creation mask and starts the
application; it does not create a nested state directory or modify the mount's
ownership or permissions. NFS access follows the server's identity mapping.
Run only one worker per data directory.

Set `TV_GUIDES_WORK_DIR` to another mounted path when needed; its default is
`/data`. The default command is `run`; use `refresh` for a single run.

```sh
mkdir -p state
docker run --rm \
  --mount type=bind,src="$PWD/state",dst=/data \
  --mount type=bind,src="$PWD/playlist-url",dst=/run/secrets/playlist-url,readonly \
  -e TV_GUIDES_WORK_DIR=/data \
  -e TV_GUIDES_PLAYLIST_URL_FILE=/run/secrets/playlist-url \
  tv-guides:v0.3.0 refresh
```

Create `playlist-url` with the provider's full M3U URL and make it readable by
UID 1000. `TV_GUIDES_PLAYLIST_URL` is available when a URL file cannot be used.
`TV_GUIDES_INTERVAL_SECONDS` sets the refresh interval, defaulting to 43200.
The upstream README documents profiles, mapping rules and backup/restore.

Matched output is published through `/data/current/playlist.m3u` and
`/data/current/guide.xml`. When upgrading from the previous `/data/state`
default, stop the worker and move the existing state contents to the mount
root before starting this layout. Update consumers' file paths accordingly.
The M3U contains private stream URLs. Keep the mount private and back up the
database before replacing state. No Jellyfin or homelab configuration is
installed by the image.

Run the container contract tests from the packaging repository root:

```sh
TEST_IMAGE=tv-guides:v0.3.0 go test -v ./apps/tv-guides/...
```

## Logs

The application writes UTC console logs to stderr. `TV_GUIDES_LOG_LEVEL`
defaults to `INFO`; set it to `DEBUG` for stage and scheduling details, or
`ERROR` for failures only. Logs show refresh counts, duration and safe failure
context without provider URLs or raw exception text. No log-file mount is
needed. When deployed in Kubernetes, read them through `kubectl logs`.

Automatic US/UK discovery is enabled by default. The working directory also
holds monthly identity caches, resumable schedule downloads and
`reports/guide-coverage.csv`, which explains unresolved channels. Set
`discover_guides` to `false` in the persisted `config.json` to use only explicit
source profiles. A first discovery run can take tens of minutes.

Confirmed live channels from other countries are included without invented
programme listings. The global identity catalogue is cached for 30 days.
`reports/group-issues.csv` and `reports/playlist-issues.csv` explain guide gaps,
excluded identities and what information would help resolve them. Matching
uses deterministic lookups and makes no AI API calls.
