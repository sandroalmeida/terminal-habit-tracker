package service

import (
	"habit-tracker/internal/repository"
)

type HabitStats struct {
	Total         int
	CurrentStreak int
	LongestStreak int // Optional for now
	Progress      int // Percent
	Emoji         string
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

	// 2. Get Current Streak
	streak, err := s.LogRepo.GetCurrentStreak(habitID)
	if err != nil {
		return HabitStats{}, err
	}

	// 3. Calculate Progress
	progress := 0
	if goalTarget > 0 {
		progress = (total * 100) / goalTarget
		if progress > 100 {
			progress = 100
		}
	}

	// 4. Get Emoji
	emoji := ""
	switch {
	case total >= 20: // Example thresholds
		emoji = "🏆"
	case total >= 10:
		emoji = "🥇"
	case total >= 5:
		emoji = "😃"
	}

	return HabitStats{
		Total:         total,
		CurrentStreak: streak,
		Progress:      progress,
		Emoji:         emoji,
	}, nil
}
