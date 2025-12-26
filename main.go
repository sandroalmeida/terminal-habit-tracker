package main

import (
	"fmt"
	"habit-tracker/internal/config"
	"habit-tracker/internal/db"
	"habit-tracker/internal/repository"
	"habit-tracker/internal/ui/setup"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	setupModel setup.Model
}

func initialModel(repo *repository.HabitRepository) model {
	return model{
		setupModel: setup.NewModel(repo),
	}
}

func (m model) Init() tea.Cmd {
	return m.setupModel.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	// For now, directly forward to setup model since it's the only view
	m.setupModel, cmd = m.setupModel.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.setupModel.View()
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

	repo := repository.NewHabitRepository(database)

	p := tea.NewProgram(initialModel(repo))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
