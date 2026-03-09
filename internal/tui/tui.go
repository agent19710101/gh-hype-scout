package tui

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/agent19710101/gh-hype-scout/internal/rank"
)

type model struct {
	content string
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.KeyMsg:
		s := strings.ToLower(v.String())
		if s == "q" || s == "esc" || s == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	return m.content + "\n\nPress q to quit."
}

func Show(scored []rank.Repo) error {
	var b strings.Builder
	b.WriteString("gh-hype-scout (TUI)\n\n")
	b.WriteString("RANK  REPO  STARS  SCORE  ACCEL\n")
	for i, r := range scored {
		fmt.Fprintf(&b, "%d  %s  %d  %.1f  %.2f\n", i+1, r.FullName, r.StargazersCount, r.HotScore, r.Acceleration)
	}
	p := tea.NewProgram(model{content: b.String()}, tea.WithOutput(os.Stdout))
	_, err := p.Run()
	return err
}
