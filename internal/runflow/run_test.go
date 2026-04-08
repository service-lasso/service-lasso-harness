package runflow

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestExecuteWritesStubArtifactsForValidContract(t *testing.T) {
	tempDir := t.TempDir()
	contractPath := filepath.Join(tempDir, "service-harness.json")
	outputDir := filepath.Join(tempDir, "output")

	contractJSON := `{
  "serviceId": "echo-service",
  "artifact": {
    "path": "dist/echo-service-win32.zip",
    "kind": "archive"
  },
  "dependencies": [],
  "lifecycle": {
    "install": true,
    "config": true,
    "start": true,
    "stop": true
  },
  "health": {
    "timeoutSeconds": 30
  },
  "expect": {
    "logs": true,
    "state": true,
    "exitClean": true
  },
  "artifacts": {
    "captureLogs": true,
    "captureState": true,
    "captureSummary": true
  }
}`

	if err := os.WriteFile(contractPath, []byte(contractJSON), 0o644); err != nil {
		t.Fatalf("write contract: %v", err)
	}

	resultPath, summaryPath, err := Execute(contractPath, outputDir, "test")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if _, err := os.Stat(resultPath); err != nil {
		t.Fatalf("expected result file, got %v", err)
	}

	if _, err := os.Stat(summaryPath); err != nil {
		t.Fatalf("expected summary file, got %v", err)
	}
}

func TestExecuteReturnsValidationFailureForInvalidContract(t *testing.T) {
	tempDir := t.TempDir()
	contractPath := filepath.Join(tempDir, "service-harness.json")
	outputDir := filepath.Join(tempDir, "output")

	if err := os.WriteFile(contractPath, []byte(`{"artifact":{"kind":"archive"}}`), 0o644); err != nil {
		t.Fatalf("write contract: %v", err)
	}

	_, _, err := Execute(contractPath, outputDir, "test")
	if err == nil {
		t.Fatal("expected validation failure")
	}

	var validationErr *ValidationFailure
	if !errors.As(err, &validationErr) {
		t.Fatalf("expected ValidationFailure, got %T", err)
	}
}
