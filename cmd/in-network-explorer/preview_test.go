package main_test

import (
	"context"
	"net/http"
	"os/exec"
	"testing"
	"time"
)

func TestCLI_Preview(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binaryPath, "preview")

	if err := cmd.Start(); err != nil {
		t.Fatalf("start preview: %v", err)
	}

	// Wait for the server to be ready.
	var resp *http.Response
	for range 20 {
		time.Sleep(100 * time.Millisecond)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8085/", nil)
		resp, _ = http.DefaultClient.Do(req)
		if resp != nil {
			break
		}
	}
	if resp == nil {
		t.Fatal("preview server did not start within 2s")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status: got %d, want 200", resp.StatusCode)
	}

	// Cleanup: cancel context kills the process.
	cancel()
	_ = cmd.Wait()
}
