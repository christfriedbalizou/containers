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
	image := testhelpers.GetTestImage("ghcr.io/christfriedbalizou/chromium:rolling")

	for _, tc := range []struct {
		name    string
		command string
		args    []string
	}{
		{name: "chromium is available", command: "chromium", args: []string{"--version"}},
		{name: "fontconfig is available", command: "fc-match", args: []string{"Cascadia Mono NF"}},
		{name: "cascadia nerd font is available", command: "test", args: []string{"-f", "/usr/share/fonts/truetype/cascadia-code/CascadiaMonoNF.ttf"}},
		{name: "powerline font is available", command: "test", args: []string{"-d", "/usr/share/fonts/opentype/powerline"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testhelpers.TestCommandSucceeds(t, ctx, image, nil, tc.command, tc.args...)
		})
	}

	t.Run("metadata is correct", func(t *testing.T) {
		config := inspectImageConfig(t, image)

		require.Contains(t, config.ExposedPorts, "3000/tcp")
		require.Contains(t, config.ExposedPorts, "3001/tcp")
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
	ExposedPorts map[string]struct{} `json:"ExposedPorts"`
}
