package contract

import "testing"

func TestApplyDefaultsSetsProcessHealth(t *testing.T) {
	doc := &Document{
		ServiceID: "echo-service",
		Artifact: Artifact{
			Path: "dist/echo.zip",
			Kind: "archive",
		},
		Lifecycle: Lifecycle{
			Install: true,
		},
		Health: Health{
			TimeoutSeconds: 30,
		},
	}

	doc.ApplyDefaults()

	if doc.Health.Type != "process" {
		t.Fatalf("expected default health type process, got %q", doc.Health.Type)
	}
}

func TestValidateRejectsMissingRequiredFields(t *testing.T) {
	doc := &Document{}
	doc.ApplyDefaults()

	errs := doc.Validate()
	if len(errs) < 4 {
		t.Fatalf("expected several validation errors, got %v", errs)
	}
}

func TestValidateRejectsUnsupportedHealthType(t *testing.T) {
	doc := &Document{
		ServiceID: "echo-service",
		Artifact: Artifact{
			Path: "dist/echo.zip",
			Kind: "archive",
		},
		Lifecycle: Lifecycle{
			Install: true,
		},
		Health: Health{
			Type:           "udp",
			TimeoutSeconds: 30,
		},
	}

	errs := doc.Validate()
	if len(errs) != 1 {
		t.Fatalf("expected exactly one validation error, got %v", errs)
	}
}
