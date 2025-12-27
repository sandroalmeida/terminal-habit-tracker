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
	streak := s.CalculateStreak(logs, time.Now())

	// 4. Calculate Progress
	progress := 0
	if goalTarget > 0 {
		progress = (total * 100) / goalTarget
	}

	return HabitStats{
		Total:         total,
		CurrentStreak: streak,
		Progress:      progress,
	}, nil
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
