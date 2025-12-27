package service

import (
	"testing"
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
