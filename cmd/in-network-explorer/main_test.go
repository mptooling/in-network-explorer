package main_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binaryPath holds the path to the compiled test binary.
var binaryPath string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "explorer-test-*")
	if err != nil {
		panic(err)
	}

	binaryPath = filepath.Join(dir, "explorer")
	build := exec.CommandContext(context.Background(), "go", "build", "-o", binaryPath, ".")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		_ = os.RemoveAll(dir)
		panic("failed to build binary: " + err.Error())
	}

	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}

func TestCLI_NoArgs(t *testing.T) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, binaryPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit code")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("unexpected error type: %v", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("exit code = %d, want 1", exitErr.ExitCode())
	}
	if !strings.Contains(stderr.String(), "usage:") {
		t.Errorf("stderr = %q, want it to contain \"usage:\"", stderr.String())
	}
}

func TestCLI_UnknownCommand(t *testing.T) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, binaryPath, "foo")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit code")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("unexpected error type: %v", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Errorf("exit code = %d, want 1", exitErr.ExitCode())
	}
	if !strings.Contains(stderr.String(), "unknown command: foo") {
		t.Errorf("stderr = %q, want it to contain \"unknown command: foo\"", stderr.String())
	}
}

func TestCLI_KnownCommands(t *testing.T) {
	// Each command should run without panicking and produce log output.
	// Commands that need unimplemented adapters log an error and exit 0.
	cases := []struct {
		name     string
		wantLike string // substring expected in stdout
	}{
		{"scrape", "SEARCH_QUERY"},
		{"analyze", "analyze failed"},
		{"report", "report"},
		{"calibrate", "calibrate failed"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			// Create empty .env so loadDotEnv does not panic.
			if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(""), 0o644); err != nil {
				t.Fatal(err)
			}

			ctx := context.Background()
			cmd := exec.CommandContext(ctx, binaryPath, tc.name)
			cmd.Dir = dir
			cmd.Env = append(os.Environ(),
				"AWS_REGION=eu-central-1",
				"DYNAMO_TABLE=prospects",
				"DYNAMO_ENDPOINT=http://localhost:1", // unreachable, fails fast
				"LINKEDIN_COOKIES_SECRET=arn:aws:secretsmanager:eu-central-1:123:secret:test",
				"CHROME_PROFILE_DIR=/tmp/chrome-test",
			)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				t.Fatalf("command %q failed: %v\nstdout: %s\nstderr: %s", tc.name, err, stdout.String(), stderr.String())
			}

			if !strings.Contains(stdout.String(), tc.wantLike) {
				t.Errorf("stdout = %q, want it to contain %q", stdout.String(), tc.wantLike)
			}
		})
	}
}
