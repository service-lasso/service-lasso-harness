package result

type RunResult struct {
	HarnessVersion   string         `json:"harnessVersion"`
	Status           string         `json:"status"`
	Message          string         `json:"message"`
	ContractPath     string         `json:"contractPath"`
	OutputDir        string         `json:"outputDir"`
	ServiceID        string         `json:"serviceId,omitempty"`
	HealthType       string         `json:"healthType,omitempty"`
	Artifact         ArtifactResult `json:"artifact"`
	Stages           []StageResult  `json:"stages"`
	ValidationErrors []string       `json:"validationErrors,omitempty"`
	Notes            []string       `json:"notes,omitempty"`
	StartedAt        string         `json:"startedAt"`
	FinishedAt       string         `json:"finishedAt"`
}

type ArtifactResult struct {
	Path         string `json:"path,omitempty"`
	ResolvedPath string `json:"resolvedPath,omitempty"`
	Kind         string `json:"kind,omitempty"`
	Exists       bool   `json:"exists"`
}

type StageResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type RunSummary struct {
	Status           string        `json:"status"`
	ServiceID        string        `json:"serviceId,omitempty"`
	ContractPath     string        `json:"contractPath"`
	OutputDir        string        `json:"outputDir"`
	HealthType       string        `json:"healthType,omitempty"`
	ResolvedArtifact string        `json:"resolvedArtifact,omitempty"`
	StageSummary     []StageResult `json:"stageSummary"`
	ValidationErrors []string      `json:"validationErrors,omitempty"`
	Notes            []string      `json:"notes,omitempty"`
}
