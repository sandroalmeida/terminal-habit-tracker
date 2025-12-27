package setup

import (
	"fmt"
	"habit-tracker/internal/models"
	"habit-tracker/internal/repository"
	"habit-tracker/internal/ui"
	"strconv"

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
	InputtingGoal       bool
	TempName            string
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
						// Update habit name and goal using stored TempName and TextInput (Goal)
						// TempName was set when transitioning to Goal input
						goalVal, err := strconv.Atoi(m.TextInput.Value())
						// Although we validate before confirmation, re-check to be safe, or just trust the previous step?
						// Flow: Name -> Enter (store TempName) -> Goal -> Enter (validate) -> Confirm -> Y (Update)

						// If we are here, it means we are in ConfirmingEdit state.
						// BUT wait, my previous logic in "enter" handler handled the validation.
						// The ConfirmingEdit state is just a "Are you sure?".
						// So at this point, m.TextInput has the Goal value.
						if err != nil || goalVal < 1 || goalVal > 7 {
							m.Err = fmt.Errorf("invalid goal value: %v", err)
						} else {
							habit.Name = m.TempName
							habit.GoalTarget = goalVal
							if err := m.Repo.Update(habit); err != nil {
								m.Err = err
							}
						}

						m.Editing = false
						m.InputtingGoal = false
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
				// Stay in editing mode?
				// If we cancel the "Are you sure to edit?" prompt, typically we should just go back to the editing flow
				// OR just cancel the whole edit. Let's cancel the whole edit for simplicity.
				m.Editing = false
				m.InputtingGoal = false
				m.TextInput.Reset()
				return m, nil
			}
			return m, nil
		}

		if m.Adding || m.Editing {
			switch msg.String() {
			case "enter":
				val := m.TextInput.Value()
				if !m.InputtingGoal {
					// Step 1: Name Input
					if val != "" {
						m.TempName = val
						m.InputtingGoal = true
						m.TextInput.Reset()
						m.TextInput.Placeholder = "Weekly Goal (1-7)"

						// Pre-fill existing goal if editing
						if m.Editing && m.Cursor >= 0 && m.Cursor < len(m.Habits) {
							m.TextInput.SetValue(strconv.Itoa(m.Habits[m.Cursor].GoalTarget))
						} else {
							m.TextInput.SetValue("7") // Default
						}
						return m, nil
					}
				} else {
					// Step 2: Goal Input
					goalVal, err := strconv.Atoi(val)
					if err != nil || goalVal < 1 || goalVal > 7 {
						m.Err = fmt.Errorf("goal must be a number between 1 and 7")
						return m, nil // Don't proceed
					}
					m.Err = nil // Clear previous error

					if m.Adding {
						habit := &models.Habit{
							Name:       m.TempName,
							GoalTarget: goalVal,
						}
						if err := m.Repo.Create(habit); err != nil {
							m.Err = err
						} else {
							m.Adding = false
							m.InputtingGoal = false
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
				m.InputtingGoal = false
				m.ConfirmingEdit = false
				m.TextInput.Reset()
				m.Err = nil
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
			m.InputtingGoal = false
			m.TextInput.Placeholder = "New Habit Name"
			m.TextInput.SetValue("")
			m.TextInput.Focus()
			return m, textinput.Blink
		case "e":
			if m.Cursor >= 0 && m.Cursor < len(m.Habits) {
				m.Editing = true
				m.InputtingGoal = false
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
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Bold(true).Render("Are you sure you want to save changes? Previous records will be affected. (y/n)") + "\n\n"
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

		// Also display goal for visibility
		goal := fmt.Sprintf("(Goal: %d/wk)", habit.GoalTarget)

		s += style.Render(fmt.Sprintf("%s %s %s", cursor, name, goal)) + "\n"
	}

	if m.Adding || m.Editing {
		title := "New Habit"
		if m.Editing {
			title = "Edit Habit"
		}

		label := "Name"
		if m.InputtingGoal {
			label = "Weekly Goal (1-7)"
		}

		s += "\n" + lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%s - %s:", title, label))
		s += "\n" + m.TextInput.View() + "\n"
		s += ui.HelpStyle.Render("(Enter to next/save, Esc to cancel)")
	} else {
		s += "\n" + ui.HelpStyle.Render("(n: new, e: edit, d: archive, u: un-archive, j/k: navigate)")
	}

	return s
}
