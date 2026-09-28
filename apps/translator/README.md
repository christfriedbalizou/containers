# Translator

Builds the private `christfried.balizou/translator` Forgejo repository at the
stable release tag pinned in `docker-bake.hcl`. The image is published as
`ghcr.io/christfriedbalizou/translator` with version tags and `rolling` by the
existing release workflow. Linux amd64 follows the upstream operating instructions.

## Release automation

Set these GitHub Actions secrets on **this containers repository**:

- `SECRET_DOMAIN`: the base domain; checkout uses `https://git.${SECRET_DOMAIN}`.
- `FORGEJO_USERNAME`: the read-only token owner's username.
- `FORGEJO_ACCESS_TOKEN`: `read:repository` access to Translator.

The shared `ChristfriedBalizou/.github` Renovate workflow also needs
`SECRET_DOMAIN` and `FORGEJO_ACCESS_TOKEN`. Its runner configuration supplies
a `forgejo-tags` host rule; repository secrets are not inherited across the
dispatch boundary. The existing Renovate schedule/dispatch discovers new tags.
The `Renovate` workflow here can trigger a scan manually.

The private hostname is resolved only in the authenticated runners from
`SECRET_DOMAIN`; it is not stored in Bake or repository Renovate configuration.
Published image labels point to this public packaging repository. Renovate
version updates omit private source links and changelogs from their PR bodies.

Renovate opens an update to `VERSION`. Existing policy automatically merges
minor/patch updates after checks pass; majors require review. Merging the update
triggers a release build. PR builds publish `sandbox` for smoke testing.
The checkout fetches the exact tag and exports committed files only. Git
credentials and `.git` never enter a build context or image. The source commit
is recorded in the OCI revision label and `/usr/share/translator/source-revision`.

## Build locally

Export the three secrets into the shell without adding them to tracked files:

```sh
cd apps/translator
version=$(docker buildx bake --list type=variables,format=json | jq -r '.[] | select(.name == "VERSION") | .value')
bash fetch-source.sh "$version" source
docker buildx bake image-local
```

The source destination must not already exist. For another version, pass the
same version to checkout and Bake: `VERSION=v0.3.0 docker buildx bake image-local`.
To build from an existing clean tag export, point `SOURCE_DIR` at the exported
directory and add `.source-revision` containing the full commit ID.

## Runtime

The default command is `api` (`python -m translator.server`) on port 8000.
Use the same image with command `worker`, `migrate`, or `retention` for those
roles. Other commands are executed directly. Run migrations as a separate job
before starting API/worker replicas; startup does not mutate the database schema.

The image runs as UID/GID 1000, with root-owned application code at
`/opt/translator`. Supply the upstream configuration and mounted secret files;
no credentials or production configuration are built in. Set the database URL
and storage paths explicitly, using `/data` for writable persistent local data.
`/health` is API liveness and `/ready` checks schema/database/keys/storage.
Configure health probes per role rather than an API-only image healthcheck.

Use the existing homelab Authelia, database, storage and model services. Provision
and mount the engine/OCR assets and corresponding-source bundle documented by
upstream. The image includes frontend assets and Alembic migrations; it does not
provision external services or download models at startup. Mount the reviewed
source archive outside `dist/` and set `SOURCE_DISTRIBUTION__ARCHIVE`,
`SOURCE_DISTRIBUTION__REVISION`, and `SOURCE_DISTRIBUTION__SHA256` together.

Workers require the upstream Linux Landlock ABI 6+ sandbox support, private
`/dev/shm` tmpfs, explicit resource limits and restricted network access.
PostgreSQL backup/restore commands additionally need client tools matching the
database server version; those operator tools are not bundled in this image.
See the upstream `docs/operations.md` and `docs/source-distribution.md` for
deployment requirements.
