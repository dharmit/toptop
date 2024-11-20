package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	err    error
	out    string
	uptime string
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

	return fmt.Sprintf("%d:%02d", hh, mm), nil
}

// loadAvg returns load average as reported by uptime command
func loadAvg() ([3]float64, error) {
	f, err := os.Open("/proc/loadavg")
	if err != nil {
		return [3]float64{}, err
	}
	defer f.Close()

	var content []byte
	content, err = io.ReadAll(f)
	if err != nil {
		return [3]float64{}, err
	}

	var strContent string
	strContent = string(content)
	sliceContent := strings.Split(strContent, " ")

	var load [3]float64
	for i := 0; i < 3; i++ {
		load[i], err = strconv.ParseFloat(sliceContent[i], 64)
		if err != nil {
			return [3]float64{}, err
		}
	}
	return load, nil
}

func userSessions() (int, error) {
	var count int // keep a count of logged in users
	err := filepath.Walk("/run/systemd/sessions", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() {
			count++
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func topOut() (string, error) {
	uptime, err := uptime()
	if err != nil {
		return "", err
	}
	load, err := loadAvg()
	if err != nil {
		return "", err
	}
	users, err := userSessions()
	if err != nil {
		return "", err
	}
	lastUpdated := time.Now()
	return fmt.Sprintf("top - %s up %s,  %d users,  load average: %.2f, %.2f, %.2f\n\n",
		lastUpdated.Format(time.TimeOnly),
		uptime,
		users,
		load[0], load[1], load[2],
	), nil
}

func (m *model) Init() tea.Cmd {
	var err error
	m.out, err = topOut()
	if err != nil {
		return tea.Quit
	}
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
		var err error
		m.out, err = topOut()
		if err != nil {
			return m, tea.Quit
		}
		return m, tick()
	}
	return m, nil
}

func (m *model) View() string {
	if m.err != nil {
		return fmt.Sprintf("encountered error: %v\n", m.err)
	}
	return m.out
}

func main() {
	p := tea.NewProgram(&model{}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
