package utils

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kashifulhaque/f1-tui/internal/models"
)

func TestTrackProgressBar(t *testing.T) {
	width := 9
	bar := TrackProgressBar(0.5, width)
	if bar == "" {
		t.Fatal("expected progress bar")
	}
	expectedLen := width + 2
	if len([]rune(bar)) != expectedLen {
		t.Fatalf("expected bar width %d, got %d", expectedLen, len([]rune(bar)))
	}
	if !strings.Contains(bar, "●") {
		t.Fatalf("expected marker in progress bar, got %q", bar)
	}
}

func TestGenerateLiveTimingDeterministic(t *testing.T) {
	session := models.UISession{
		Kind:  "Race",
		Start: time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC),
		End:   time.Date(2024, 3, 1, 14, 0, 0, 0, time.UTC),
	}
	now := session.Start.Add(30 * time.Minute)
	first := GenerateLiveTiming(session, now)
	second := GenerateLiveTiming(session, now)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("expected deterministic live timing for same timestamp")
	}
	if len(first) != len(liveDrivers) {
		t.Fatalf("expected %d drivers, got %d", len(liveDrivers), len(first))
	}
	if first[0].Gap != "Leader" {
		t.Fatalf("expected leader gap, got %q", first[0].Gap)
	}
	if first[0].Progress == "" {
		t.Fatal("expected progress bar")
	}
	later := GenerateLiveTiming(session, now.Add(6*time.Second))
	if reflect.DeepEqual(first, later) {
		t.Fatal("expected live timing to change between ticks")
	}
}
