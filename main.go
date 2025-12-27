package main

import (
	"fmt"
	"habit-tracker/internal/config"
	"habit-tracker/internal/db"
	"habit-tracker/internal/repository"
	"habit-tracker/internal/ui/setup"
	"habit-tracker/internal/ui/tracker"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sessionState int

const (
	viewTracker sessionState = iota
	viewSetup
)

type model struct {
	state        sessionState
	trackerModel tracker.Model
	setupModel   setup.Model
}

func initialModel(habitRepo *repository.HabitRepository, logRepo *repository.LogRepository) model {
	return model{
		state:        viewTracker,
		trackerModel: tracker.NewModel(habitRepo, logRepo),
		setupModel:   setup.NewModel(habitRepo),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.trackerModel.Init(), m.setupModel.Init())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.state == viewTracker {
				m.state = viewSetup
				// Refresh habits when switching to setup? Or when switching back?
				// Actually, when switching BACK to tracker, we should reload habits in case changes happened.
			} else {
				m.state = viewTracker
				// Reload habits and logs when returning to tracker
				cmds = append(cmds, m.trackerModel.LoadHabits, m.trackerModel.LoadLogs)
			}
		}
	}

	switch m.state {
	case viewTracker:
		m.trackerModel, cmd = m.trackerModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewSetup:
		m.setupModel, cmd = m.setupModel.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	// Simple tab bar
	trackerTab := "Tracker"
	setupTab := "Setup"

	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Border(lipgloss.NormalBorder(), false, false, true, false).BorderForeground(lipgloss.Color("205"))
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	if m.state == viewTracker {
		trackerTab = activeStyle.Render(trackerTab)
		setupTab = inactiveStyle.Render(setupTab)
	} else {
		trackerTab = inactiveStyle.Render(trackerTab)
		setupTab = activeStyle.Render(setupTab)
	}

	header := lipgloss.JoinHorizontal(lipgloss.Top, trackerTab, "  ", setupTab)
	header = lipgloss.NewStyle().MarginBottom(1).Render(header)

	content := ""
	switch m.state {
	case viewTracker:
		content = m.trackerModel.View()
	case viewSetup:
		content = m.setupModel.View()
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, content)
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer database.Close()

	habitRepo := repository.NewHabitRepository(database)
	logRepo := repository.NewLogRepository(database)

	p := tea.NewProgram(initialModel(habitRepo, logRepo), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
