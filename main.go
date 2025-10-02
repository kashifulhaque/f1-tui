package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/browser"
)

// -----------------------------
// Data layer (Ergast API)
// -----------------------------

const ergastBase = "http://api.jolpi.ca/ergast/f1"

type ErgastMRData struct {
	RaceTable struct {
		Season string  `json:"season"`
		Races  []Race  `json:"Races"`
	} `json:"RaceTable"`
}

type Race struct {
	Season       string   `json:"season"`
	Round        string   `json:"round"`
	RaceName     string   `json:"raceName"`
	Circuit      Circuit  `json:"Circuit"`
	Date         string   `json:"date"`       // race date (UTC)
	Time         string   `json:"time"`       // race time (UTC)
	FirstPractice *Session `json:"FirstPractice,omitempty"`
	SecondPractice *Session `json:"SecondPractice,omitempty"`
	ThirdPractice  *Session `json:"ThirdPractice,omitempty"`
	Sprint         *Session `json:"Sprint,omitempty"`
	SprintShootout *Session `json:"SprintShootout,omitempty"`
	Qualifying     *Session `json:"Qualifying,omitempty"`
}

type Circuit struct {
	CircuitName string `json:"circuitName"`
	Location    struct {
		Locality  string `json:"locality"`
		Country   string `json:"country"`
		Lat       string `json:"lat"`
		Long      string `json:"long"`
	} `json:"Location"`
	URL string `json:"url"` // Wikipedia
}

type Session struct {
	Date string `json:"date"`
	Time string `json:"time"`
}

// Typed session for UI
type UISession struct {
	Kind  string    // P1/P2/P3/Qualifying/Sprint/Race
	Start time.Time // local time
	End   time.Time // local time (approx duration)
	URL   string    // results link
}

// Fetch current season schedule
func fetchCurrentSchedule(ctx context.Context) ([]Race, error) {
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
		MRData ErgastMRData `json:"MRData"`
	}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&outer); err != nil {
		return nil, err
	}
	return outer.MRData.RaceTable.Races, nil
}

// Helpers
func parseUTC(date, t string) (time.Time, error) {
	// Ergast time often ends with "Z"
	if t == "" {
		return time.Time{}, errors.New("empty time")
	}
	stamp := fmt.Sprintf("%sT%s", date, strings.TrimSuffix(t, "Z"))
	// Try a few layouts
	layouts := []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02T15:04"}
	var last error
	for _, l := range layouts {
		if ts, err := time.Parse(l, stamp); err == nil {
			return ts.UTC(), nil
		} else {
			last = err
		}
	}
	return time.Time{}, last
}

func toLocal(t time.Time) time.Time { return t.In(time.Local) }

func approxEnd(kind string, start time.Time) time.Time {
	switch kind {
	case "Race":
		return start.Add(2*time.Hour + 5*time.Minute)
	case "Qualifying", "Sprint":
		return start.Add(1 * time.Hour)
	default:
		return start.Add(90 * time.Minute)
	}
}

func resultsURL(season, round string) string {
	// Using Motorsport Stats public results pages for reliability
	return fmt.Sprintf("https://motorsportstats.com/results/formula-one/%s/round-%s", season, round)
}

// Map Ergast race -> UI sessions
func buildUISessions(r Race) ([]UISession, UISession, error) {
	var out []UISession
	season, round := r.Season, r.Round
	add := func(kind string, s *Session) error {
		if s == nil || s.Time == "" {
			return nil
		}
		utc, err := parseUTC(s.Date, s.Time)
		if err != nil { return err }
		start := toLocal(utc)
		out = append(out, UISession{
			Kind:  kind,
			Start: start,
			End:   approxEnd(kind, start),
			URL:   resultsURL(season, round),
		})
		return nil
	}
	var err error
	if err = add("Practice 1", r.FirstPractice); err != nil { return nil, UISession{}, err }
	if err = add("Practice 2", r.SecondPractice); err != nil { return nil, UISession{}, err }
	if err = add("Practice 3", r.ThirdPractice); err != nil { return nil, UISession{}, err }
	if err = add("Sprint Shootout", r.SprintShootout); err != nil { return nil, UISession{}, err }
	if err = add("Sprint", r.Sprint); err != nil { return nil, UISession{}, err }
	if err = add("Qualifying", r.Qualifying); err != nil { return nil, UISession{}, err }
	// Race
	raceUTC, err := parseUTC(r.Date, r.Time)
	if err != nil { return nil, UISession{}, err }
	raceLocal := toLocal(raceUTC)
	race := UISession{
		Kind:  "Race",
		Start: raceLocal,
		End:   approxEnd("Race", raceLocal),
		URL:   resultsURL(season, round),
	}
	return out, race, nil
}

// -----------------------------
// UI Layer (Bubble Tea)
// -----------------------------

var (
	accent  = lipgloss.AdaptiveColor{Light: "#0EA5E9", Dark: "#22D3EE"}
	okGreen = lipgloss.AdaptiveColor{Light: "#10B981", Dark: "#34D399"}
	errRed  = lipgloss.AdaptiveColor{Light: "#EF4444", Dark: "#F87171"}
	muted  = lipgloss.AdaptiveColor{Light: "#64748B", Dark: "#94A3B8"}
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(accent).MarginBottom(1)
	gpStyle      = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A7F3D0")).MarginBottom(1)
	cardStyle    = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(okGreen)
	labelStyle   = lipgloss.NewStyle().Faint(true).Foreground(muted)
	liveBadge    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#111827")).Background(lipgloss.Color("#F59E0B")).Padding(0, 1)
	errorStyle   = lipgloss.NewStyle().Foreground(errRed)
)

type sessionRow struct{ Title, Time string }

type model struct {
	loading bool
	err     error

	season string
	races  []Race
	idx    int // selected round index

	sessions []sessionRow
	race     UISession

	tbl   table.Model
	showCircuit bool
	filter textinput.Model
}

func initialModel() model {
	columns := []table.Column{{Title: "Session", Width: 18}, {Title: "Local Time", Width: 22}}
	t := table.New(table.WithColumns(columns), table.WithFocused(true))
	t.SetHeight(9)
	inp := textinput.New()
	inp.Placeholder = "Filter by GP name…"
	inp.Prompt = ""
	return model{loading: true, tbl: t, filter: inp}
}

// Messages

type dataMsg struct{ season string; races []Race }
type errMsg struct{ err error }

type refreshMsg struct{}

type toggleCircuitMsg struct{}

func fetchCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		r, err := fetchCurrentSchedule(ctx)
		if err != nil { return errMsg{err} }
		season := time.Now().Format("2006")
		return dataMsg{season: season, races: r}
	}
}

func (m model) Init() tea.Cmd { return fetchCmd() }

func (m *model) selectIndex(i int) {
	if i < 0 || i >= len(m.races) { return }
	m.idx = i
	m.rebuild()
}

func (m *model) rebuild() {
	if len(m.races) == 0 { return }
	r := m.races[m.idx]
	s, race, err := buildUISessions(r)
	if err != nil { m.err = err; return }
	m.err = nil
	m.season = r.Season
	m.sessions = nil
	for _, x := range s { m.sessions = append(m.sessions, sessionRow{Title: x.Kind, Time: x.Start.Format("Mon 15:04 - 16:04")}) }
	// Right-side table rows
	rows := []table.Row{}
	for _, x := range s {
		rows = append(rows, table.Row{x.Kind, x.Start.Format("Jan _2 Mon 15:04") + " - " + x.End.Format("15:04")})
	}
	m.tbl.SetRows(rows)
	m.tbl.GotoTop()
	m.race = race
}

func (m model) View() string {
	if m.loading {
		return titleStyle.Render("F1 TUI") + "\n" + labelStyle.Render("Loading schedule…")
	}
	if m.err != nil {
		return titleStyle.Render("F1 TUI") + "\n" + errorStyle.Render(m.err.Error()) + "\nPress r to retry."
	}
	if len(m.races) == 0 {
		return "No races found."
	}
	r := m.races[m.idx]

	// Left pane
	left := gpStyle.Render(fmt.Sprintf("%s", r.RaceName)) + "\n" +
		labelStyle.Render(fmt.Sprintf("%s", r.Circuit.CircuitName)) + "\n\n"

	// Race card
	live := time.Now().After(m.race.Start) && time.Now().Before(m.race.End)
	raceHeader := "Race"
	if live { raceHeader += " " + liveBadge.Render("LIVE") }
	card := lipgloss.NewStyle().BorderForeground(okGreen).Width(36)
	left += card.Render(
		titleStyle.Margin(0).Render(raceHeader) + "\n" +
		lipgloss.NewStyle().Bold(true).Render(m.race.Start.Format("Jan 2 Mon 15:04")),
	) + "\n\n" + labelStyle.Render("Enter on \"Race\" opens results – live view later")

	// Circuit pane (optional)
	circuit := ""
	if m.showCircuit {
		c := r.Circuit
		facts := []string{
			fmt.Sprintf("Circuit: %s", c.CircuitName),
			fmt.Sprintf("Location: %s, %s", c.Location.Locality, c.Location.Country),
			fmt.Sprintf("Wiki: %s", c.URL),
		}
		circuit = lipgloss.NewStyle().MarginTop(1).Render(strings.Join(facts, "\n"))

		// Best-effort: try chafa to render an SVG outline if supported (optional)
		if hasChafa() {
			outline := fetchCircuitSVG(c.URL)
			if outline != "" {
				art := renderWithChafa(outline)
				if art != "" { circuit += "\n\n" + art }
			}
		}
	}

	leftPane := left
	if circuit != "" { leftPane += "\n" + circuit }

	// Right pane: sessions table
	rightTitle := titleStyle.Render("Sessions") + "\n"
	right := rightTitle + m.tbl.View()

	// Footer
	footer := labelStyle.Render("←/→ switch GP • Enter open results • c circuit • r refresh • q quit")

	// Layout
	w := lipgloss.Width
	gap := 3
	leftW := 42
	ui := lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(leftW).Render(leftPane), strings.Repeat(" ", gap), right) + "\n\n" + footer
	_ = w
	return ui
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()
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
			// If table focused and a row selected -> open results
			if m.tbl.Focused() {
				openURL(m.race.URL)
				return m, nil
			}
			return m, nil
		case "c":
			m.showCircuit = !m.showCircuit
			return m, nil
		case "r":
			m.loading = true
			return m, fetchCmd()
		}
	case dataMsg:
		m.loading = false
		m.err = nil
		m.races = filterAndSortRaces(msg.races)
		if len(m.races) == 0 {
			m.err = errors.New("no races in season")
			return m, nil
		}
		// Pick next upcoming or last race
		m.idx = pickRelevantIndex(m.races)
		m.rebuild()
		return m, nil
	case errMsg:
		m.loading = false
		m.err = msg.err
		return m, nil
	}
	// Delegate to table for nav
	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)
	return m, cmd
}

func pickRelevantIndex(races []Race) int {
	now := time.Now()
	idx := 0
	bestDelta := time.Duration(1<<62 - 1)
	for i, r := range races {
		ts, err := parseUTC(r.Date, r.Time)
		if err != nil { continue }
		local := toLocal(ts)
		d := local.Sub(now)
		if d >= -4*time.Hour && d < bestDelta { // choose closest upcoming or very recent
			bestDelta = d
			idx = i
		}
	}
	return idx
}

func filterAndSortRaces(in []Race) []Race {
	out := make([]Race, 0, len(in))
	for _, r := range in {
		if r.RaceName == "" || r.Circuit.CircuitName == "" { continue }
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool {
		// sort by round number (string -> int)
		return out[i].Round < out[j].Round
	})
	return out
}

// -----------------------------
// Circuit outline helpers (best-effort)
// -----------------------------

func hasChafa() bool {
	_, err := exec.LookPath("chafa")
	return err == nil
}

func fetchCircuitSVG(wikiURL string) string {
	// Very light heuristic: replace wiki page with the first SVG on the page via Wikipedia REST summary (no guarantee)
	// To keep dependencies tiny/offline, we just return empty and rely on the link; real impl would parse
	return ""
}

func renderWithChafa(svgPath string) string {
	if svgPath == "" { return "" }
	cmd := exec.Command("chafa", svgPath, "--size=40x12")
	out, err := cmd.Output()
	if err != nil { return "" }
	return string(out)
}

func openURL(url string) {
	_ = browser.OpenURL(url)
}

// ---------------
// main
// ---------------

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
}
