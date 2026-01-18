package ui

import (
	"time"

	"github.com/kashifulhaque/f1-tui/internal/models"
	"github.com/kashifulhaque/f1-tui/internal/utils"
)

const (
	simulatedRaceIndex       = 0
	simulatedRaceName        = "Simulated Grand Prix"
	simulatedCircuitName     = "Simulation Circuit"
	simulatedCircuitURL      = "simulated://circuit"
	simulatedCircuitLocality = "Copilot City"
	simulatedCircuitCountry  = "Simuland"
	simulatedCircuitLat      = "0"
	simulatedCircuitLong     = "0"
	raceStartOffset          = -15 * time.Minute
	qualifyingStartOffset    = -10 * time.Minute
	practiceStartOffset      = -5 * time.Minute
)

// NewSimulatedModel builds a model with deterministic local data for demo runs.
func NewSimulatedModel(now time.Time) Model {
	m := InitialModel()
	m.loading = false
	m.showCircuit = true

	race := buildSimulatedRace(now)
	m.races = []models.Race{race}
	m.idx = simulatedRaceIndex
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
	raceStart := now.UTC().Add(raceStartOffset)
	qualifyingStart := now.UTC().Add(qualifyingStartOffset)
	practiceStart := now.UTC().Add(practiceStartOffset)

	race := models.Race{
		Season:        now.Format("2006"),
		Round:         "1",
		RaceName:      simulatedRaceName,
		Date:          raceStart.Format("2006-01-02"),
		Time:          raceStart.Format("15:04:05Z"),
		FirstPractice: sessionAt(practiceStart),
		Qualifying:    sessionAt(qualifyingStart),
	}

	race.Circuit.CircuitName = simulatedCircuitName
	race.Circuit.URL = simulatedCircuitURL
	race.Circuit.Location.Locality = simulatedCircuitLocality
	race.Circuit.Location.Country = simulatedCircuitCountry
	race.Circuit.Location.Lat = simulatedCircuitLat
	race.Circuit.Location.Long = simulatedCircuitLong

	return race
}

func sessionAt(t time.Time) *models.Session {
	return &models.Session{
		Date: t.Format("2006-01-02"),
		Time: t.Format("15:04:05Z"),
	}
}
