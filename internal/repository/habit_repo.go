package repository

import (
	"database/sql"
	"habit-tracker/internal/models"
)

type HabitRepository struct {
	DB *sql.DB
}

func NewHabitRepository(db *sql.DB) *HabitRepository {
	return &HabitRepository{DB: db}
}

func (r *HabitRepository) Create(habit *models.Habit) error {
	query := `INSERT INTO habits (name, description, position, goal_target) VALUES (?, ?, ?, ?)`
	result, err := r.DB.Exec(query, habit.Name, habit.Description, habit.Position, habit.GoalTarget)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	habit.ID = int(id)
	return nil
}

func (r *HabitRepository) Update(habit *models.Habit) error {
	query := `UPDATE habits SET name = ?, description = ?, position = ?, goal_target = ?, is_archived = ? WHERE id = ?`
	_, err := r.DB.Exec(query, habit.Name, habit.Description, habit.Position, habit.GoalTarget, habit.IsArchived, habit.ID)
	return err
}

func (r *HabitRepository) List(includeArchived bool) ([]models.Habit, error) {
	query := `SELECT id, name, description, position, goal_target, is_archived, created_at FROM habits`
	if !includeArchived {
		query += ` WHERE is_archived = FALSE`
	}
	query += ` ORDER BY position ASC, id ASC`

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var habits []models.Habit
	for rows.Next() {
		var h models.Habit
		if err := rows.Scan(&h.ID, &h.Name, &h.Description, &h.Position, &h.GoalTarget, &h.IsArchived, &h.CreatedAt); err != nil {
			return nil, err
		}
		habits = append(habits, h)
	}

	return habits, nil
}
