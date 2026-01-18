package ui

import (
	"time"

	"github.com/kashifulhaque/f1-tui/internal/models"
	"github.com/kashifulhaque/f1-tui/internal/utils"
)

// NewSimulatedModel builds a model with deterministic local data for demo runs.
func NewSimulatedModel(now time.Time) Model {
	m := InitialModel()
	m.loading = false
	m.showCircuit = true

	race := buildSimulatedRace(now)
	m.races = []models.Race{race}
	m.idx = 0
	m.rebuild()

	m.showResults = true
	m.resultsSession = m.race
	m.resultsView = models.ResultsView{
		SessionName: m.race.Kind,
		RaceName:    race.RaceName,
		Loading:     true,
		Live:        utils.IsLiveSession(m.race, now),
	}
	m.resultsTbl = newResultsTable(true)
	m.initCmd = fetchLiveResultsCmd(m.race, race.RaceName)
	return m
}

func buildSimulatedRace(now time.Time) models.Race {
	raceStart := now.UTC().Add(-15 * time.Minute)
	qualifyingStart := now.UTC().Add(-10 * time.Minute)
	practiceStart := now.UTC().Add(-5 * time.Minute)

	race := models.Race{
		Season:        now.Format("2006"),
		Round:         "1",
		RaceName:      "Simulated Grand Prix",
		Date:          raceStart.Format("2006-01-02"),
		Time:          raceStart.Format("15:04:05Z"),
		FirstPractice: sessionAt(practiceStart),
		Qualifying:    sessionAt(qualifyingStart),
	}

	race.Circuit.CircuitName = "Simulation Circuit"
	race.Circuit.URL = "simulated://circuit"
	race.Circuit.Location.Locality = "Copilot City"
	race.Circuit.Location.Country = "Simuland"
	race.Circuit.Location.Lat = "0"
	race.Circuit.Location.Long = "0"

	return race
}

func sessionAt(t time.Time) *models.Session {
	return &models.Session{
		Date: t.Format("2006-01-02"),
		Time: t.Format("15:04:05Z"),
	}
}
