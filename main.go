package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	err error
	out []byte
}

func (m model) Init() tea.Cmd {
	return tick()
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	case tickMsg:
		c := exec.Command("uptime")
		out, err := c.Output()
		if err != nil {
			return m, tea.Quit
		}
		m.out = out
		return m, tick()
	}
	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("encountered error: %v\n", m.err)
	}

	return string(m.out)
}

func main() {
	p := tea.NewProgram(model{}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
