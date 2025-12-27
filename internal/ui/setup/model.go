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
	Repo                *repository.HabitRepository
	Habits              []models.Habit
	Cursor              int
	Adding              bool
	Editing             bool
	ConfirmingArchive   bool
	ConfirmingUnarchive bool
	ConfirmingEdit      bool
	TextInput           textinput.Model
	Err                 error
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
	habits, err := m.Repo.List(true)
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
		if m.ConfirmingArchive || m.ConfirmingUnarchive || m.ConfirmingEdit {
			switch msg.String() {
			case "y", "Y":
				if m.Cursor >= 0 && m.Cursor < len(m.Habits) {
					habit := &m.Habits[m.Cursor]
					if m.ConfirmingArchive {
						habit.IsArchived = true
						if err := m.Repo.Update(habit); err != nil {
							m.Err = err
						}
					} else if m.ConfirmingUnarchive {
						habit.IsArchived = false
						if err := m.Repo.Update(habit); err != nil {
							m.Err = err
						}
					} else if m.ConfirmingEdit {
						// Update habit name
						str := m.TextInput.Value()
						if str == "" {
							// Don't restart, just keep it? Or use old name?
							// Original logic: if name != ""
							// If user cleared it, arguably we shouldn't save empty name.
						} else {
							habit.Name = str
							if err := m.Repo.Update(habit); err != nil {
								m.Err = err
							}
						}
						m.Editing = false
						m.TextInput.Reset()
					}
				}
				m.ConfirmingArchive = false
				m.ConfirmingUnarchive = false
				m.ConfirmingEdit = false
				return m, m.LoadHabits
			case "n", "N", "esc":
				m.ConfirmingArchive = false
				m.ConfirmingUnarchive = false
				m.ConfirmingEdit = false // Cancel confirmation
				// If we were editing, we remain in editing mode (implied by not setting Editing=false)
				// But we need to make sure we don't fall through to other handlers if we want to consume this key.
				return m, nil
			}
			// Ignore other keys while confirming? Or allow pass through?
			// Ideally ignore others to force Y/N choice.
			return m, nil
		}

		if m.Adding || m.Editing {
			switch msg.String() {
			case "enter":
				name := m.TextInput.Value()
				if name != "" {
					if m.Adding {
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
					} else if m.Editing {
						m.ConfirmingEdit = true
						return m, nil
					}
				}
			case "esc":
				m.Adding = false
				m.Editing = false
				m.ConfirmingEdit = false
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
			m.TextInput.Placeholder = "New Habit Name"
			m.TextInput.SetValue("")
			m.TextInput.Focus()
			return m, textinput.Blink
		case "e":
			if m.Cursor >= 0 && m.Cursor < len(m.Habits) {
				m.Editing = true
				m.TextInput.Placeholder = "Edit Habit Name"
				m.TextInput.SetValue(m.Habits[m.Cursor].Name)
				m.TextInput.Focus()
				return m, textinput.Blink
			}
		case "d":
			if m.Cursor >= 0 && m.Cursor < len(m.Habits) {
				if !m.Habits[m.Cursor].IsArchived {
					m.ConfirmingArchive = true
				}
			}
		case "u":
			if m.Cursor >= 0 && m.Cursor < len(m.Habits) {
				if m.Habits[m.Cursor].IsArchived {
					m.ConfirmingUnarchive = true
				}
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	s := ui.TitleStyle.Render("Setup Habits") + "\n\n"

	if m.Err != nil {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("Error: %v", m.Err)) + "\n\n"
	}

	if m.ConfirmingArchive {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true).Render("Are you sure you want to archive this habit? It will be hidden from the tracker. (y/n)") + "\n\n"
	} else if m.ConfirmingUnarchive {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true).Render("Are you sure you want to un-archive this habit? (y/n)") + "\n\n"
	} else if m.ConfirmingEdit {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true).Render("Are you sure you want to rename this habit? Previous records will be affected. (y/n)") + "\n\n"
	}

	s += ui.ListHeaderStyle.Render("Your Habits") + "\n"

	for i, habit := range m.Habits {
		cursor := " "
		style := ui.ListItemStyle
		if m.Cursor == i {
			cursor = ">"
			style = ui.SelectedListItemStyle
		}

		name := habit.Name
		if habit.IsArchived {
			name += " (Archived)"
			style = style.Faint(true)
		}

		s += style.Render(fmt.Sprintf("%s %s", cursor, name)) + "\n"
	}

	if m.Adding || m.Editing {
		title := "New Habit"
		if m.Editing {
			title = "Edit Habit"
		}
		s += "\n" + lipgloss.NewStyle().Bold(true).Render(title+":")
		s += "\n" + m.TextInput.View() + "\n"
		s += ui.HelpStyle.Render("(Enter to save, Esc to cancel)")
	} else {
		s += "\n" + ui.HelpStyle.Render("(n: new, e: edit, d: archive, u: un-archive, j/k: navigate)")
	}

	return s
}
