package monthly

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
	Stats        map[int]service.HabitStats
	CurrentDate  time.Time // Always 1st of the month
	CursorY      int       // Habit index
	Err          error
}

func NewModel(habitRepo *repository.HabitRepository, logRepo *repository.LogRepository, statsService *service.StatsService) Model {
	now := time.Now()
	// Start at beginning of current month
	currentDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)

	return Model{
		HabitRepo:    habitRepo,
		LogRepo:      logRepo,
		StatsService: statsService,
		CurrentDate:  currentDate,
		Stats:        make(map[int]service.HabitStats),
	}
}

func (m Model) Init() tea.Cmd {
	return m.LoadHabits
}

func (m Model) LoadHabits() tea.Msg {
	habits, err := m.HabitRepo.List(false)
	if err != nil {
		return err
	}
	return habits
}

// Bulk fetch logs for the month
func (m Model) LoadMonthLogs() tea.Msg {
	nextMonth := m.CurrentDate.AddDate(0, 1, 0)
	logs, err := m.LogRepo.GetLogsBetween(m.CurrentDate, nextMonth)
	if err != nil {
		return err
	}
	return logs
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case []models.Habit:
		m.Habits = msg
		// Adjust cursor
		if m.CursorY >= len(m.Habits) && len(m.Habits) > 0 {
			m.CursorY = len(m.Habits) - 1
		}
		// Load logs for the month once habits are ready
		return m, m.LoadMonthLogs

	case map[int]map[string]bool:
		// Calculated stats
		stats := make(map[int]service.HabitStats)
		for _, habit := range m.Habits {
			habitLogs := msg[habit.ID]
			// Handle nil map
			if habitLogs == nil {
				habitLogs = make(map[string]bool)
			}
			s := m.StatsService.CalculateMonthlyStats(habitLogs, m.CurrentDate.Year(), m.CurrentDate.Month(), habit.GoalTarget)
			stats[habit.ID] = s
		}
		m.Stats = stats

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
			// Previous Month
			m.CurrentDate = m.CurrentDate.AddDate(0, -1, 0)
			return m, m.LoadMonthLogs
		case "right", "l":
			// Next Month
			m.CurrentDate = m.CurrentDate.AddDate(0, 1, 0)
			return m, m.LoadMonthLogs
		}
	}

	return m, nil
}

func (m Model) View() string {
	s := ui.TitleStyle.Render("Monthly Tracker") + "\n\n"

	if m.Err != nil {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("Error: %v", m.Err)) + "\n\n"
	}

	// Month Header
	monthStr := m.CurrentDate.Format("January 2006")
	s += lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render(monthStr) + "\n\n"

	// Table Header
	header := lipgloss.NewStyle().Width(26).Render("Habit")
	header += lipgloss.NewStyle().Width(8).Align(lipgloss.Center).Render("Total")
	header += lipgloss.NewStyle().Width(8).Align(lipgloss.Center).Render("Current")
	header += lipgloss.NewStyle().Width(8).Align(lipgloss.Center).Render("Longest")
	header += lipgloss.NewStyle().Width(8).Align(lipgloss.Center).Render("Goal")
	header += lipgloss.NewStyle().Width(8).Align(lipgloss.Center).Render("Progress")
	s += header + "\n"

	for i, habit := range m.Habits {
		stats := m.Stats[habit.ID]

		nameStyle := lipgloss.NewStyle().Width(26).Align(lipgloss.Right).PaddingRight(1)
		if m.CursorY == i {
			nameStyle = nameStyle.Foreground(lipgloss.Color("205")).Bold(true)
		}
		row := nameStyle.Render(habit.Name)

		// Reduce length if too long
		if len(habit.Name) > 25 {
			row = nameStyle.Render(habit.Name[:25])
		}

		row += lipgloss.NewStyle().Width(8).Align(lipgloss.Center).Render(fmt.Sprintf("%d", stats.Total))
		row += lipgloss.NewStyle().Width(8).Align(lipgloss.Center).Render(fmt.Sprintf("%d", stats.CurrentStreak))
		row += lipgloss.NewStyle().Width(8).Align(lipgloss.Center).Render(fmt.Sprintf("%d", stats.LongestStreak))

		// Calculate Monthly Goal for Display
		// (DaysInMonth * WeeklyGoal) / 7
		// Recalculate or store? We have it in stats, but stats struct doesn't have "Goal" field, only Progress.
		// Wait, Stats struct doesn't have Goal value. It has Progress.
		// Service computes progress base on goal.
		// I should probably duplicate the goal calc here for display or add it to HabitStats struct.
		// Adding to HabitStats struct breaks separation slightly (model vs view), but is convenient.
		// Or just recalc here.
		nextMonth := m.CurrentDate.AddDate(0, 1, 0)
		daysInMonth := nextMonth.AddDate(0, 0, -1).Day()
		monthlyGoal := 0
		if habit.GoalTarget > 0 {
			monthlyGoal = (daysInMonth * habit.GoalTarget) / 7
		}

		goalStr := "∞"
		if monthlyGoal > 0 {
			goalStr = fmt.Sprintf("%d", monthlyGoal)
		}
		row += lipgloss.NewStyle().Width(8).Align(lipgloss.Center).Render(goalStr)

		progDisplay := fmt.Sprintf("%d%%", stats.Progress)
		row += lipgloss.NewStyle().Width(8).Align(lipgloss.Center).Render(progDisplay)

		s += row + "\n"
	}

	s += "\n" + ui.HelpStyle.Render("(Arrow Keys: Navigate, Tab: Switch View)")
	return s
}
