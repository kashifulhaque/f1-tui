package utils

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kashifulhaque/f1-tui/internal/models"
)

// ParseUTC parses Ergast date and time strings into UTC time
func ParseUTC(date, t string) (time.Time, error) {
	if t == "" {
		return time.Time{}, errors.New("empty time")
	}

	stamp := fmt.Sprintf("%sT%s", date, strings.TrimSuffix(t, "Z"))
	layouts := []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02T15:04"}

	var lastErr error
	for _, layout := range layouts {
		if ts, err := time.Parse(layout, stamp); err == nil {
			return ts.UTC(), nil
		} else {
			lastErr = err
		}
	}
	return time.Time{}, lastErr
}

// ToLocal converts UTC time to local time
func ToLocal(t time.Time) time.Time {
	return t.In(time.Local)
}

// ApproxEnd calculates approximate end time based on session type
func ApproxEnd(kind string, start time.Time) time.Time {
	switch kind {
	case "Race":
		return start.Add(2*time.Hour + 5*time.Minute)
	case "Qualifying", "Sprint":
		return start.Add(1 * time.Hour)
	default:
		return start.Add(90 * time.Minute)
	}
}

// BuildUISessions converts Race data to UI-friendly session format
func BuildUISessions(r models.Race, season, round string, resultsURL func(string, string) string) ([]models.UISession, models.UISession, error) {
	var sessions []models.UISession

	addSession := func(kind string, s *models.Session) error {
		if s == nil || s.Time == "" {
			return nil
		}

		utc, err := ParseUTC(s.Date, s.Time)
		if err != nil {
			return err
		}
		start := ToLocal(utc)
		sessions = append(sessions, models.UISession{
			Kind:  kind,
			Start: start,
			End:   ApproxEnd(kind, start),
			URL:   resultsURL(season, round),
		})
		return nil
	}

	var err error
	if err = addSession("Practice 1", r.FirstPractice); err != nil {
		return nil, models.UISession{}, err
	}
	if err = addSession("Practice 2", r.SecondPractice); err != nil {
		return nil, models.UISession{}, err
	}
	if err = addSession("Practice 3", r.ThirdPractice); err != nil {
		return nil, models.UISession{}, err
	}
	if err = addSession("Sprint Shootout", r.SprintShootout); err != nil {
		return nil, models.UISession{}, err
	}
	if err = addSession("Sprint", r.Sprint); err != nil {
		return nil, models.UISession{}, err
	}
	if err = addSession("Qualifying", r.Qualifying); err != nil {
		return nil, models.UISession{}, err
	}

	// Race
	raceUTC, err := ParseUTC(r.Date, r.Time)
	if err != nil {
		return nil, models.UISession{}, err
	}
	raceLocal := ToLocal(raceUTC)
	race := models.UISession{
		Kind:  "Race",
		Start: raceLocal,
		End:   ApproxEnd("Race", raceLocal),
		URL:   resultsURL(season, round),
	}

	return sessions, race, nil
}
