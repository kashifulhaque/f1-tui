package ui

import (
	"testing"
	"time"
)

func TestNewSimulatedModel(t *testing.T) {
	now := time.Date(2026, 1, 18, 12, 0, 0, 0, time.Local)
	model := NewSimulatedModel(now)

	if model.loading {
		t.Fatalf("expected loading to be false")
	}
	if len(model.races) != 1 {
		t.Fatalf("expected 1 race, got %d", len(model.races))
	}
	if !model.showResults {
		t.Fatalf("expected results view to be enabled")
	}
	if model.resultsView.RaceName != "Simulated Grand Prix" {
		t.Fatalf("unexpected race name: %s", model.resultsView.RaceName)
	}
	if model.initCmd == nil {
		t.Fatalf("expected init command to be set")
	}
}
