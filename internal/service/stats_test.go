package service

import (
	"testing"
	"time"
)

func TestGetDailyEmoji(t *testing.T) {
	tests := []struct {
		name      string
		total     int
		completed int
		want      string
	}{
		{"No items", 0, 0, ""},
		{"0% progress", 10, 0, "  "},
		{"40% progress", 10, 4, "  "},
		{"50% progress", 10, 5, "🙂"},
		{"60% progress", 10, 6, "🙂"},
		{"70% progress", 10, 7, "🥇"},
		{"80% progress", 10, 8, "🥇"},
		{"90% progress", 10, 9, "🏆"},
		{"100% progress", 10, 10, "🏆"},
		{"Over 100% (should probably cap to 100 theoretically but logic handles it)", 10, 11, "🏆"},
	}

	s := &StatsService{} // Method doesn't use repo, so nil is fine for this test

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.GetDailyEmoji(tt.total, tt.completed); got != tt.want {
				t.Errorf("GetDailyEmoji(%d, %d) = %v, want %v", tt.total, tt.completed, got, tt.want)
			}
		})
	}
}

func TestCalculateStreak(t *testing.T) {
	// Setup dates
	now := time.Now()
	// today := now
	yesterday := now.AddDate(0, 0, -1)
	dayBefore := now.AddDate(0, 0, -2)
	tomorrow := now.AddDate(0, 0, 1) // Sunday if today checked is Sat
	nextDay := now.AddDate(0, 0, 2)

	tests := []struct {
		name string
		logs []time.Time
		want int
	}{
		{
			name: "No logs",
			logs: []time.Time{},
			want: 0,
		},
		{
			name: "Today only",
			logs: []time.Time{now},
			want: 1,
		},
		{
			name: "Yesterday only",
			logs: []time.Time{yesterday},
			want: 1,
		},
		{
			name: "Day before yesterday only (break)",
			logs: []time.Time{dayBefore},
			want: 0,
		},
		{
			name: "Streak 2 (Today + Yesterday)",
			logs: []time.Time{now, yesterday},
			want: 2,
		},
		{
			name: "Streak 2 (Today + Tomorrow) - Forward counting case",
			logs: []time.Time{now, tomorrow},
			want: 2,
		},
		{
			name: "Streak 3 (Yesterday + Today + Tomorrow)",
			logs: []time.Time{yesterday, now, tomorrow},
			want: 3,
		},
		{
			name: "Break in future (Today logged, tomorrow missed, next day logged)",
			logs: []time.Time{now, nextDay},
			want: 1,
		},
		{
			name: "Streak 2 (Yesterday + DayBefore)",
			logs: []time.Time{yesterday, dayBefore},
			want: 2,
		},
	}

	s := &StatsService{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.CalculateStreak(tt.logs, now); got != tt.want {
				t.Errorf("CalculateStreak() = %v, want %v", got, tt.want)
			}
		})
	}
}
