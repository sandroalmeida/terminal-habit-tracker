package tracker

import (
	"fmt"
	"habit-tracker/internal/models"
	"habit-tracker/internal/repository"
	"habit-tracker/internal/service"
	"habit-tracker/internal/ui"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	HabitRepo    *repository.HabitRepository
	LogRepo      *repository.LogRepository
	StatsService *service.StatsService
	Habits       []models.Habit
	Logs         map[int]map[string]bool
	Stats        map[int]service.HabitStats
	CursorX      int // Day index (0-6)
	CursorY      int // Habit index
	StartDate    time.Time
	Err          error
}

func NewModel(habitRepo *repository.HabitRepository, logRepo *repository.LogRepository, statsService *service.StatsService) Model {
	// Find start of current week (Monday)
	now := time.Now()
	offset := int(now.Weekday()) - 1
	if offset < 0 {
		offset = 6 // Sunday
	}
	startDate := now.AddDate(0, 0, -offset)

	return Model{
		HabitRepo:    habitRepo,
		LogRepo:      logRepo,
		StatsService: statsService,
		StartDate:    startDate,
		Logs:         make(map[int]map[string]bool),
		Stats:        make(map[int]service.HabitStats),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.LoadHabits, m.LoadLogs)
}

func (m Model) LoadHabits() tea.Msg {
	habits, err := m.HabitRepo.List(false)
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

func (m Model) LoadStats() tea.Msg {
	stats := make(map[int]service.HabitStats)
	for _, habit := range m.Habits {
		s, err := m.StatsService.GetStats(habit.ID, habit.GoalTarget)
		if err == nil {
			stats[habit.ID] = s
		}
	}
	return stats
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case []models.Habit:
		m.Habits = msg
		// Adjust cursor if habits list changed size
		if m.CursorY >= len(m.Habits) && len(m.Habits) > 0 {
			m.CursorY = len(m.Habits) - 1
		}
		// Reload stats when habits load
		return m, m.LoadStats

	case map[int]map[string]bool:
		m.Logs = msg
		// Reload stats when logs change
		return m, m.LoadStats

	case map[int]service.HabitStats:
		m.Stats = msg

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
					return m, tea.Batch(m.LoadLogs, m.LoadStats) // Helper to reload both probably better, but this works
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
	header := lipgloss.NewStyle().Width(26).Render(" ") // Padding for Habit Name
	for i := 0; i < 7; i++ {
		date := m.StartDate.AddDate(0, 0, i)
		dayStr := date.Format("Mon 02")
		style := lipgloss.NewStyle().Width(8).Align(lipgloss.Center)
		if i == m.CursorX {
			style = style.Bold(true).Foreground(lipgloss.Color("205")).Underline(true)
		}
		header += style.Render(dayStr)
	}

	// Stats Headers
	header += lipgloss.NewStyle().Width(6).Align(lipgloss.Center).Render("Total") // Reduced width to fit
	header += lipgloss.NewStyle().Width(6).Align(lipgloss.Center).Render("Cur")   // Current Streak
	header += lipgloss.NewStyle().Width(6).Align(lipgloss.Center).Render("Long")  // Longest Streak
	header += lipgloss.NewStyle().Width(6).Align(lipgloss.Center).Render("Goal")
	header += lipgloss.NewStyle().Width(8).Align(lipgloss.Center).Render("Progress")

	s += header + "\n"

	for i, habit := range m.Habits {
		// Habit Name
		name := habit.Name
		if len(name) > 25 {
			name = name[:25]
		}
		nameStyle := lipgloss.NewStyle().Width(26).Align(lipgloss.Right).PaddingRight(1)
		if m.CursorY == i {
			nameStyle = nameStyle.Foreground(lipgloss.Color("205")).Bold(true)
		}
		row := nameStyle.Render(name)

		// Days
		for d := 0; d < 7; d++ {
			date := m.StartDate.AddDate(0, 0, d)
			dateStr := date.Format("2006-01-02")

			isChecked := false
			if m.Logs[habit.ID] != nil && m.Logs[habit.ID][dateStr] {
				isChecked = true
			}

			box := "⬜"
			if isChecked {
				box = "✅"
			}

			cellStyle := lipgloss.NewStyle().Width(8).Align(lipgloss.Center)
			if m.CursorY == i && m.CursorX == d {
				cellStyle = cellStyle.Background(lipgloss.Color("62")).Foreground(lipgloss.Color("255"))
			}

			row += cellStyle.Render(box)
		}

		// Stats
		stats := m.Stats[habit.ID]
		// Adjusted column widths to 6 to fit everything
		row += lipgloss.NewStyle().Width(6).Align(lipgloss.Center).Render(fmt.Sprintf("%d", stats.Total))
		row += lipgloss.NewStyle().Width(6).Align(lipgloss.Center).Render(fmt.Sprintf("%d", stats.CurrentStreak))
		row += lipgloss.NewStyle().Width(6).Align(lipgloss.Center).Render(fmt.Sprintf("%d", stats.LongestStreak))

		goalStr := "∞"
		if habit.GoalTarget > 0 {
			goalStr = fmt.Sprintf("%d", habit.GoalTarget)
		}
		row += lipgloss.NewStyle().Width(6).Align(lipgloss.Center).Render(goalStr)

		// Progress / Emoji
		progDisplay := fmt.Sprintf("%d%%", stats.Progress)
		row += lipgloss.NewStyle().Width(8).Align(lipgloss.Center).Render(progDisplay)

		s += row + "\n"
	}

	// Daily Emojis Footer
	emojiRow := lipgloss.NewStyle().Width(26).Align(lipgloss.Right).PaddingRight(1).Render(" ")
	for d := 0; d < 7; d++ {
		date := m.StartDate.AddDate(0, 0, d)
		dateStr := date.Format("2006-01-02")

		totalActive := 0
		completed := 0

		for _, habit := range m.Habits {
			// Skip archived habits if necessary, but list only returns active ones anyway
			totalActive++
			if m.Logs[habit.ID] != nil && m.Logs[habit.ID][dateStr] {
				completed++
			}
		}

		emoji := m.StatsService.GetDailyEmoji(totalActive, completed)
		emojiRow += lipgloss.NewStyle().Width(8).Align(lipgloss.Center).Render(emoji)
	}
	s += "\n" + emojiRow + "\n"

	s += "\n" + ui.HelpStyle.Render("(Arrow Keys: Navigate, Space: Toggle, Tab: Switch View)")
	return s
}
