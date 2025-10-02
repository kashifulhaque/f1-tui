package models

import "time"

// Ergast API response structure
type ErgastMRData struct {
	RaceTable struct {
		Season string `json:"season"`
		Races  []Race `json:"Races"`
	} `json:"RaceTable"`
}

type Race struct {
	Season         string   `json:"season"`
	Round          string   `json:"round"`
	RaceName       string   `json:"raceName"`
	Circuit        Circuit  `json:"Circuit"`
	Date           string   `json:"date"`
	Time           string   `json:"time"`
	FirstPractice  *Session `json:"FirstPractice,omitempty"`
	SecondPractice *Session `json:"SecondPractice,omitempty"`
	ThirdPractice  *Session `json:"ThirdPractice,omitempty"`
	Sprint         *Session `json:"Sprint,omitempty"`
	SprintShootout *Session `json:"SprintShootout,omitempty"`
	Qualifying     *Session `json:"Qualifying,omitempty"`
}

type Circuit struct {
	CircuitName string `json:"circuitName"`
	Location    struct {
		Locality string `json:"locality"`
		Country  string `json:"country"`
		Lat      string `json:"lat"`
		Long     string `json:"long"`
	} `json:"Location"`
	URL string `json:"url"`
}

type Session struct {
	Date string `json:"date"`
	Time string `json:"time"`
}

// UI-specific session type
type UISession struct {
	Kind  string
	Start time.Time
	End   time.Time
	URL   string
}

type SessionRow struct {
	Title string
	Time  string
}

type ResultsView struct {
	SessionName string
	RaceName    string
	Results     []DriverResult
	Loading     bool
	Error       error
}

type DriverResult struct {
	Position     string
	Driver       string
	Constructor  string
	Time         string
	Status       string
	Points       string
}
