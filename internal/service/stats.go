package service

import (
	"habit-tracker/internal/repository"
	"time"
)

type HabitStats struct {
	Total         int
	CurrentStreak int
	LongestStreak int // Optional for now
	Progress      int // Percent
}

type StatsService struct {
	LogRepo *repository.LogRepository
}

func NewStatsService(logRepo *repository.LogRepository) *StatsService {
	return &StatsService{LogRepo: logRepo}
}

func (s *StatsService) GetStats(habitID int, goalTarget int) (HabitStats, error) {
	// 1. Get Total Count
	total, err := s.LogRepo.GetTotalCount(habitID)
	if err != nil {
		return HabitStats{}, err
	}

	// 2. Get All Logs for Streak Calculation
	logs, err := s.LogRepo.GetAllLogs(habitID)
	if err != nil {
		return HabitStats{}, err
	}

	// 3. Calculate Streak
	currentStreak := s.CalculateStreak(logs, time.Now())
	longestStreak := s.CalculateLongestStreak(logs)

	// 4. Calculate Progress
	progress := 0
	if goalTarget > 0 {
		progress = (total * 100) / goalTarget
	}

	return HabitStats{
		Total:         total,
		CurrentStreak: currentStreak,
		LongestStreak: longestStreak,
		Progress:      progress,
	}, nil
}

// CalculateLongestStreak calculates the longest streak of consecutive days
func (s *StatsService) CalculateLongestStreak(logs []time.Time) int {
	if len(logs) == 0 {
		return 0
	}

	maxStreak := 0
	current := 0

	// Logs are ordered by date from repo
	for i, date := range logs {
		if i == 0 {
			current = 1
			maxStreak = 1
			continue
		}

		prevDate := logs[i-1]
		diff := date.Sub(prevDate).Hours() / 24

		// Check if consecutive (approx 1 day diff)
		// Allowing some slack for DST or slight time variances if they existed,
		// but repo uses "2006-01-02", so times should be aligned or we just check day diff.
		// Since we scan time.Time from DB which usually sets time to 00:00:00 for DATE columns.

		if int(diff+0.5) == 1 {
			current++
		} else if int(diff+0.5) > 1 {
			// Gap found, reset
			current = 1
		}
		// If diff == 0 (duplicate day), do not increment but do not reset.
		// (Though ToggleLog logic usually prevents duplicates)

		if current > maxStreak {
			maxStreak = current
		}
	}
	return maxStreak
}

// CalculateStreak calculates the current streak including connected future days
func (s *StatsService) CalculateStreak(logs []time.Time, refDate time.Time) int {
	if len(logs) == 0 {
		return 0
	}

	logMap := make(map[string]bool)
	for _, date := range logs {
		logMap[date.Format("2006-01-02")] = true
	}

	streak := 0
	checkDate := refDate

	// 1. Check if today is logged
	if logMap[checkDate.Format("2006-01-02")] {
		streak++

		// Count forward
		forwardDate := checkDate.AddDate(0, 0, 1)
		for {
			if logMap[forwardDate.Format("2006-01-02")] {
				streak++
				forwardDate = forwardDate.AddDate(0, 0, 1)
			} else {
				break
			}
		}

		// Count backward
		backwardDate := checkDate.AddDate(0, 0, -1)
		for {
			if logMap[backwardDate.Format("2006-01-02")] {
				streak++
				backwardDate = backwardDate.AddDate(0, 0, -1)
			} else {
				break
			}
		}
	} else {
		// 2. Check if yesterday is logged
		checkDate = checkDate.AddDate(0, 0, -1)
		if logMap[checkDate.Format("2006-01-02")] {
			streak++

			// Count backward from yesterday
			backwardDate := checkDate.AddDate(0, 0, -1)
			for {
				if logMap[backwardDate.Format("2006-01-02")] {
					streak++
					backwardDate = backwardDate.AddDate(0, 0, -1)
				} else {
					break
				}
			}
		}
	}

	return streak
}

func (s *StatsService) GetDailyEmoji(totalItems, completedItems int) string {
	if totalItems == 0 {
		return ""
	}

	percentage := (completedItems * 100) / totalItems

	switch {
	case percentage >= 90:
		return "🏆"
	case percentage >= 70:
		return "🥇"
	case percentage >= 50:
		return "🙂"
	default:
		return "  " // Empty space to keep alignment
	}
}

func (s *StatsService) CalculateWeeklyStats(habitLogs map[string]bool, startDate time.Time, goal int) HabitStats {
	total := 0
	longestStreak := 0
	currentStreak := 0

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// Determine "End of Calculation" (Latest day to consider)
	weekEnd := startDate.AddDate(0, 0, 6)
	calcEnd := weekEnd
	if calcEnd.After(today) {
		calcEnd = today
	}

	// 1. Scan for Total and Longest Logged (strictly within the week)
	activeRun := 0
	for i := 0; i < 7; i++ {
		d := startDate.AddDate(0, 0, i)
		dStr := d.Format("2006-01-02")

		// Only count if within valid range (though map lookup handles it, logical check helps)
		if habitLogs[dStr] {
			total++
			activeRun++
		} else {
			if activeRun > longestStreak {
				longestStreak = activeRun
			}
			activeRun = 0
		}
	}
	// Check verify last run
	if activeRun > longestStreak {
		longestStreak = activeRun
	}

	// 2. Calculate Current Streak (Standard habit logic: "Streak alive?")
	// Scoped to Week: Count backwards from min(Today, Sunday).
	// ... (same logic as before)
	// Ensure we don't look before startDate
	if !calcEnd.Before(startDate) {
		checkDate := calcEnd

		// Grace period for Today
		if calcEnd.Equal(today) && !habitLogs[calcEnd.Format("2006-01-02")] {
			// Today is empty. Check yesterday.
			yesterday := calcEnd.AddDate(0, 0, -1)
			if !yesterday.Before(startDate) && habitLogs[yesterday.Format("2006-01-02")] {
				checkDate = yesterday
			} else {
				// Streak broken or never started
				goto Finish
			}
		} else if !habitLogs[checkDate.Format("2006-01-02")] {
			// Not today, and empty at end of period. Streak 0.
			goto Finish
		}

		// Count backwards loop
		for {
			if checkDate.Before(startDate) {
				break
			}
			if habitLogs[checkDate.Format("2006-01-02")] {
				currentStreak++
			} else {
				break
			}
			checkDate = checkDate.AddDate(0, 0, -1)
		}
	}

Finish:
	progress := 0
	if goal > 0 {
		progress = (total * 100) / goal
	}

	return HabitStats{
		Total:         total,
		CurrentStreak: currentStreak,
		LongestStreak: longestStreak,
		Progress:      progress,
	}
}

func (s *StatsService) CalculateMonthlyStats(habitLogs map[string]bool, year int, month time.Month, weeklyGoal int) HabitStats {
	total := 0
	longestStreak := 0
	currentStreak := 0

	// Determine start and end of month
	startOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	nextMonth := startOfMonth.AddDate(0, 1, 0)
	endOfMonth := nextMonth.AddDate(0, 0, -1)
	daysInMonth := endOfMonth.Day()

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	// Effective end for calculations (don't count future days for streak purposes if current month)
	calcEnd := endOfMonth
	if calcEnd.After(today) {
		calcEnd = today
	}

	// 1. Total & Longest
	activeRun := 0
	for d := 1; d <= daysInMonth; d++ {
		date := time.Date(year, month, d, 0, 0, 0, 0, time.Local)
		dStr := date.Format("2006-01-02")

		if habitLogs[dStr] {
			total++
			activeRun++
		} else {
			if activeRun > longestStreak {
				longestStreak = activeRun
			}
			activeRun = 0
		}
	}
	if activeRun > longestStreak {
		longestStreak = activeRun
	}

	// 2. Current Streak (Monthly Context)
	// If month is past: check from end of month backwards.
	// If month is current: check from today backwards.
	// If month is future: 0.

	if startOfMonth.After(today) {
		currentStreak = 0
	} else {
		checkDate := calcEnd
		// Grace period logic similar to weekly?
		// If checkDate is today, we allow it to be unchecked IF yesterday was checked.
		// If checkDate is end of past month, we strictly expect it to be checked to start a streak from there?
		// Requirement: "current will be the current streak for the month."
		// Let's stick to standard streak logic: count backwards from 'calcEnd'.

		// If calcEnd is Today, standard logic applies (today or yesterday).
		// If calcEnd is EndOfMonth (past), we probably just want "active streak ending on this day".

		streakStartFn := func(start time.Time) time.Time {
			// If start is unchecked, maybe yesterday?
			if habitLogs[start.Format("2006-01-02")] {
				return start
			}
			// Only allow lookback if start is Today
			if start.Equal(today) {
				prev := start.AddDate(0, 0, -1)
				// Ensure prev is still in this month
				if prev.Month() == month && habitLogs[prev.Format("2006-01-02")] {
					return prev
				}
			}
			return time.Time{} // Invalid
		}

		startCountingDate := streakStartFn(checkDate)
		if !startCountingDate.IsZero() {
			// Count backwards
			curr := startCountingDate
			for {
				if curr.Before(startOfMonth) {
					break
				}
				if habitLogs[curr.Format("2006-01-02")] {
					currentStreak++
				} else {
					break
				}
				curr = curr.AddDate(0, 0, -1)
			}
		}
	}

	// 3. Goal & Progress
	// Goal = (DaysInMonth * WeeklyGoal) / 7
	// Note: Weekly goal is usually small (e.g. 5).
	monthlyGoal := 0
	if weeklyGoal > 0 {
		monthlyGoal = (daysInMonth * weeklyGoal) / 7
	}

	progress := 0
	if monthlyGoal > 0 {
		progress = (total * 100) / monthlyGoal
	}

	return HabitStats{
		Total:         total,
		CurrentStreak: currentStreak,
		LongestStreak: longestStreak,
		Progress:      progress,
	}
}
