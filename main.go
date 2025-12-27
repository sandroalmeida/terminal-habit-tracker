package main

import (
	"fmt"
	"habit-tracker/internal/config"
	"habit-tracker/internal/db"
	"habit-tracker/internal/repository"
	"habit-tracker/internal/service"
	"habit-tracker/internal/ui/annual"
	"habit-tracker/internal/ui/monthly"
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
	viewMonthly
	viewAnnual
	viewSetup
)

type model struct {
	state        sessionState
	trackerModel tracker.Model
	monthlyModel monthly.Model
	annualModel  annual.Model
	setupModel   setup.Model
}

func initialModel(habitRepo *repository.HabitRepository, logRepo *repository.LogRepository, statsService *service.StatsService) model {
	return model{
		state:        viewTracker,
		trackerModel: tracker.NewModel(habitRepo, logRepo, statsService),
		monthlyModel: monthly.NewModel(habitRepo, logRepo, statsService),
		annualModel:  annual.NewModel(habitRepo, logRepo, statsService),
		setupModel:   setup.NewModel(habitRepo),
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.trackerModel.Init(), m.monthlyModel.Init(), m.annualModel.Init(), m.setupModel.Init())
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
				m.state = viewMonthly
				cmds = append(cmds, m.monthlyModel.LoadHabits, m.monthlyModel.LoadMonthLogs)
			} else if m.state == viewMonthly {
				m.state = viewAnnual
				cmds = append(cmds, m.annualModel.LoadHabits, m.annualModel.LoadAnnualLogs)
			} else if m.state == viewAnnual {
				m.state = viewSetup
				cmds = append(cmds, m.setupModel.LoadHabits)
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
	case viewMonthly:
		m.monthlyModel, cmd = m.monthlyModel.Update(msg)
		cmds = append(cmds, cmd)
	case viewAnnual:
		m.annualModel, cmd = m.annualModel.Update(msg)
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
	monthlyTab := "Monthly"
	annualTab := "Annual"
	setupTab := "Setup"

	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true).Border(lipgloss.NormalBorder(), false, false, true, false).BorderForeground(lipgloss.Color("205"))
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	if m.state == viewTracker {
		trackerTab = activeStyle.Render(trackerTab)
		monthlyTab = inactiveStyle.Render(monthlyTab)
		annualTab = inactiveStyle.Render(annualTab)
		setupTab = inactiveStyle.Render(setupTab)
	} else if m.state == viewMonthly {
		trackerTab = inactiveStyle.Render(trackerTab)
		monthlyTab = activeStyle.Render(monthlyTab)
		annualTab = inactiveStyle.Render(annualTab)
		setupTab = inactiveStyle.Render(setupTab)
	} else if m.state == viewAnnual {
		trackerTab = inactiveStyle.Render(trackerTab)
		monthlyTab = inactiveStyle.Render(monthlyTab)
		annualTab = activeStyle.Render(annualTab)
		setupTab = inactiveStyle.Render(setupTab)
	} else {
		trackerTab = inactiveStyle.Render(trackerTab)
		monthlyTab = inactiveStyle.Render(monthlyTab)
		annualTab = inactiveStyle.Render(annualTab)
		setupTab = activeStyle.Render(setupTab)
	}

	header := lipgloss.JoinHorizontal(lipgloss.Top, trackerTab, "  ", monthlyTab, "  ", annualTab, "  ", setupTab)
	header = lipgloss.NewStyle().MarginBottom(1).Render(header)

	content := ""
	switch m.state {
	case viewTracker:
		content = m.trackerModel.View()
	case viewMonthly:
		content = m.monthlyModel.View()
	case viewAnnual:
		content = m.annualModel.View()
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
	statsService := service.NewStatsService(logRepo)

	p := tea.NewProgram(initialModel(habitRepo, logRepo, statsService), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
