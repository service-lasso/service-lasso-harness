package runflow

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestExecuteWritesArtifactsForValidContract(t *testing.T) {
	tempDir := t.TempDir()
	contractPath := filepath.Join(tempDir, "service-harness.json")
	outputDir := filepath.Join(tempDir, "output")
	artifactPath := filepath.Join(tempDir, "dist", "echo-service-win32.zip")
	writeFixtureArtifact(t, artifactPath)

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

	var result struct {
		Status string `json:"status"`
		Stages []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"stages"`
	}
	data, err := os.ReadFile(resultPath)
	if err != nil {
		t.Fatalf("read result: %v", err)
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	if result.Status != "passed" {
		t.Fatalf("result status = %q", result.Status)
	}
	assertStage(t, result.Stages, "install", "passed")
	assertStage(t, result.Stages, "config", "passed")
	assertStage(t, result.Stages, "start", "passed")
	assertStage(t, result.Stages, "health", "passed")
	assertStage(t, result.Stages, "stop", "passed")
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

func writeFixtureArtifact(t *testing.T, artifactPath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(artifactPath), 0o755); err != nil {
		t.Fatalf("mkdir artifact dir: %v", err)
	}
	file, err := os.Create(artifactPath)
	if err != nil {
		t.Fatalf("create artifact: %v", err)
	}
	defer file.Close()
	archive := zip.NewWriter(file)
	defer archive.Close()

	scriptName := "run.sh"
	scriptBody := "#!/bin/sh\necho harness-fixture\n"
	if runtime.GOOS == "windows" {
		scriptName = "run.ps1"
		scriptBody = "Write-Output 'harness-fixture'\n"
	}

	manifest := `{
  "id": "echo-service",
  "execconfig": {
    "execcwd": ".",
    "executable": "` + scriptName + `",
    "env": {}
  }
}`
	addZipFile(t, archive, "service.json", manifest)
	addZipFile(t, archive, scriptName, scriptBody)
}

func addZipFile(t *testing.T, archive *zip.Writer, name string, body string) {
	t.Helper()
	writer, err := archive.Create(name)
	if err != nil {
		t.Fatalf("create zip entry: %v", err)
	}
	if _, err := writer.Write([]byte(body)); err != nil {
		t.Fatalf("write zip entry: %v", err)
	}
}

func assertStage(t *testing.T, stages []struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}, name string, status string) {
	t.Helper()
	for _, stage := range stages {
		if stage.Name == name {
			if stage.Status != status {
				t.Fatalf("stage %s status = %s", name, stage.Status)
			}
			return
		}
	}
	t.Fatalf("stage %s missing from %#v", name, stages)
}
