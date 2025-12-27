package models

import "time"

type HabitLog struct {
	ID      int       `json:"id"`
	HabitID int       `json:"habit_id"`
	LogDate time.Time `json:"log_date"`
}
