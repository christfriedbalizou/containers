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
		Env: map[string]string{
			// The image deliberately refuses MinIO's default credentials.
			"MINIO_ROOT_USER":     "testadmin",
			"MINIO_ROOT_PASSWORD": "test-password-please-change",
		},
		Tmpfs: []string{"/data"},
	}

	healthCheck := testhelpers.HTTPTestConfig{
		Port:       "9000",
		Path:       "/minio/health/ready",
		StatusCode: 200,
	}

	testhelpers.TestHTTPEndpoint(t, ctx, image, healthCheck, containerConfig)

	consoleCheck := testhelpers.HTTPTestConfig{
		Port:       "9001",
		Path:       "/",
		StatusCode: 200,
	}
	testhelpers.TestHTTPEndpoint(t, ctx, image, consoleCheck, containerConfig)
}
