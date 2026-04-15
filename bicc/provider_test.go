package bicc

import (
	"testing"
)

func TestProvider(t *testing.T) {
	provider := Provider()
	if provider == nil {
		t.Fatal("Provider should not be nil")
	}

	resources := provider.ResourcesMap
	t.Logf("Resources found: %v", len(resources))
	for name := range resources {
		t.Logf("  - %s", name)
	}

	if _, ok := resources["bicc_job"]; !ok {
		t.Error("bicc_job resource not found")
	}

	if _, ok := resources["bicc_job_backfill"]; !ok {
		t.Error("bicc_job_backfill resource not found")
	}
}
