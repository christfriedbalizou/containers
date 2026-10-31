package main

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"testing"
	"time"

	"github.com/christfriedbalizou/containers/testhelpers"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestGuidePipelineAndWorkingDirectory(t *testing.T) {
	image := testhelpers.GetTestImage("ghcr.io/christfriedbalizou/tv-guides:rolling")
	volume := fmt.Sprintf("tv-guides-test-%d", time.Now().UnixNano())
	t.Cleanup(func() {
		output, err := exec.Command("docker", "volume", "rm", volume).CombinedOutput()
		require.NoError(t, err, string(output))
	})
	runGuideCommand(t, image, volume, `
test "$(id -u)" = 1000
test ! -d /opt/tv-guides/.git
test ! -e /opt/tv-guides/.env
test -s /usr/share/tv-guides/source-revision
tv-guides init
python - <<'PY'
import importlib.metadata
import os
import sqlite3
from datetime import datetime, timedelta, timezone
from pathlib import Path
from unittest.mock import patch
import xml.etree.ElementTree as ET
import tv_guides
from tv_guides.config import load_settings
from tv_guides.pipeline import refresh
from tv_guides.sources import GuideChannel, Programme

assert importlib.metadata.version("tv-guides") == Path("/usr/share/tv-guides/source-tag").read_text().strip().removeprefix("v")
assert Path(tv_guides.__file__).suffix == ".py"
work = Path(os.environ["TV_GUIDES_WORK_DIR"])
settings = load_settings(work)
now = datetime.now(timezone.utc)
guide = GuideChannel("fixture", "100", "BBC Scotland", "United Kingdom", "Central Scotland", frozenset({"BBCSCOTLAND"}), programmes=(Programme(now, now + timedelta(hours=1), "Fixture"),))
playlist = '#EXTM3U\n#EXTINF:-1 group-title="UK VIP NEW",UK: BBC Scotland\nhttp://example.invalid/fixture/1\n#EXTINF:-1 group-title="UK VIP NEW",UK: BBC Scotland\nhttp://example.invalid/fixture/2\n'
with patch("tv_guides.pipeline.fetch_playlist", return_value=playlist), patch("tv_guides.pipeline.fetch_profiles", return_value=[guide]):
    assert refresh(settings, "http://example.invalid/playlist") == 1
    assert refresh(settings, "http://example.invalid/playlist") == 1
assert ET.parse(work / "current/guide.xml").getroot().find("channel").attrib["id"] == "100"
assert (work / "current/playlist.m3u").read_text().count("#EXTINF") == 1
with sqlite3.connect(settings.database) as database:
    assert database.execute("SELECT duplicate_count FROM channel").fetchone()[0] == 2
    assert database.execute("SELECT guide_id FROM guide_mapping").fetchone()[0] == "100"
    assert database.execute("PRAGMA integrity_check").fetchone()[0] == "ok"
PY
tv-guides summary
tv-guides backup --output /data/state/backup.sqlite3
test -s /data/state/backup.sqlite3
if tv-guides run; then exit 1; fi
`)
	runGuideCommand(t, image, volume, `
test -s /data/state/config.json
test -s /data/state/current/guide.xml
tv-guides integrity
python -c 'import sqlite3; assert sqlite3.connect("/data/state/guide-mappings.sqlite3").execute("SELECT guide_id FROM guide_mapping").fetchone()[0] == "100"'
`)
}

func runGuideCommand(t *testing.T, image, volume, command string) {
	t.Helper()
	ctx := context.Background()
	container, err := testcontainers.Run(ctx, image,
		testcontainers.WithMounts(testcontainers.VolumeMount(volume, "/data")),
		testcontainers.WithEnv(map[string]string{"TV_GUIDES_WORK_DIR": "/data/state"}),
		testcontainers.WithEntrypoint("sh"),
		testcontainers.WithCmd("-ec", command),
		testcontainers.WithWaitStrategy(wait.ForExit()),
	)
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)
	state, err := container.State(ctx)
	require.NoError(t, err)
	if state.ExitCode != 0 {
		logs, logErr := container.Logs(ctx)
		if logErr == nil {
			defer logs.Close()
			contents, _ := io.ReadAll(logs)
			t.Logf("container logs: %s", contents)
		}
	}
	require.Zero(t, state.ExitCode)
}
