package utils

import (
	"os"
	"strings"
	"testing"
)

func TestCircuitASCIIIsDeterministic(t *testing.T) {
	first := CircuitASCII("https://example.com/track")
	second := CircuitASCII("https://example.com/track")
	if first == "" {
		t.Fatal("expected ascii circuit")
	}
	if first != second {
		t.Fatalf("expected deterministic output, got %q and %q", first, second)
	}
}

func TestFetchCircuitSVGCreatesFile(t *testing.T) {
	path := FetchCircuitSVG("https://example.com/track")
	if path == "" {
		t.Fatal("expected svg path")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected svg file: %v", err)
	}
	if !strings.Contains(string(content), "<svg") {
		t.Fatalf("expected svg content, got %q", string(content))
	}
}
