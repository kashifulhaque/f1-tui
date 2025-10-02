package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kashifulhaque/f1-tui/internal/models"
)

const ergastBase = "http://api.jolpi.ca/ergast/f1"

// FetchCurrentSchedule retrieves the current F1 season schedule
func FetchCurrentSchedule(ctx context.Context) ([]models.Race, error) {
	url := fmt.Sprintf("%s/current.json", ergastBase)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ergast schedule error: %s", string(b))
	}

	var outer struct {
		MRData models.ErgastMRData `json:"MRData"`
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&outer); err != nil {
		return nil, err
	}

	return outer.MRData.RaceTable.Races, nil
}

// ResultsURL generates the results URL for a specific race
func ResultsURL(season, round string) string {
	return fmt.Sprintf("https://motorsportstats.com/results/formula-one/%s/round-%s", season, round)
}

func FetchSessionResults(ctx context.Context, season, round, sessionType string) ([]models.DriverResult, error) {
    var endpoint string
    switch sessionType {
    case "Race":
        endpoint = fmt.Sprintf("%s/%s/%s/results.json", ergastBase, season, round)
    case "Qualifying":
        endpoint = fmt.Sprintf("%s/%s/%s/qualifying.json", ergastBase, season, round)
    case "Sprint":
        endpoint = fmt.Sprintf("%s/%s/%s/sprint.json", ergastBase, season, round)
    default:
        return nil, fmt.Errorf("detailed results not available for practice sessions")
    }

    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
    }

    var data map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return nil, err
    }

    results := []models.DriverResult{}
    mrData := data["MRData"].(map[string]interface{})

    if sessionType == "Qualifying" {
        raceTable := mrData["RaceTable"].(map[string]interface{})
        races := raceTable["Races"].([]interface{})
        if len(races) == 0 {
            return nil, fmt.Errorf("no results available yet")
        }
        qualResults := races[0].(map[string]interface{})["QualifyingResults"].([]interface{})

        for _, r := range qualResults {
            result := r.(map[string]interface{})
            driver := result["Driver"].(map[string]interface{})
            constructor := result["Constructor"].(map[string]interface{})

            q3Time := "-"
            if q3, ok := result["Q3"].(string); ok && q3 != "" {
                q3Time = q3
            } else if q2, ok := result["Q2"].(string); ok && q2 != "" {
                q3Time = q2
            } else if q1, ok := result["Q1"].(string); ok && q1 != "" {
                q3Time = q1
            }

            results = append(results, models.DriverResult{
                Position:    result["position"].(string),
                Driver:      fmt.Sprintf("%s %s", driver["givenName"], driver["familyName"]),
                Constructor: constructor["name"].(string),
                Time:        q3Time,
                Status:      "Completed",
            })
        }
    } else {
        // Race or Sprint results
        raceTable := mrData["RaceTable"].(map[string]interface{})
        races := raceTable["Races"].([]interface{})
        if len(races) == 0 {
            return nil, fmt.Errorf("no results available yet")
        }

        var raceResults []interface{}
        race := races[0].(map[string]interface{})
        if sessionType == "Sprint" {
            raceResults = race["SprintResults"].([]interface{})
        } else {
            raceResults = race["Results"].([]interface{})
        }

        for _, r := range raceResults {
            result := r.(map[string]interface{})
            driver := result["Driver"].(map[string]interface{})
            constructor := result["Constructor"].(map[string]interface{})

            time := "-"
            if t, ok := result["Time"].(map[string]interface{}); ok {
                time = t["time"].(string)
            }

            // Add points for Race results
            points := ""
            if sessionType == "Race" {
                if p, ok := result["points"].(string); ok {
                    points = p
                }
            }

            results = append(results, models.DriverResult{
                Position:    result["position"].(string),
                Driver:      fmt.Sprintf("%s %s", driver["givenName"], driver["familyName"]),
                Constructor: constructor["name"].(string),
                Time:        time,
                Status:      result["status"].(string),
                Points:      points,  // New field
            })
        }
    }

    return results, nil
}
