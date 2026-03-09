package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/agent19710101/gh-hype-scout/internal/rank"
)

type model struct {
	scored  []rank.Repo
	active  int
	width   int
	height  int
	compact bool
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = v.Width
		m.height = v.Height
		m.compact = v.Width < 120
	case tea.KeyMsg:
		s := strings.ToLower(v.String())
		switch s {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		case "tab", "right", "l":
			m.active = (m.active + 1) % 2
		case "left", "h":
			m.active = (m.active + 1) % 2
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		m.width = 140
	}
	bg := lipgloss.NewStyle().Background(lipgloss.Color("235")).Foreground(lipgloss.Color("252"))
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Padding(0, 1)
	sub := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	panelWidth := (m.width - 6) / 2
	if m.compact {
		panelWidth = m.width - 4
	}
	if panelWidth < 56 {
		panelWidth = 56
	}
	panel := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(0, 1).Width(panelWidth)
	activePanel := panel.Copy().BorderForeground(lipgloss.Color("86"))

	title := header.Render("gh-hype-scout v1.0 TUI") + sub.Render("  live terminal scouting dashboard")
	meta := sub.Render(time.Now().Format("2006-01-02 15:04:05") + "  •  Tab switch panels  •  q quit")

	repoCol := panelWidth - 27
	if repoCol < 14 {
		repoCol = 14
	}
	left := renderTop(m.scored, repoCol)
	right := renderSignals(m.scored, repoCol)

	if m.active == 0 {
		left = activePanel.Render(left)
		right = panel.Render(right)
	} else {
		left = panel.Render(left)
		right = activePanel.Render(right)
	}

	var row string
	if m.compact {
		row = lipgloss.JoinVertical(lipgloss.Left, left, "", right)
	} else {
		row = lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	}
	return bg.Render(lipgloss.JoinVertical(lipgloss.Left, title, meta, "", row))
}

func renderTop(scored []rank.Repo, repoCol int) string {
	var b strings.Builder
	b.WriteString("TOP REPOS\n\n")
	b.WriteString(fmt.Sprintf("#  %-*s  stars   score   accel\n", repoCol, "repo"))
	for i, r := range scored {
		if i >= 12 {
			break
		}
		fmt.Fprintf(&b, "%-2d %-*s  %-7d %-7.1f %+6.2f\n", i+1, repoCol, fitWidth(r.FullName, repoCol), r.StargazersCount, r.HotScore, r.Acceleration)
	}
	return b.String()
}

func renderSignals(scored []rank.Repo, repoCol int) string {
	var b strings.Builder
	b.WriteString("MOMENTUM SIGNALS\n\n")
	type item struct{ name string; accel float64; stars int }
	top := make([]item, 0, len(scored))
	for _, r := range scored {
		top = append(top, item{name: r.FullName, accel: r.Acceleration, stars: r.StargazersCount})
	}
	printed := 0
	for i := 0; i < len(top) && printed < 10; i++ {
		if top[i].accel < 0.1 {
			continue
		}
		fmt.Fprintf(&b, "▲ %-*s  %+6.2f/h  %d★\n", repoCol, fitWidth(top[i].name, repoCol), top[i].accel, top[i].stars)
		printed++
	}
	if printed == 0 {
		b.WriteString("No momentum signals yet.\n")
	}
	b.WriteString("\nTIPS\n")
	b.WriteString("• stdout mode remains default\n")
	b.WriteString("• use --watch for continuous updates\n")
	b.WriteString("• use --sort accel --momentum-model trend\n")
	return b.String()
}

func fitWidth(s string, width int) string {
	if width <= 1 {
		return ""
	}
	if runewidth.StringWidth(s) <= width {
		return s
	}
	out := ""
	for _, r := range s {
		next := out + string(r)
		if runewidth.StringWidth(next+"…") > width {
			return out + "…"
		}
		out = next
	}
	return out
}

func Show(scored []rank.Repo) error {
	p := tea.NewProgram(model{scored: scored}, tea.WithOutput(os.Stdout), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
