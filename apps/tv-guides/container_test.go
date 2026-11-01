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
	runGuideCommand(t, image, volume, 0, "sh", "-ec", `
test "$(id -u)" = 1000
test ! -d /opt/tv-guides/.git
test ! -e /opt/tv-guides/.env
test -s /usr/share/tv-guides/source-revision
tv-guides init
python - <<'PY'
import importlib.metadata
import os
import sqlite3
from dataclasses import replace
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
settings = replace(load_settings(work), discover_guides=False)
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
	runGuideCommand(t, image, volume, 0, "sh", "-ec", `
test -s /data/state/config.json
test -s /data/state/current/guide.xml
tv-guides integrity
python -c 'import sqlite3; assert sqlite3.connect("/data/state/guide-mappings.sqlite3").execute("SELECT guide_id FROM guide_mapping").fetchone()[0] == "100"'
`)
}

func TestGuideStatePreparation(t *testing.T) {
	image := testhelpers.GetTestImage("ghcr.io/christfriedbalizou/tv-guides:rolling")
	for _, layout := range []string{"fresh", "legacy"} {
		t.Run(layout, func(t *testing.T) {
			volume := fmt.Sprintf("tv-guides-state-%d", time.Now().UnixNano())
			t.Cleanup(func() {
				output, err := exec.Command("docker", "volume", "rm", volume).CombinedOutput()
				require.NoError(t, err, string(output))
			})
			if layout == "legacy" {
				runGuideCommand(t, image, volume, 0, "sh", "-ec", `
TV_GUIDES_WORK_DIR=/data tv-guides init
python - <<'PY'
import sqlite3
from pathlib import Path

with sqlite3.connect("/data/guide-mappings.sqlite3") as database:
    database.execute("CREATE TABLE preparation_fixture (guide_id TEXT)")
    database.execute("INSERT INTO preparation_fixture VALUES ('100')")
for name in ("discovery", "reports", "generations/fixture"):
    directory = Path("/data") / name
    directory.mkdir(mode=0o700, parents=True)
    (directory / "retained").write_text("fixture")
Path("/data/guide-mappings.sqlite3.backup-test").write_text("backup")
Path("/data/current").symlink_to("generations/fixture")
Path("/data/unrelated").write_text("untouched")
PY
chmod 0755 /data
`)
			} else {
				runGuideCommand(t, image, volume, 0, "sh", "-ec", "chmod 0755 /data; printf untouched > /data/unrelated")
			}
			runGuideCommand(t, image, volume, 0, "/usr/local/bin/entrypoint.sh", "init")
			runGuideCommand(t, image, volume, 0, "/usr/local/bin/entrypoint.sh", "integrity")
			runGuideCommand(t, image, volume, 0, "sh", "-ec", `
python - <<'PY'
import stat
from pathlib import Path

assert stat.S_IMODE(Path("/data").stat().st_mode) == 0o755
assert stat.S_IMODE(Path("/data/state").stat().st_mode) == 0o700
assert Path("/data/state/config.json").is_file()
PY
`)
			if layout == "legacy" {
				runGuideCommand(t, image, volume, 0, "sh", "-ec", `
python - <<'PY'
import sqlite3
from pathlib import Path

with sqlite3.connect("/data/state/guide-mappings.sqlite3") as database:
    assert database.execute("SELECT guide_id FROM preparation_fixture").fetchone()[0] == "100"
for name in ("discovery", "reports", "generations/fixture", "current"):
    assert (Path("/data/state") / name / "retained").read_text() == "fixture"
assert Path("/data/state/guide-mappings.sqlite3.backup-test").read_text() == "backup"
assert Path("/data/unrelated").read_text() == "untouched"
assert not Path("/data/config.json").exists()
Path("/data/config.json").write_text("conflicting source")
PY
`)
				runGuideCommand(t, image, volume, 1, "/usr/local/bin/entrypoint.sh", "integrity")
				runGuideCommand(t, image, volume, 0, "sh", "-ec", `
python - <<'PY'
import json
from pathlib import Path

assert Path("/data/config.json").read_text() == "conflicting source"
assert "profiles" in json.loads(Path("/data/state/config.json").read_text())
PY
TV_GUIDES_WORK_DIR=/data/custom /usr/local/bin/entrypoint.sh init
test -s /data/custom/config.json
`)
			}
		})
	}
}

func runGuideCommand(t *testing.T, image, volume string, expectedExit int, entrypoint string, command ...string) {
	t.Helper()
	ctx := context.Background()
	container, err := testcontainers.Run(ctx, image,
		testcontainers.WithMounts(testcontainers.VolumeMount(volume, "/data")),
		testcontainers.WithEnv(map[string]string{"TV_GUIDES_WORK_DIR": "/data/state"}),
		testcontainers.WithEntrypoint(entrypoint),
		testcontainers.WithCmd(command...),
		testcontainers.WithWaitStrategy(wait.ForExit()),
	)
	testcontainers.CleanupContainer(t, container)
	require.NoError(t, err)
	state, err := container.State(ctx)
	require.NoError(t, err)
	if state.ExitCode != expectedExit {
		logs, logErr := container.Logs(ctx)
		if logErr == nil {
			defer logs.Close()
			contents, _ := io.ReadAll(logs)
			t.Logf("container logs: %s", contents)
		}
	}
	require.Equal(t, expectedExit, state.ExitCode)
}
