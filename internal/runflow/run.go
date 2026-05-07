package runflow

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/service-lasso/service-lasso-harness/internal/contract"
	"github.com/service-lasso/service-lasso-harness/internal/result"
)

type ValidationFailure struct {
	Errors []string
}

func (e *ValidationFailure) Error() string {
	if e == nil || len(e.Errors) == 0 {
		return "contract validation failed"
	}
	return "contract validation failed: " + strings.Join(e.Errors, "; ")
}

type serviceManifest struct {
	ID         string `json:"id"`
	ExecConfig struct {
		ExecCwd    string            `json:"execcwd"`
		Executable string            `json:"executable"`
		Env        map[string]string `json:"env"`
	} `json:"execconfig"`
}

func Execute(contractPath, outputDir, harnessVersion string) (string, string, error) {
	resultPath := ""
	summaryPath := ""
	if strings.TrimSpace(contractPath) == "" {
		return resultPath, summaryPath, errors.New("contract path is required")
	}
	if strings.TrimSpace(outputDir) == "" {
		return resultPath, summaryPath, errors.New("output directory is required")
	}
	resultPath = filepath.Join(outputDir, "run-result.json")
	summaryPath = filepath.Join(outputDir, "summary.json")

	startedAt := time.Now().UTC()
	res := result.RunResult{
		HarnessVersion: harnessVersion,
		Status:         "running",
		Message:        "running harness lifecycle",
		ContractPath:   contractPath,
		OutputDir:      outputDir,
		StartedAt:      startedAt.Format(time.RFC3339),
	}

	finish := func() error {
		res.FinishedAt = time.Now().UTC().Format(time.RFC3339)
		summary := result.RunSummary{
			Status:           res.Status,
			ServiceID:        res.ServiceID,
			ContractPath:     res.ContractPath,
			OutputDir:        res.OutputDir,
			HealthType:       res.HealthType,
			ResolvedArtifact: res.Artifact.ResolvedPath,
			StageSummary:     res.Stages,
			ValidationErrors: res.ValidationErrors,
			Notes:            res.Notes,
		}
		if err := os.MkdirAll(outputDir, 0o755); err != nil {
			return err
		}
		if err := writeJSON(resultPath, res); err != nil {
			return err
		}
		if err := writeJSON(summaryPath, summary); err != nil {
			return err
		}
		return nil
	}

	doc, err := contract.LoadFile(contractPath)
	if err != nil {
		res.Status = "failed"
		res.Message = fmt.Sprintf("failed to load contract: %v", err)
		return resultPath, summaryPath, errors.Join(err, finish())
	}
	res.ServiceID = doc.ServiceID
	res.HealthType = doc.Health.Type
	resolvedArtifact := contract.ResolveArtifactPath(filepath.Clean(contractPath), doc)
	_, statErr := os.Stat(resolvedArtifact)
	res.Artifact = result.ArtifactResult{
		Path:         doc.Artifact.Path,
		ResolvedPath: resolvedArtifact,
		Kind:         doc.Artifact.Kind,
		Exists:       statErr == nil,
	}

	if errs := doc.Validate(); len(errs) > 0 {
		res.Status = "failed"
		res.Message = "contract validation failed"
		res.ValidationErrors = errs
		validationErr := &ValidationFailure{Errors: errs}
		return resultPath, summaryPath, errors.Join(validationErr, finish())
	}

	workspace := filepath.Join(outputDir, "workspace")
	serviceRoot := filepath.Join(workspace, "service")
	if err := os.RemoveAll(workspace); err != nil {
		res.Status = "failed"
		res.Message = fmt.Sprintf("failed to clean workspace: %v", err)
		return resultPath, summaryPath, errors.Join(err, finish())
	}
	if err := os.MkdirAll(serviceRoot, 0o755); err != nil {
		res.Status = "failed"
		res.Message = fmt.Sprintf("failed to create workspace: %v", err)
		return resultPath, summaryPath, errors.Join(err, finish())
	}

	if doc.Lifecycle.Install {
		if err := extractArtifact(res.Artifact.ResolvedPath, serviceRoot); err != nil {
			res.Status = "failed"
			res.Message = fmt.Sprintf("install failed: %v", err)
			res.Stages = append(res.Stages, result.StageResult{Name: "install", Status: "failed", Detail: err.Error()})
			return resultPath, summaryPath, errors.Join(err, finish())
		}
		res.Stages = append(res.Stages, result.StageResult{Name: "install", Status: "passed", Detail: "artifact extracted into workspace"})
	}

	manifestPath := filepath.Join(serviceRoot, "service.json")
	manifest, err := loadManifest(manifestPath)
	if err != nil {
		res.Status = "failed"
		res.Message = fmt.Sprintf("failed to load service manifest: %v", err)
		res.Stages = append(res.Stages, result.StageResult{Name: "config", Status: "failed", Detail: err.Error()})
		return resultPath, summaryPath, errors.Join(err, finish())
	}

	if doc.Lifecycle.Config {
		res.Stages = append(res.Stages, result.StageResult{Name: "config", Status: "passed", Detail: "service manifest loaded and runtime env prepared"})
	}

	var runCmd *exec.Cmd
	var combinedOutput []byte
	if doc.Lifecycle.Start {
		runCmd, err = buildCommand(serviceRoot, manifest)
		if err != nil {
			res.Status = "failed"
			res.Message = fmt.Sprintf("start failed: %v", err)
			res.Stages = append(res.Stages, result.StageResult{Name: "start", Status: "failed", Detail: err.Error()})
			return resultPath, summaryPath, errors.Join(err, finish())
		}
		runCmd.Env = mergedEnv(os.Environ(), manifest.ExecConfig.Env)
		combinedOutput, err = runCmd.CombinedOutput()
		logPath := filepath.Join(outputDir, "process-output.log")
		_ = os.WriteFile(logPath, combinedOutput, 0o644)
		if err != nil {
			res.Status = "failed"
			res.Message = fmt.Sprintf("process exited with error: %v", err)
			res.Stages = append(res.Stages, result.StageResult{Name: "start", Status: "failed", Detail: fmt.Sprintf("process failed, see %s", logPath)})
			res.Notes = append(res.Notes, string(combinedOutput))
			return resultPath, summaryPath, errors.Join(err, finish())
		}
		res.Stages = append(res.Stages, result.StageResult{Name: "start", Status: "passed", Detail: fmt.Sprintf("process completed successfully, see %s", logPath)})
	}

	if doc.Health.Type != "" {
		detail := "health contract accepted"
		if doc.Health.Type == "process" && runCmd != nil && runCmd.ProcessState != nil && runCmd.ProcessState.Success() {
			detail = "process executed successfully"
		}
		res.Stages = append(res.Stages, result.StageResult{Name: "health", Status: "passed", Detail: detail})
	}

	if doc.Lifecycle.Stop {
		res.Stages = append(res.Stages, result.StageResult{Name: "stop", Status: "passed", Detail: "no running process remained after lifecycle execution"})
	}

	res.Status = "passed"
	res.Message = "contract validated and lifecycle executed"
	if len(combinedOutput) > 0 {
		res.Notes = append(res.Notes, strings.TrimSpace(string(combinedOutput)))
	}
	return resultPath, summaryPath, finish()
}

func loadManifest(path string) (serviceManifest, error) {
	var m serviceManifest
	data, err := os.ReadFile(path)
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return m, err
	}
	if strings.TrimSpace(m.ExecConfig.Executable) == "" {
		return m, errors.New("service manifest missing execconfig.executable")
	}
	return m, nil
}

func buildCommand(serviceRoot string, manifest serviceManifest) (*exec.Cmd, error) {
	workingDir := filepath.Join(serviceRoot, manifest.ExecConfig.ExecCwd)
	execPath := filepath.Join(workingDir, manifest.ExecConfig.Executable)
	if _, err := os.Stat(execPath); err != nil {
		return nil, fmt.Errorf("executable not found: %s", execPath)
	}
	var cmd *exec.Cmd
	switch strings.ToLower(filepath.Ext(execPath)) {
	case ".ps1":
		cmd = exec.Command("pwsh", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", execPath)
	case ".sh":
		cmd = exec.Command("bash", execPath)
	default:
		cmd = exec.Command(execPath)
	}
	cmd.Dir = workingDir
	return cmd, nil
}

func mergedEnv(base []string, extra map[string]string) []string {
	m := map[string]string{}
	for _, entry := range base {
		parts := strings.SplitN(entry, "=", 2)
		key := parts[0]
		val := ""
		if len(parts) == 2 {
			val = parts[1]
		}
		m[key] = val
	}
	for k, v := range extra {
		m[k] = v
	}
	out := make([]string, 0, len(m))
	for k, v := range m {
		out = append(out, k+"="+v)
	}
	return out
}

func extractArtifact(artifactPath, dest string) error {
	lower := strings.ToLower(artifactPath)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return extractZip(artifactPath, dest)
	case strings.HasSuffix(lower, ".tar.gz") || strings.HasSuffix(lower, ".tgz"):
		return extractTarGz(artifactPath, dest)
	default:
		return fmt.Errorf("unsupported artifact type: %s", artifactPath)
	}
}

func extractZip(zipPath, dest string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	for _, f := range r.File {
		target, err := safeJoin(dest, f.Name)
		if err != nil {
			return fmt.Errorf("zip entry escapes destination: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		src, err := f.Open()
		if err != nil {
			return err
		}
		dst, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, f.Mode())
		if err != nil {
			src.Close()
			return err
		}
		_, copyErr := io.Copy(dst, src)
		src.Close()
		dst.Close()
		if copyErr != nil {
			return copyErr
		}
	}
	return nil
}

func extractTarGz(tarPath, dest string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		target, err := safeJoin(dest, hdr.Name)
		if err != nil {
			return fmt.Errorf("tar entry escapes destination: %s", hdr.Name)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, tr); err != nil {
				out.Close()
				return err
			}
			out.Close()
		}
	}
}

func safeJoin(root string, name string) (string, error) {
	target := filepath.Join(root, filepath.FromSlash(name))
	rel, err := filepath.Rel(root, target)
	if err != nil || rel == "." || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		return "", fmt.Errorf("path escapes root")
	}
	return target, nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
