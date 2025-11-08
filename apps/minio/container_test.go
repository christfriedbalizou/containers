package main

import (
	"context"
	"testing"

	"github.com/christfriedbalizou/containers/testhelpers"
)

func Test(t *testing.T) {
	ctx := context.Background()
	image := testhelpers.GetTestImage("ghcr.io/christfriedbalizou/minio:rolling")

	containerConfig := &testhelpers.ContainerConfig{
		Tmpfs: []string{"/data"},
	}

	healthCheck := testhelpers.HTTPTestConfig{
		Port:       "9000",
		Path:       "/minio/health/ready",
		StatusCode: 200,
	}

	testhelpers.TestHTTPEndpoint(t, ctx, image, healthCheck, containerConfig)
}
