package setup

import (
	"fmt"
	"habit-tracker/internal/models"
	"habit-tracker/internal/repository"
	"habit-tracker/internal/ui"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Repo      *repository.HabitRepository
	Habits    []models.Habit
	Cursor    int
	Adding    bool
	TextInput textinput.Model
	Err       error
}

func NewModel(repo *repository.HabitRepository) Model {
	ti := textinput.New()
	ti.Placeholder = "New Habit Name"
	ti.Focus()

	return Model{
		Repo:      repo,
		TextInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return m.LoadHabits
}

func (m Model) LoadHabits() tea.Msg {
	habits, err := m.Repo.List()
	if err != nil {
		return err
	}
	return habits
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case []models.Habit:
		m.Habits = msg
		m.Err = nil

	case error:
		m.Err = msg

	case tea.KeyMsg:
		if m.Adding {
			switch msg.String() {
			case "enter":
				name := m.TextInput.Value()
				if name != "" {
					habit := &models.Habit{
						Name:       name,
						GoalTarget: 100, // Default goal
					}
					if err := m.Repo.Create(habit); err != nil {
						m.Err = err
					} else {
						m.Adding = false
						m.TextInput.Reset()
						return m, m.LoadHabits
					}
				}
			case "esc":
				m.Adding = false
				m.TextInput.Reset()
			}
			m.TextInput, cmd = m.TextInput.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Habits)-1 {
				m.Cursor++
			}
		case "n":
			m.Adding = true
			m.TextInput.Focus()
			return m, textinput.Blink
		}
	}

	return m, nil
}

func (m Model) View() string {
	s := ui.TitleStyle.Render("Setup Habits") + "\n\n"

	if m.Err != nil {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("Error: %v", m.Err)) + "\n\n"
	}

	s += ui.ListHeaderStyle.Render("Your Habits") + "\n"

	for i, habit := range m.Habits {
		cursor := " "
		style := ui.ListItemStyle
		if m.Cursor == i {
			cursor = ">"
			style = ui.SelectedListItemStyle
		}
		s += style.Render(fmt.Sprintf("%s %s", cursor, habit.Name)) + "\n"
	}

	if m.Adding {
		s += "\n" + m.TextInput.View() + "\n"
		s += ui.HelpStyle.Render("(Enter to save, Esc to cancel)")
	} else {
		s += "\n" + ui.HelpStyle.Render("(n: new habit, j/k: navigate)")
	}

	return s
}
