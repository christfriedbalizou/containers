package main

import (
	"context"
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/christfriedbalizou/containers/testhelpers"
)

func Test(t *testing.T) {
	ctx := context.Background()
	image := testhelpers.GetTestImage("ghcr.io/christfriedbalizou/vscode:rolling")

	for _, tc := range []struct {
		name    string
		command string
		args    []string
	}{
		{name: "code-server is available", command: "code-server", args: []string{"--version"}},
		{name: "mise is available", command: "mise", args: []string{"--version"}},
		{name: "codex is available", command: "codex", args: []string{"--version"}},
		{name: "bubblewrap is available", command: "bwrap", args: []string{"--version"}},
		{name: "python3 is available", command: "python3", args: []string{"--version"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testhelpers.TestCommandSucceeds(t, ctx, image, nil, tc.command, tc.args...)
		})
	}

	for _, tc := range []struct {
		name string
		path string
	}{
		{name: "defaults directory exists", path: "/opt/code-server-defaults"},
		{name: "home directory exists", path: "/home/coder"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testhelpers.TestCommandSucceeds(t, ctx, image, nil, "test", "-d", tc.path)
		})
	}

	t.Run("metadata is correct", func(t *testing.T) {
		config := inspectImageConfig(t, image)

		require.Equal(t, "root", config.User)
		require.Equal(t, "/home/coder/src", config.WorkingDir)
		require.Contains(t, config.ExposedPorts, "8080/tcp")
	})
}

func inspectImageConfig(t *testing.T, image string) imageConfig {
	t.Helper()

	output, err := exec.Command("docker", "pull", image).CombinedOutput()
	require.NoError(t, err, string(output))

	output, err = exec.Command("docker", "image", "inspect", image).CombinedOutput()
	require.NoError(t, err, string(output))

	var inspected []struct {
		Config imageConfig `json:"Config"`
	}
	require.NoError(t, json.Unmarshal(output, &inspected))
	require.NotEmpty(t, inspected)

	return inspected[0].Config
}

type imageConfig struct {
	User         string              `json:"User"`
	WorkingDir   string              `json:"WorkingDir"`
	ExposedPorts map[string]struct{} `json:"ExposedPorts"`
}
