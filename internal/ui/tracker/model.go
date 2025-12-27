package tracker

import (
	"fmt"
	"habit-tracker/internal/models"
	"habit-tracker/internal/repository"
	"habit-tracker/internal/ui"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	HabitRepo *repository.HabitRepository
	LogRepo   *repository.LogRepository
	Habits    []models.Habit
	Logs      map[int]map[string]bool
	CursorX   int // Day index (0-6)
	CursorY   int // Habit index
	StartDate time.Time
	Err       error
}

func NewModel(habitRepo *repository.HabitRepository, logRepo *repository.LogRepository) Model {
	// Find start of current week (Monday)
	now := time.Now()
	offset := int(now.Weekday()) - 1
	if offset < 0 {
		offset = 6 // Sunday
	}
	startDate := now.AddDate(0, 0, -offset)

	return Model{
		HabitRepo: habitRepo,
		LogRepo:   logRepo,
		StartDate: startDate,
		Logs:      make(map[int]map[string]bool),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.LoadHabits, m.LoadLogs)
}

func (m Model) LoadHabits() tea.Msg {
	habits, err := m.HabitRepo.List()
	if err != nil {
		return err
	}
	return habits
}

func (m Model) LoadLogs() tea.Msg {
	endDate := m.StartDate.AddDate(0, 0, 7)
	logs, err := m.LogRepo.GetLogsBetween(m.StartDate, endDate)
	if err != nil {
		return err
	}
	return logs
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case []models.Habit:
		m.Habits = msg
		// Adjust cursor if habits list changed size
		if m.CursorY >= len(m.Habits) && len(m.Habits) > 0 {
			m.CursorY = len(m.Habits) - 1
		}

	case map[int]map[string]bool:
		m.Logs = msg

	case error:
		m.Err = msg

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.CursorY > 0 {
				m.CursorY--
			}
		case "down", "j":
			if m.CursorY < len(m.Habits)-1 {
				m.CursorY++
			}
		case "left", "h":
			if m.CursorX > 0 {
				m.CursorX--
			} else {
				// Move to previous week
				m.StartDate = m.StartDate.AddDate(0, 0, -7)
				m.CursorX = 6
				return m, m.LoadLogs
			}
		case "right", "l":
			if m.CursorX < 6 {
				m.CursorX++
			} else {
				// Move to next week
				m.StartDate = m.StartDate.AddDate(0, 0, 7)
				m.CursorX = 0
				return m, m.LoadLogs
			}
		case " ":
			if len(m.Habits) > 0 {
				habitID := m.Habits[m.CursorY].ID
				date := m.StartDate.AddDate(0, 0, m.CursorX)

				if err := m.LogRepo.ToggleLog(habitID, date); err != nil {
					m.Err = err
				} else {
					return m, m.LoadLogs
				}
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	s := ui.TitleStyle.Render("Weekly Tracker") + "\n\n"

	if m.Err != nil {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("Error: %v", m.Err)) + "\n\n"
	}

	// Helper to format date header
	header := "                     " // Padding for Habit Name
	for i := 0; i < 7; i++ {
		date := m.StartDate.AddDate(0, 0, i)
		dayStr := date.Format("Mon 02")
		style := lipgloss.NewStyle().Width(8).Align(lipgloss.Center)
		if i == m.CursorX {
			style = style.Bold(true).Foreground(lipgloss.Color("205")).Underline(true)
		}
		header += style.Render(dayStr)
	}
	s += header + "\n"

	for i, habit := range m.Habits {
		// Habit Name
		nameStyle := lipgloss.NewStyle().Width(20).Align(lipgloss.Right).PaddingRight(1)
		if m.CursorY == i {
			nameStyle = nameStyle.Foreground(lipgloss.Color("205")).Bold(true)
		}
		row := nameStyle.Render(habit.Name)

		// Days
		for d := 0; d < 7; d++ {
			date := m.StartDate.AddDate(0, 0, d)
			dateStr := date.Format("2006-01-02")

			isChecked := false
			if m.Logs[habit.ID] != nil && m.Logs[habit.ID][dateStr] {
				isChecked = true
			}

			box := "NO"
			if isChecked {
				box = "YES"
			}

			// Emoji replacement
			if isChecked {
				box = "✅"
			} else {
				box = "⬜"
			}

			cellStyle := lipgloss.NewStyle().Width(8).Align(lipgloss.Center)
			if m.CursorY == i && m.CursorX == d {
				cellStyle = cellStyle.Background(lipgloss.Color("62")).Foreground(lipgloss.Color("255"))
			}

			row += cellStyle.Render(box)
		}

		s += row + "\n"
	}

	s += "\n" + ui.HelpStyle.Render("(Arrow Keys: Navigate, Space: Toggle, Tab: Switch View)")
	return s
}
