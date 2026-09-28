package main

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/christfriedbalizou/containers/testhelpers"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRuntimeContents(t *testing.T) {
	image := testhelpers.GetTestImage("ghcr.io/christfriedbalizou/translator:rolling")
	testhelpers.TestCommandSucceeds(t, context.Background(), image, nil, "sh", "-ec", `
test "$(id -u)" = 1000
test ! -w /opt/translator
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
		testcontainers.WithCmd("-ec", "printf smoke-test > /tmp/oidc-secret; entrypoint.sh migrate; exec entrypoint.sh api"),
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
