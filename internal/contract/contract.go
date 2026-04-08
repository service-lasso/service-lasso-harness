package contract

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var allowedHealthTypes = map[string]struct{}{
	"process":  {},
	"http":     {},
	"tcp":      {},
	"file":     {},
	"variable": {},
}

type Document struct {
	ServiceID    string       `json:"serviceId"`
	Artifact     Artifact     `json:"artifact"`
	Dependencies []Dependency `json:"dependencies"`
	Lifecycle    Lifecycle    `json:"lifecycle"`
	Health       Health       `json:"health"`
	Expect       Expect       `json:"expect"`
	Artifacts    Artifacts    `json:"artifacts"`
}

type Artifact struct {
	Path string `json:"path"`
	Kind string `json:"kind"`
}

type Dependency struct {
	ServiceID string `json:"serviceId,omitempty"`
}

type Lifecycle struct {
	Install bool `json:"install"`
	Config  bool `json:"config"`
	Start   bool `json:"start"`
	Stop    bool `json:"stop"`
}

type Health struct {
	Type           string `json:"type"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
}

type Expect struct {
	Logs      bool `json:"logs"`
	State     bool `json:"state"`
	ExitClean bool `json:"exitClean"`
}

type Artifacts struct {
	CaptureLogs    bool `json:"captureLogs"`
	CaptureState   bool `json:"captureState"`
	CaptureSummary bool `json:"captureSummary"`
}

func LoadFile(path string) (*Document, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read contract: %w", err)
	}

	var doc Document
	decoder := json.NewDecoder(strings.NewReader(string(content)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode contract JSON: %w", err)
	}

	doc.ApplyDefaults()
	return &doc, nil
}

func (d *Document) ApplyDefaults() {
	if strings.TrimSpace(d.Health.Type) == "" {
		d.Health.Type = "process"
	}

	if d.Dependencies == nil {
		d.Dependencies = []Dependency{}
	}
}

func (d *Document) Validate() []string {
	var errs []string

	if strings.TrimSpace(d.ServiceID) == "" {
		errs = append(errs, "serviceId is required")
	}

	if strings.TrimSpace(d.Artifact.Path) == "" {
		errs = append(errs, "artifact.path is required")
	}

	if strings.TrimSpace(d.Artifact.Kind) == "" {
		errs = append(errs, "artifact.kind is required")
	}

	if strings.TrimSpace(d.Health.Type) == "" {
		errs = append(errs, "health.type is required")
	} else if _, ok := allowedHealthTypes[d.Health.Type]; !ok {
		errs = append(errs, fmt.Sprintf("health.type must be one of process, http, tcp, file, variable; got %q", d.Health.Type))
	}

	if d.Health.TimeoutSeconds <= 0 {
		errs = append(errs, "health.timeoutSeconds must be greater than zero")
	}

	if !d.Lifecycle.Install && !d.Lifecycle.Config && !d.Lifecycle.Start && !d.Lifecycle.Stop {
		errs = append(errs, "at least one lifecycle stage must be enabled")
	}

	return errs
}

func ResolveArtifactPath(contractPath string, doc *Document) string {
	contractDir := filepath.Dir(contractPath)
	return filepath.Clean(filepath.Join(contractDir, doc.Artifact.Path))
}
