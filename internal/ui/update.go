package ui

import (
	"errors"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kashifulhaque/f1-tui/internal/models"
	"github.com/kashifulhaque/f1-tui/internal/utils"
)

const liveRefreshInterval = 5 * time.Second

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()

		if m.showResults {
			switch s {
			case "esc", "q", "backspace":
				m.showResults = false
				m.resultsView = models.ResultsView{}
				m.resultsSession = models.UISession{}
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.resultsTbl, cmd = m.resultsTbl.Update(msg)
			return m, cmd
		}

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
					session, ok := m.sessionByName[sessionName]
					if !ok {
						session = m.race
					}
					isLive := utils.IsLiveSession(session, time.Now())

					m.showResults = true
					m.resultsView = models.ResultsView{
						SessionName: sessionName,
						RaceName:    r.RaceName,
						Loading:     true,
						Live:        isLive,
					}
					m.resultsSession = session
					m.resultsTbl = newResultsTable(isLive)

					if isLive {
						return m, fetchLiveResultsCmd(session, r.RaceName)
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
		m.resultsView.Live = false
		m.resultsView.UpdatedAt = time.Time{}

		rows := []table.Row{}
		for _, res := range msg.results {
			rows = append(rows, table.Row{
				res.Position,
				res.Driver,
				res.Constructor,
				res.Time,
				res.Points,
			})
		}
		m.resultsTbl.SetRows(rows)
		m.resultsTbl.GotoTop()
		return m, nil

	case liveResultsMsg:
		m.resultsView.Loading = false
		m.resultsView.Results = msg.results
		m.resultsView.SessionName = msg.sessionName
		m.resultsView.RaceName = msg.raceName
		m.resultsView.Live = true
		m.resultsView.UpdatedAt = msg.updated

		rows := []table.Row{}
		for _, res := range msg.results {
			rows = append(rows, table.Row{
				res.Position,
				res.Driver,
				res.Gap,
				res.Laps,
				res.Speed,
				res.Progress,
			})
		}
		m.resultsTbl.SetRows(rows)
		m.resultsTbl.GotoTop()
		return m, liveTickCmd()

	case resultsErrMsg:
		m.resultsView.Loading = false
		m.resultsView.Error = msg.err
		m.resultsView.Live = false
		m.resultsView.UpdatedAt = time.Time{}
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

	case liveTickMsg:
		if !m.showResults || !m.resultsView.Live {
			return m, nil
		}
		if !utils.IsLiveSession(m.resultsSession, time.Now()) {
			return m, nil
		}
		return m, fetchLiveResultsCmd(m.resultsSession, m.resultsView.RaceName)
	}

	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)
	return m, cmd
}

func liveTickCmd() tea.Cmd {
	return tea.Tick(liveRefreshInterval, func(t time.Time) tea.Msg {
		return liveTickMsg(t)
	})
}
