package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestMinIOContainer(t *testing.T) {
	ctx := context.Background()

	// Create a testcontainer request for MinIO
	req := testcontainers.ContainerRequest{
		Image:        "minio:RELEASE.2025-10-15T17-29-55Z",
		ExposedPorts: []string{"9000/tcp", "9001/tcp"},
		Env: map[string]string{
			"MINIO_ROOT_USER":     "testuser",
			"MINIO_ROOT_PASSWORD": "testpass123",
		},
		Tmpfs: map[string]string{
			"/data": "rw,noexec,nosuid,size=1g",
		},
		WaitingFor: wait.ForHTTP("/minio/health/live").WithPort("9000/tcp").WithStartupTimeout(60 * time.Second),
	}

	// Start the container
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Failed to start MinIO container: %v", err)
	}
	defer container.Terminate(ctx)

	// Get the mapped port
	mappedPort, err := container.MappedPort(ctx, "9000")
	if err != nil {
		t.Fatalf("Failed to get mapped port: %v", err)
	}

	// Get the host
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	// Test the health endpoint
	healthURL := fmt.Sprintf("http://%s:%s/minio/health/live", host, mappedPort.Port())
	resp, err := http.Get(healthURL)
	if err != nil {
		t.Fatalf("Failed to call health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Health endpoint returned non-200 status: %d", resp.StatusCode)
	}

	t.Logf("MinIO container started successfully with tmpfs mount")
	t.Logf("Health endpoint accessible at: %s", healthURL)

	// Test the API endpoint is accessible
	apiURL := fmt.Sprintf("http://%s:%s/", host, mappedPort.Port())
	resp2, err := http.Get(apiURL)
	if err != nil {
		t.Fatalf("Failed to call API endpoint: %v", err)
	}
	defer resp2.Body.Close()

	// MinIO returns specific XML response for root path, we just check it's reachable
	if resp2.StatusCode < 200 || resp2.StatusCode >= 500 {
		t.Fatalf("API endpoint returned unexpected status: %d", resp2.StatusCode)
	}

	t.Logf("MinIO API endpoint accessible at: %s", apiURL)
}
