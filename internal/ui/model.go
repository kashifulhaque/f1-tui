package ui

import (
	"strconv"
	"context"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/kashifulhaque/f1-tui/internal/api"
	"github.com/kashifulhaque/f1-tui/internal/models"
	"github.com/kashifulhaque/f1-tui/internal/utils"
)

type Model struct {
	loading       bool
	err           error
	season        string
	races         []models.Race
	idx           int
	sessions      []models.SessionRow
	race          models.UISession
	tbl           table.Model
	showCircuit   bool
	filter        textinput.Model

	showResults   bool
	resultsView   models.ResultsView
	resultsTbl    table.Model
}

type dataMsg struct {
	season string
	races  []models.Race
}

type resultsMsg struct {
	sessionName string
	raceName    string
	results     []models.DriverResult
}

type resultsErrMsg struct{ err error }
type errMsg struct{ err error }
type refreshMsg struct{}
type toggleCircuitMsg struct{}

func InitialModel() Model {
	columns := []table.Column{
		{Title: "Session", Width: 18},
		{Title: "Local Time", Width: 30},
	}
	t := table.New(table.WithColumns(columns), table.WithFocused(true))
	t.SetHeight(9)

	resultColumns := []table.Column{
		{Title: "Pos", Width: 4},
		{Title: "Driver", Width: 24},
		{Title: "Team", Width: 22},
		{Title: "Time", Width: 16},
		{Title: "Pts", Width: 4},
	}
	resultsTbl := table.New(table.WithColumns(resultColumns), table.WithFocused(true))
	resultsTbl.SetHeight(20)

	inp := textinput.New()
	inp.Placeholder = "Filter by GP name…"
	inp.Prompt = ""

	return Model{
		loading:    true,
		tbl:        t,
		resultsTbl: resultsTbl,
		filter:     inp,
	}
}

func (m Model) Init() tea.Cmd {
	return fetchCmd()
}

func fetchCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		races, err := api.FetchCurrentSchedule(ctx)
		if err != nil {
			return errMsg{err}
		}
		season := time.Now().Format("2006")
		return dataMsg{season: season, races: races}
	}
}

func fetchResultsCmd(season, round, sessionType, sessionName, raceName string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		results, err := api.FetchSessionResults(ctx, season, round, sessionType)
		if err != nil {
			return resultsErrMsg{err}
		}
		return resultsMsg{
			sessionName: sessionName,
			raceName:    raceName,
			results:     results,
		}
	}
}

func (m *Model) selectIndex(i int) {
	if i < 0 || i >= len(m.races) {
		return
	}
	m.idx = i
	m.rebuild()
}

func (m *Model) rebuild() {
    if len(m.races) == 0 {
        return
    }

    r := m.races[m.idx]
    sessions, race, err := utils.BuildUISessions(r, r.Season, r.Round, api.ResultsURL)
    if err != nil {
        m.err = err
        return
    }

    m.err = nil
    m.season = r.Season
    m.sessions = nil

    allSessions := append(sessions, race)

    for _, s := range allSessions {
        m.sessions = append(m.sessions, models.SessionRow{
            Title: s.Kind,
            Time:  s.Start.Format("Mon 15:04 - 16:04"),
        })
    }

    var rows []table.Row
    for _, s := range allSessions {
        rows = append(rows, table.Row{
            s.Kind,
            s.Start.Format("Jan _2 Mon 15:04") + " - " + s.End.Format("15:04"),
        })
    }

    m.tbl.SetRows(rows)
    m.tbl.GotoTop()
    m.race = race
}

func pickRelevantIndex(races []models.Race) int {
	now := time.Now()
	idx := 0
	bestDelta := time.Duration(1<<62 - 1)

	for i, r := range races {
		ts, err := utils.ParseUTC(r.Date, r.Time)
		if err != nil {
			continue
		}
		local := utils.ToLocal(ts)
		d := local.Sub(now)
		if d >= -4*time.Hour && d < bestDelta {
			bestDelta = d
			idx = i
		}
	}
	return idx
}

func filterAndSortRaces(in []models.Race) []models.Race {
	out := make([]models.Race, 0, len(in))
	for _, r := range in {
		if r.RaceName == "" || r.Circuit.CircuitName == "" {
			continue
		}
		out = append(out, r)
	}

	sort.Slice(out, func(i, j int) bool {
		ri, err1 := strconv.Atoi(out[i].Round)
		rj, err2 := strconv.Atoi(out[j].Round)
		if err1 != nil || err2 != nil {
			return out[i].Round < out[j].Round
		}
		return ri < rj
	})
	return out
}
