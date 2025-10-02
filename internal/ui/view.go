package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/kashifulhaque/f1-tui/internal/utils"
)

func (m Model) View() string {
	if m.showResults {
		return m.renderResultsView()
	}

	if m.loading {
		return TitleStyle.Render("F1 TUI") + "\n" + LabelStyle.Render("Loading schedule…")
	}

	if m.err != nil {
		return TitleStyle.Render("F1 TUI") + "\n" + ErrorStyle.Render(m.err.Error()) + "\nPress r to retry."
	}

	if len(m.races) == 0 {
		return "No races found."
	}

	r := m.races[m.idx]

	// Left pane
	left := RoundBadge.Render(fmt.Sprintf("ROUND %s", r.Round)) + "\n" +
			GPStyle.Render(fmt.Sprintf("%s", r.RaceName)) + "\n" +
			LabelStyle.Render(fmt.Sprintf("%s", r.Circuit.CircuitName)) + "\n"


	// Race card
	live := time.Now().After(m.race.Start) && time.Now().Before(m.race.End)
	raceHeader := "Race"
	if live {
		raceHeader += " " + LiveBadge.Render("LIVE")
	}

	card := lipgloss.NewStyle().BorderForeground(brightGreen).Width(36)
	left += card.Render(
		TitleStyle.Margin(0).Render(raceHeader) + "\n" +
			lipgloss.NewStyle().Bold(true).Render(m.race.Start.Format("Jan 2 Mon 15:04")),
	)

	// Circuit pane (optional)
	circuit := ""
	if m.showCircuit {
		c := r.Circuit
		facts := []string{
			fmt.Sprintf("Circuit: %s", c.CircuitName),
			fmt.Sprintf("Location: %s, %s", c.Location.Locality, c.Location.Country),
		}
		circuit = lipgloss.NewStyle().MarginTop(1).Render(strings.Join(facts, "\n"))

		if utils.HasChafa() {
			outline := utils.FetchCircuitSVG(c.URL)
			if outline != "" {
				art := utils.RenderWithChafa(outline)
				if art != "" {
					circuit += "\n\n" + art
				}
			}
		}
	}

	leftPane := left
	if circuit != "" {
		leftPane += "\n" + circuit
	}

	// Right pane: sessions table
	rightTitle := TitleStyle.Render("Sessions") + "\n"
	right := rightTitle + m.tbl.View()

	// Footer
	footer := LabelStyle.Render("←/→ switch GP • ↑/↓ sessions • Enter view results • c circuit • r refresh • q quit")

	// Layout
	gap := 3
	leftW := 42
	ui := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(leftW).Render(leftPane),
		strings.Repeat(" ", gap),
		right,
	) + "\n\n" + footer

	return ui
}

func (m Model) renderResultsView() string {
    if m.resultsView.Loading {
        return TitleStyle.Render("Loading Results...") + "\n" +
            LabelStyle.Render("Fetching session data...")
    }

    if m.resultsView.Error != nil {
        // Check if race is currently live
        now := time.Now()
        isLive := now.After(m.race.Start) && now.Before(m.race.End)

        errorMsg := m.resultsView.Error.Error()
        if isLive && m.resultsView.SessionName == "Race" {
            errorMsg = "Live timing not available - results will appear after race completion"
        }

        return TitleStyle.Render(m.resultsView.RaceName) + "\n" +
            GPStyle.Render(m.resultsView.SessionName) + "\n" +
            LabelStyle.Render(errorMsg) + "\n" +
            LabelStyle.Render("Press ESC or Q to go back")
    }

    header := TitleStyle.Render(m.resultsView.RaceName) + "\n" +
        GPStyle.Render(m.resultsView.SessionName + " Results") + "\n\n"

    footer := "\n" + LabelStyle.Render("↑/↓ scroll • ESC/Q go back • Ctrl+C quit")

    return header + m.resultsTbl.View() + footer
}
