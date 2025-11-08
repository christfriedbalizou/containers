package main

import (
	"context"
	"testing"

	"github.com/christfriedbalizou/containers/testhelpers"
)

func Test(t *testing.T) {
	ctx := context.Background()
	image := testhelpers.GetTestImage("ghcr.io/christfriedbalizou/minio:rolling")
	testhelpers.TestHTTPEndpoint(t, ctx, image, testhelpers.HTTPTestConfig{Port: "9000", Path: "/minio/health/live", StatusCode: 200}, nil)
}