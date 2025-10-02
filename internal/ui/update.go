package ui

import (
	"errors"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kashifulhaque/f1-tui/internal/models"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()

		// If in results view, handle back navigation
		if m.showResults {
			switch s {
			case "esc", "q", "backspace":
				m.showResults = false
				m.resultsView = models.ResultsView{}
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}
			// Allow scrolling in results
			var cmd tea.Cmd
			m.resultsTbl, cmd = m.resultsTbl.Update(msg)
			return m, cmd
		}

		// Main view navigation
		switch s {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "left":
			m.selectIndex(m.idx - 1)
			return m, nil
		case "right":
			m.selectIndex(m.idx + 1)
			return m, nil
		case "enter":
			if m.tbl.Focused() && len(m.races) > 0 {
				selectedRow := m.tbl.SelectedRow()
				if len(selectedRow) > 0 {
					sessionName := selectedRow[0]
					r := m.races[m.idx]

					// Show loading and fetch results
					m.showResults = true
					m.resultsView = models.ResultsView{
						SessionName: sessionName,
						RaceName:    r.RaceName,
						Loading:     true,
					}

					return m, fetchResultsCmd(r.Season, r.Round, sessionName, sessionName, r.RaceName)
				}
			}
			return m, nil
		case "c":
			m.showCircuit = !m.showCircuit
			return m, nil
		case "r":
			m.loading = true
			return m, fetchCmd()
		}

	case resultsMsg:
	    m.resultsView.Loading = false
	    m.resultsView.Results = msg.results
	    m.resultsView.SessionName = msg.sessionName
	    m.resultsView.RaceName = msg.raceName

	    // Build table rows
	    rows := []table.Row{}
	    for _, res := range msg.results {
	        rows = append(rows, table.Row{
	            res.Position,
	            res.Driver,
	            res.Constructor,
	            res.Time,
	            res.Points,  // Add points column
	        })
	    }
	    m.resultsTbl.SetRows(rows)
	    m.resultsTbl.GotoTop()
	    return m, nil

	case resultsErrMsg:
		m.resultsView.Loading = false
		m.resultsView.Error = msg.err
		return m, nil

	case dataMsg:
		m.loading = false
		m.err = nil
		m.races = filterAndSortRaces(msg.races)
		if len(m.races) == 0 {
			m.err = errors.New("no races in season")
			return m, nil
		}
		m.idx = pickRelevantIndex(m.races)
		m.rebuild()
		return m, nil

	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, nil
	}

	// Delegate to table for navigation
	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)
	return m, cmd
}
