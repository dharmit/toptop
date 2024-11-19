package main

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	err error
	out []byte
}

// uptime reads the file /proc/uptime and returns uptime in the HH:MM format.
func uptime() (string, error) {
	f, err := os.Open("/proc/uptime")
	if err != nil {
		return "", err
	}
	defer f.Close()

	var content []byte
	content, err = io.ReadAll(f)
	if err != nil {
		return "", err
	}

	var strContent string
	strContent = string(content)
	sliceContent := strings.Split(strContent, " ")

	upt, err := strconv.ParseFloat(sliceContent[0], 64)
	if err != nil {
		return "", err
	}

	hh := int(upt / 3600)
	mm := int(math.Mod(upt, 3600) / 60)

	return fmt.Sprintf("%d:%d", hh, mm), nil
}

func (m *model) Init() tea.Cmd {
	c := exec.Command("uptime")
	out, err := c.Output()
	if err != nil {
		return tea.Quit
	}
	m.out = out
	return tick()
}

type tickMsg time.Time

func tick() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *model) View() string {
	if m.err != nil {
		return fmt.Sprintf("encountered error: %v\n", m.err)
	}
	return string(m.out)
}

func main() {
	p := tea.NewProgram(&model{}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
