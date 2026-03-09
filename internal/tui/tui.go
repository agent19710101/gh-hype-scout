package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/agent19710101/gh-hype-scout/internal/rank"
)

type model struct {
	scored []rank.Repo
	active int
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
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
	bg := lipgloss.NewStyle().Background(lipgloss.Color("235")).Foreground(lipgloss.Color("252"))
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).Padding(0, 1)
	sub := lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	panel := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("62")).Padding(0, 1).Width(58)
	activePanel := panel.Copy().BorderForeground(lipgloss.Color("86"))

	title := header.Render("gh-hype-scout v0.8 TUI") + sub.Render("  live terminal scouting dashboard")
	meta := sub.Render(time.Now().Format("2006-01-02 15:04:05") + "  •  Tab switch panels  •  q quit")

	left := renderTop(m.scored)
	right := renderSignals(m.scored)
	if m.active == 0 {
		left = activePanel.Render(left)
		right = panel.Render(right)
	} else {
		left = panel.Render(left)
		right = activePanel.Render(right)
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	return bg.Render(lipgloss.JoinVertical(lipgloss.Left, title, meta, "", row))
}

func renderTop(scored []rank.Repo) string {
	var b strings.Builder
	b.WriteString("TOP REPOS\n\n")
	b.WriteString("#  repo                           stars   score   accel\n")
	for i, r := range scored {
		if i >= 12 {
			break
		}
		fmt.Fprintf(&b, "%-2d %-30s %-7d %-7.1f %+6.2f\n", i+1, truncate(r.FullName, 30), r.StargazersCount, r.HotScore, r.Acceleration)
	}
	return b.String()
}

func renderSignals(scored []rank.Repo) string {
	var b strings.Builder
	b.WriteString("MOMENTUM SIGNALS\n\n")
	type item struct {
		name  string
		accel float64
		stars int
	}
	top := make([]item, 0, len(scored))
	for _, r := range scored {
		top = append(top, item{name: r.FullName, accel: r.Acceleration, stars: r.StargazersCount})
	}
	for i := 0; i < len(top) && i < 10; i++ {
		if top[i].accel < 0.1 {
			continue
		}
		fmt.Fprintf(&b, "▲ %-30s  %+6.2f/h  %d★\n", truncate(top[i].name, 30), top[i].accel, top[i].stars)
	}
	if len(top) == 0 {
		b.WriteString("No momentum signals yet.\n")
	}
	b.WriteString("\nTIPS\n")
	b.WriteString("• use --watch for continuous updates\n")
	b.WriteString("• use --sort accel for momentum ranking\n")
	b.WriteString("• use --routing-profile teamA for policy pack routing\n")
	return b.String()
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return string(r[:n-1]) + "…"
}

func Show(scored []rank.Repo) error {
	p := tea.NewProgram(model{scored: scored}, tea.WithOutput(os.Stdout), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
