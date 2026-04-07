package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestLogging_JSONFormat(t *testing.T) {
	var buf bytes.Buffer
	log := newLogger(&buf, "test-run-id")

	log.Info("hello", "key", "value")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("log output is not valid JSON: %v\nraw: %s", err, buf.String())
	}

	if msg, _ := m["msg"].(string); msg != "hello" {
		t.Errorf("msg = %q, want %q", msg, "hello")
	}
}

func TestLogging_RunIDField(t *testing.T) {
	var buf bytes.Buffer
	log := newLogger(&buf, "abc-123")

	log.Info("test message")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("log output is not valid JSON: %v\nraw: %s", err, buf.String())
	}

	runID, ok := m["run_id"]
	if !ok {
		t.Fatal("expected run_id field in log output")
	}
	if runID != "abc-123" {
		t.Errorf("run_id = %q, want %q", runID, "abc-123")
	}
}

func TestLogging_CorrelationLogger(t *testing.T) {
	var buf bytes.Buffer
	log := newLogger(&buf, "run-xyz")

	child := log.With("prospect_url", "https://linkedin.com/in/alice")
	child.Info("processing prospect")

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("log output is not valid JSON: %v\nraw: %s", err, buf.String())
	}

	if m["run_id"] != "run-xyz" {
		t.Errorf("run_id = %v, want %q", m["run_id"], "run-xyz")
	}
	if m["prospect_url"] != "https://linkedin.com/in/alice" {
		t.Errorf("prospect_url = %v, want %q", m["prospect_url"], "https://linkedin.com/in/alice")
	}
}
