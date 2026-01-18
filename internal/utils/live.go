package utils

import (
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/kashifulhaque/f1-tui/internal/models"
)

type driverProfile struct {
	Name string
	Team string
}

// Snapshot of the 2024 grid for deterministic live timing simulation.
var liveDrivers = []driverProfile{
	{Name: "Max Verstappen", Team: "Red Bull"},
	{Name: "Sergio Perez", Team: "Red Bull"},
	{Name: "Lewis Hamilton", Team: "Mercedes"},
	{Name: "George Russell", Team: "Mercedes"},
	{Name: "Charles Leclerc", Team: "Ferrari"},
	{Name: "Carlos Sainz", Team: "Ferrari"},
	{Name: "Lando Norris", Team: "McLaren"},
	{Name: "Oscar Piastri", Team: "McLaren"},
	{Name: "Fernando Alonso", Team: "Aston Martin"},
	{Name: "Lance Stroll", Team: "Aston Martin"},
	{Name: "Esteban Ocon", Team: "Alpine"},
	{Name: "Pierre Gasly", Team: "Alpine"},
	{Name: "Yuki Tsunoda", Team: "RB"},
	{Name: "Daniel Ricciardo", Team: "RB"},
	{Name: "Valtteri Bottas", Team: "Sauber"},
	{Name: "Zhou Guanyu", Team: "Sauber"},
	{Name: "Kevin Magnussen", Team: "Haas"},
	{Name: "Nico Hulkenberg", Team: "Haas"},
	{Name: "Alexander Albon", Team: "Williams"},
	{Name: "Logan Sargeant", Team: "Williams"},
}

const progressJitter = 0.3

func GenerateLiveTiming(session models.UISession, now time.Time) []models.DriverResult {
	seed := sessionSeed(session)
	step := now.Unix() / 5
	rng := rand.New(rand.NewSource(seed + step))

	drivers := make([]driverProfile, len(liveDrivers))
	copy(drivers, liveDrivers)
	rng.Shuffle(len(drivers), func(i, j int) { drivers[i], drivers[j] = drivers[j], drivers[i] })

	sessionProgress := sessionCompletion(session, now)
	totalLaps := sessionLapTarget(session.Kind)
	lapCount := int(math.Max(1, sessionProgress*float64(totalLaps)))

	results := make([]models.DriverResult, 0, len(drivers))
	gapSeconds := 0.0
	for i, d := range drivers {
		if i > 0 {
			gapSeconds += 0.25 + rng.Float64()*1.4
		}
		progress := math.Mod(sessionProgress+rng.Float64()*progressJitter, 1.0)
		speed := 295 + rng.Float64()*38
		gap := "Leader"
		if i > 0 {
			gap = fmt.Sprintf("+%.3fs", gapSeconds)
		}
		results = append(results, models.DriverResult{
			Position:    fmt.Sprintf("%d", i+1),
			Driver:      d.Name,
			Constructor: d.Team,
			Gap:         gap,
			Laps:        fmt.Sprintf("%d/%d", lapCount, totalLaps),
			Speed:       fmt.Sprintf("%.0f km/h", speed),
			Progress:    TrackProgressBar(progress, 14),
		})
	}
	return results
}

func TrackProgressBar(progress float64, width int) string {
	if width < 3 {
		width = 3
	}
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	pos := int(math.Round(progress * float64(width-1)))
	bar := make([]rune, width)
	for i := range bar {
		bar[i] = '─'
	}
	bar[pos] = '●'
	return "[" + string(bar) + "]"
}

func IsLiveSession(session models.UISession, now time.Time) bool {
	if session.Kind == "" {
		return false
	}
	if !strings.HasPrefix(session.Kind, "Practice") &&
		session.Kind != "Qualifying" &&
		session.Kind != "Race" &&
		session.Kind != "Sprint" {
		return false
	}
	return now.After(session.Start) && now.Before(session.End)
}

func sessionSeed(session models.UISession) int64 {
	h := fnv.New64a()
	h.Write([]byte(session.Kind))
	h.Write([]byte(session.Start.UTC().Format(time.RFC3339)))
	h.Write([]byte(session.End.UTC().Format(time.RFC3339)))
	return int64(h.Sum64())
}

func sessionCompletion(session models.UISession, now time.Time) float64 {
	if now.Before(session.Start) {
		return 0
	}
	total := session.End.Sub(session.Start)
	if total <= 0 {
		return 0
	}
	elapsed := now.Sub(session.Start)
	if elapsed > total {
		elapsed = total
	}
	return elapsed.Seconds() / total.Seconds()
}

func sessionLapTarget(kind string) int {
	switch {
	case kind == "Race":
		return 60
	case kind == "Sprint":
		return 20
	case strings.HasPrefix(kind, "Practice"):
		return 28
	default:
		return 15
	}
}
