package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/christfriedbalizou/containers/testhelpers"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRuntimeContents(t *testing.T) {
	image := testhelpers.GetTestImage("ghcr.io/christfriedbalizou/translator:rolling")
	testCommandSucceeds(t, context.Background(), image, nil, "sh", "-ec", `
/usr/local/bin/python /usr/local/lib/translator-decrypt.py

test "$(id -u)" = 1000
test -f /opt/translator/.decrypted
test ! -d /opt/translator/.git
test ! -f /opt/translator/.env
test -f dist/index.html
test -f alembic.ini
test -d migrations/versions
python -c 'import ctypes, importlib.metadata, pathlib; import translator.server, translator.jobs.command, rapidocr, onnxruntime, cv2; ctypes.CDLL("libseccomp.so.2"); assert importlib.metadata.version("translator") == pathlib.Path("/usr/share/translator/source-tag").read_text().strip().removeprefix("v"); assert len(pathlib.Path("/usr/share/translator/source-revision").read_text().strip()) == 40'
`)
}

func TestAPIAndMigrations(t *testing.T) {
	ctx := context.Background()
	image := testhelpers.GetTestImage("ghcr.io/christfriedbalizou/translator:rolling")
	c, err := testcontainers.Run(ctx, image,
		testcontainers.WithExposedPorts("8000/tcp"),
		translatorSecrets(t),
		testcontainers.WithEnv(map[string]string{
			"DATABASE_URL":                     "sqlite+aiosqlite:////tmp/translator-test.sqlite3",
			"OIDC_ISSUER":                      "https://issuer.invalid",
			"OIDC_CLIENT_ID":                   "container-smoke-test",
			"OIDC_CLIENT_SECRET_FILE":          "/tmp/oidc-secret",
			"OIDC_REDIRECT_URI":                "https://translator.invalid/api/v1/auth/callback",
			"SESSION_IDLE_TIMEOUT_SECONDS":     "900",
			"SESSION_ABSOLUTE_TIMEOUT_SECONDS": "3600",
		}),
		testcontainers.WithEntrypoint("sh"),
		testcontainers.WithCmd("-ec", "printf smoke-test > /tmp/oidc-secret; exec entrypoint.sh api"),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/health").WithPort("8000/tcp").WithStartupTimeout(90*time.Second)),
	)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)
	endpoint, err := c.Endpoint(ctx, "http")
	require.NoError(t, err)
	client := &http.Client{Timeout: 10 * time.Second}
	for _, path := range []string{"/", "/account"} {
		response, err := client.Get(endpoint + path)
		require.NoError(t, err)
		body, err := io.ReadAll(response.Body)
		response.Body.Close()
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, response.StatusCode)
		require.Contains(t, string(body), "<html")
	}
	// Readiness verifies that the packaged migrations reached the expected schema.
	response, err := client.Get(endpoint + "/ready")
	require.NoError(t, err)
	response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)
}

func TestConcurrentStartupMigrations(t *testing.T) {
	image := testhelpers.GetTestImage("ghcr.io/christfriedbalizou/translator:rolling")
	testCommandSucceeds(t, context.Background(), image, &testhelpers.ContainerConfig{Env: map[string]string{
		"DATABASE_URL":                     "sqlite+aiosqlite:////tmp/concurrent.sqlite3",
		"OIDC_ISSUER":                      "https://issuer.invalid",
		"OIDC_CLIENT_ID":                   "container-smoke-test",
		"OIDC_CLIENT_SECRET_FILE":          "/tmp/oidc-secret",
		"OIDC_REDIRECT_URI":                "https://translator.invalid/api/v1/auth/callback",
		"SESSION_IDLE_TIMEOUT_SECONDS":     "900",
		"SESSION_ABSOLUTE_TIMEOUT_SECONDS": "3600",
	}}, "sh", "-ec", `
printf smoke-test > /tmp/oidc-secret
entrypoint.sh migrate & first=$!
entrypoint.sh migrate & second=$!
wait "$first"
wait "$second"
python -c 'import sqlite3; from alembic.config import Config; from alembic.script import ScriptDirectory; assert sqlite3.connect("/tmp/concurrent.sqlite3").execute("select version_num from alembic_version").fetchone()[0] == ScriptDirectory.from_config(Config("alembic.ini")).get_current_head()'
`)
}

func TestMigrationFailureStopsStartup(t *testing.T) {
	image := testhelpers.GetTestImage("ghcr.io/christfriedbalizou/translator:rolling")
	testCommandSucceeds(t, context.Background(), image, &testhelpers.ContainerConfig{Env: map[string]string{
		"DATABASE_URL": "sqlite+aiosqlite:////does-not-exist/database.sqlite3",
	}}, "sh", "-ec", `
/usr/local/bin/python /usr/local/lib/translator-decrypt.py

if entrypoint.sh api; then exit 1; fi
if entrypoint.sh worker; then exit 1; fi
`)
}

func translatorSecrets(t *testing.T) testcontainers.ContainerCustomizer {
	t.Helper()
	key := os.Getenv("TRANSLATOR_TEST_PRIVATE_KEY_FILE")
	require.NotEmpty(t, key, "set TRANSLATOR_TEST_PRIVATE_KEY_FILE to the exported GPG private key")
	files := []testcontainers.ContainerFile{{HostFilePath: key, ContainerFilePath: "/run/secrets/translator-private-key", FileMode: 0444}}
	env := map[string]string{}
	if passphrase := os.Getenv("TRANSLATOR_TEST_PASSPHRASE_FILE"); passphrase != "" {
		files = append(files, testcontainers.ContainerFile{HostFilePath: passphrase, ContainerFilePath: "/run/secrets/translator-passphrase", FileMode: 0444})
		env["TRANSLATOR_GPG_PASSPHRASE_FILE"] = "/run/secrets/translator-passphrase"
	}
	opts := []testcontainers.ContainerCustomizer{
		testcontainers.WithFiles(files...),
		testcontainers.WithEnv(env),
		testcontainers.WithTmpfs(map[string]string{"/opt/translator": "rw,exec,nosuid,nodev,size=4g,uid=1000,gid=1000,mode=0700"}),
	}
	return testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
		for _, opt := range opts {
			if err := opt.Customize(req); err != nil {
				return err
			}
		}
		return nil
	})
}

func testCommandSucceeds(t *testing.T, ctx context.Context, image string, config *testhelpers.ContainerConfig, entrypoint string, args ...string) {
	t.Helper()
	opts := []testcontainers.ContainerCustomizer{translatorSecrets(t), testcontainers.WithEntrypoint(entrypoint), testcontainers.WithCmd(args...), testcontainers.WithWaitStrategy(wait.ForExit().WithExitTimeout(3 * time.Minute))}
	if config != nil {
		opts = append(opts, testcontainers.WithEnv(config.Env))
	}
	c, err := testcontainers.Run(ctx, image, opts...)
	testcontainers.CleanupContainer(t, c)
	require.NoError(t, err)
	state, err := c.State(ctx)
	require.NoError(t, err)
	if state.ExitCode != 0 {
		logs, logErr := c.Logs(ctx)
		if logErr == nil {
			defer logs.Close()
			output, _ := io.ReadAll(logs)
			t.Log(string(output))
		}
	}
	require.Equal(t, 0, state.ExitCode)
}

func TestEncryptedImageWithoutKey(t *testing.T) {
	image := testhelpers.GetTestImage("ghcr.io/christfriedbalizou/translator:rolling")
	testhelpers.TestCommandSucceeds(t, context.Background(), image, nil, "sh", "-ec", `
 test -s /usr/share/translator/payload.tar.gz.gpg
 test ! -e /opt/translator/.venv
 test ! -e /opt/translator/migrations
 if entrypoint.sh true; then exit 1; fi
 `)
}

func TestDecryptionFailures(t *testing.T) {
	image := testhelpers.GetTestImage("ghcr.io/christfriedbalizou/translator:rolling")
	for _, key := range []string{"/run/secrets/missing", "/usr/share/translator/LICENSE"} {
		t.Run(key, func(t *testing.T) {
			testCommandSucceeds(t, context.Background(), image, &testhelpers.ContainerConfig{Env: map[string]string{"TRANSLATOR_GPG_PRIVATE_KEY_FILE": key}}, "sh", "-ec", `
if entrypoint.sh true; then exit 1; fi
test ! -e /opt/translator/.decrypted
test ! -e /opt/translator/.venv
`)
		})
	}
	testCommandSucceeds(t, context.Background(), image, nil, "sh", "-ec", `
mkdir -m 700 /opt/translator/wrong-key-home
gpg --homedir /opt/translator/wrong-key-home --batch --pinentry-mode loopback --passphrase '' --quick-generate-key 'Wrong test key' rsa2048 encr 1d >/dev/null 2>&1
gpg --homedir /opt/translator/wrong-key-home --batch --armor --export-secret-keys > /opt/translator/wrong-key
if TRANSLATOR_GPG_PRIVATE_KEY_FILE=/opt/translator/wrong-key entrypoint.sh true; then exit 1; fi
gpgconf --homedir /opt/translator/wrong-key-home --kill gpg-agent
test ! -e /opt/translator/.decrypted
/usr/local/bin/python - <<'PY'
import importlib.util
from pathlib import Path
import subprocess
spec = importlib.util.spec_from_file_location("decrypt", "/usr/local/lib/translator-decrypt.py")
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)
data = module.PAYLOAD.read_bytes()
module.PAYLOAD = Path("/opt/translator/corrupt.gpg")
module.PAYLOAD.write_bytes(data[:-32])
del data
try:
    module.unlock()
except subprocess.CalledProcessError:
    pass
else:
    raise AssertionError("corrupted payload accepted")
assert not Path("/opt/translator/.decrypted").exists()
assert not Path("/opt/translator/.venv").exists()
assert not list(Path("/opt/translator").glob(".decrypt-*"))
PY
`)
}
