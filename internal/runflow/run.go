package runflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/service-lasso/service-lasso-harness/internal/contract"
	"github.com/service-lasso/service-lasso-harness/internal/result"
)

const (
	resultFileName  = "run-result.json"
	summaryFileName = "summary.json"
)

type ValidationFailure struct {
	Errors []string
}

func (v *ValidationFailure) Error() string {
	if len(v.Errors) == 0 {
		return "contract validation failed"
	}

	return fmt.Sprintf("contract validation failed: %s", v.Errors[0])
}

func Execute(contractPath string, outputDir string, version string) (string, string, error) {
	startedAt := time.Now().UTC()
	cleanContractPath := filepath.Clean(contractPath)
	cleanOutputDir := filepath.Clean(outputDir)

	if err := os.MkdirAll(cleanOutputDir, 0o755); err != nil {
		return "", "", fmt.Errorf("create output directory: %w", err)
	}

	resultPath := filepath.Join(cleanOutputDir, resultFileName)
	summaryPath := filepath.Join(cleanOutputDir, summaryFileName)

	runResult := result.RunResult{
		HarnessVersion: version,
		Status:         "invalid-contract",
		Message:        "contract validation failed",
		ContractPath:   cleanContractPath,
		OutputDir:      cleanOutputDir,
		Stages:         defaultStages(contract.Document{}),
		StartedAt:      startedAt.Format(time.RFC3339),
		FinishedAt:     startedAt.Format(time.RFC3339),
		Notes: []string{
			"Starter repo stub only; lifecycle execution is not implemented yet.",
		},
	}

	runSummary := result.RunSummary{
		Status:       runResult.Status,
		ContractPath: cleanContractPath,
		OutputDir:    cleanOutputDir,
		StageSummary: runResult.Stages,
		Notes:        runResult.Notes,
	}

	doc, loadErr := contract.LoadFile(cleanContractPath)
	if loadErr != nil {
		runResult.ValidationErrors = []string{loadErr.Error()}
		runSummary.ValidationErrors = runResult.ValidationErrors
		if writeErr := writeArtifacts(resultPath, summaryPath, &runResult, &runSummary); writeErr != nil {
			return resultPath, summaryPath, errors.Join(loadErr, writeErr)
		}
		return resultPath, summaryPath, &ValidationFailure{Errors: runResult.ValidationErrors}
	}

	validationErrs := doc.Validate()
	resolvedArtifactPath := contract.ResolveArtifactPath(cleanContractPath, doc)
	artifactExists := fileExists(resolvedArtifactPath)

	runResult.ServiceID = doc.ServiceID
	runResult.HealthType = doc.Health.Type
	runResult.Artifact = result.ArtifactResult{
		Path:         doc.Artifact.Path,
		ResolvedPath: resolvedArtifactPath,
		Kind:         doc.Artifact.Kind,
		Exists:       artifactExists,
	}
	runResult.Stages = defaultStages(*doc)
	runSummary.ServiceID = doc.ServiceID
	runSummary.HealthType = doc.Health.Type
	runSummary.ResolvedArtifact = resolvedArtifactPath
	runSummary.StageSummary = runResult.Stages

	if len(validationErrs) > 0 {
		runResult.ValidationErrors = validationErrs
		runSummary.ValidationErrors = validationErrs
		if writeErr := writeArtifacts(resultPath, summaryPath, &runResult, &runSummary); writeErr != nil {
			return resultPath, summaryPath, errors.Join((&ValidationFailure{Errors: validationErrs}), writeErr)
		}
		return resultPath, summaryPath, &ValidationFailure{Errors: validationErrs}
	}

	runResult.Status = "stubbed"
	runResult.Message = "contract accepted; execution stages were recorded but not run"
	runSummary.Status = runResult.Status
	runSummary.Notes = append(runSummary.Notes, fmt.Sprintf("artifact exists: %t", artifactExists))

	if writeErr := writeArtifacts(resultPath, summaryPath, &runResult, &runSummary); writeErr != nil {
		return resultPath, summaryPath, writeErr
	}

	return resultPath, summaryPath, nil
}

func defaultStages(doc contract.Document) []result.StageResult {
	stageStatus := func(enabled bool) result.StageResult {
		if enabled {
			return result.StageResult{Status: "planned", Detail: "recorded by starter stub"}
		}

		return result.StageResult{Status: "not-requested", Detail: "not enabled in contract"}
	}

	install := stageStatus(doc.Lifecycle.Install)
	install.Name = "install"

	config := stageStatus(doc.Lifecycle.Config)
	config.Name = "config"

	start := stageStatus(doc.Lifecycle.Start)
	start.Name = "start"

	health := stageStatus(doc.Lifecycle.Start)
	health.Name = "health"

	stop := stageStatus(doc.Lifecycle.Stop)
	stop.Name = "stop"

	return []result.StageResult{install, config, start, health, stop}
}

func writeArtifacts(resultPath string, summaryPath string, runResult *result.RunResult, runSummary *result.RunSummary) error {
	finishedAt := time.Now().UTC().Format(time.RFC3339)
	runResult.FinishedAt = finishedAt
	runSummary.Status = runResult.Status

	if err := writeJSON(resultPath, runResult); err != nil {
		return fmt.Errorf("write result artifact: %w", err)
	}

	if err := writeJSON(summaryPath, runSummary); err != nil {
		return fmt.Errorf("write summary artifact: %w", err)
	}

	return nil
}

func writeJSON(path string, value any) error {
	content, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}

	content = append(content, '\n')
	return os.WriteFile(path, content, 0o644)
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}

	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !info.IsDir()
}
