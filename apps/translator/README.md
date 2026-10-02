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
- `TRANSLATOR_GPG_PUBLIC_KEY`: ASCII-armored public encryption key for builds.
- `TRANSLATOR_GPG_PRIVATE_KEY`: matching ASCII-armored private key, used only by
  the container test job.
- `TRANSLATOR_GPG_PASSPHRASE`: private-key passphrase for tests, if set.

These are repository secrets inherited by the reusable App Builder workflow.
The private key is written to a temporary file only in the test job, copied into
its test containers, and removed after tests. Keep secret-bearing PR test runs
restricted to trusted changes. No private key is passed to Docker builds.

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
export TRANSLATOR_GPG_PUBLIC_KEY="$(cat /secure/translator-public.asc)"
docker buildx bake image-local
```

The source destination must not already exist. For another version, pass the
same version to checkout and Bake: `VERSION=v0.6.0 docker buildx bake image-local`.
To build from an existing clean tag export, point `SOURCE_DIR` at the exported
directory and add `.source-revision` containing the full commit ID.

## Encryption keys and deployment

Generate a dedicated key on a trusted machine (GPG prompts for a passphrase):

```sh
gpg --quick-generate-key 'Translator image encryption' rsa4096 encr 0
gpg --list-keys --keyid-format long 'Translator image encryption'
# Replace FINGERPRINT with the full fingerprint from the output.
umask 077
gpg --armor --export FINGERPRINT > /secure/translator-public.asc
gpg --armor --export-secret-keys FINGERPRINT > /secure/translator-private.asc
```

Create `/secure` first or use another private directory outside this repository.
Upload the armored files as the matching GitHub secrets listed above. Store the
passphrase separately as `TRANSLATOR_GPG_PASSPHRASE` when using a protected key.
Keep an offline backup: images cannot be recovered without their private key.

The final image contains an encrypted archive of the entire installed Python
environment, Alembic configuration and migrations. Plaintext backend source exists in
trusted build stages and their local cache, but is never copied into a final
image layer. Registry cache export stays disabled for Translator. The encryption
stage always reruns so rotating the public key cannot reuse an old encrypted
payload. Protect build runners and their local caches.

Every role (including maintenance commands) requires a readable private-key file
and a fresh, executable tmpfs at `/opt/translator`, owned by UID/GID 1000. Example:

```sh
docker run --rm \
  --tmpfs /opt/translator:rw,exec,nosuid,nodev,size=4g,uid=1000,gid=1000,mode=0700 \
  --mount type=bind,src=/secure/translator-private.asc,dst=/run/secrets/translator-private-key,readonly \
  --mount type=bind,src=/secure/translator-passphrase,dst=/run/secrets/translator-passphrase,readonly \
  -e TRANSLATOR_GPG_PASSPHRASE_FILE=/run/secrets/translator-passphrase \
  --env-file /secure/translator.env \
  -p 8000:8000 translator:v0.6.0 api
```

Ensure UID 1000 can read the mounted secrets. Omit the passphrase mount and
variable for an unprotected key. `TRANSLATOR_GPG_PRIVATE_KEY_FILE` overrides the
default private-key path. Never pass the key or passphrase contents as container
environment variables. In Kubernetes use a memory-backed `emptyDir` for
`/opt/translator`, mount Secrets as files, and set ownership via the pod security
context. Each replica needs its own tmpfs.

Startup fails on missing keys, incorrect passphrases, corrupted ciphertext or a
missing/non-executable tmpfs. It verifies decryption before extracting files and
serializes concurrent entrypoint calls. Decrypted files and temporary GPG state
stay on tmpfs; the imported key and temporary archive are removed after startup.
The application remains decrypted there for the container's lifetime. Size the
mount and container memory for the expanded environment plus the compressed
archive during startup; 4 GiB is an initial limit to adjust for the release.
Disable host swap or encrypt it if plaintext must not reach disk: tmpfs may swap.
See [Docker tmpfs documentation](https://docs.docker.com/engine/storage/tmpfs/).

This protects an image obtained without the key. Administrators of the running
host can still access plaintext and keys. Browser frontend assets remain public.
Image SBOM/vulnerability scans cannot inspect encrypted Python dependencies;
scan those in the trusted upstream build. Previously published plaintext images
and existing copies remain readable. Rotation protects newly built images only;
retain old keys if you need to roll back to images encrypted with them.

For local container tests, export `TRANSLATOR_TEST_PRIVATE_KEY_FILE` and optionally
`TRANSLATOR_TEST_PASSPHRASE_FILE` with host file paths, then run from the repository root:

```sh
TEST_IMAGE=translator:v0.6.0 go test -v ./apps/translator/...
```

## Runtime

The default command is `api` (`python -m translator.server`) on port 8000.
Use the same image with command `worker`, `migrate`, or `retention` for those
roles. API and worker startup automatically migrate the database before executing
the application. PostgreSQL advisory locks (or a file lock for file-backed SQLite)
serialize migrations across containers. Migration failures stop startup. No init
container or separate migration job is required. Other commands execute directly.

Configure providers with `TRANSLATION__PROVIDERS` JSON and mount their API keys
using each entry's `api_key_file`. The API reconciles these settings at startup;
no deployment bootstrap script is needed. The first verified OIDC login on a
fresh database becomes the administrator. See upstream `docs/operations.md`
for provider configuration and S3 deletion permissions.

The image runs as UID/GID 1000, with application code decrypted into its private tmpfs at
`/opt/translator`. Supply the upstream configuration and mounted secret files;
no credentials or production configuration are built in. Set the database URL
and storage backend explicitly. `INGESTION__LOCAL_STORAGE_PATH` is required
only for local storage; omit it for S3 deployments.
`/health` is API liveness and `/ready` checks schema/database/keys/storage.
Configure health probes per role rather than an API-only image healthcheck.

Glossary authoring defaults to 2000 entries and 1 MiB per glossary. Configure
`GLOSSARY__MAX_ENTRIES` and `GLOSSARY__MAX_CONTENT_BYTES` with positive integers;
the API exposes the effective limits to the paginated editor. The combined
`INGESTION__MAX_GLOSSARY_SNAPSHOT_BYTES` budget remains independent (1 MiB by
default). Lowering authoring settings preserves existing records and backups,
but older application releases cannot read newly created oversized glossaries;
review stored content before rolling back.

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
