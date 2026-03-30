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

func (r *HabitRepository) NormalizePositions() error {
	rows, err := r.DB.Query(`SELECT id FROM habits ORDER BY position ASC, id ASC`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}

	for i, id := range ids {
		_, err := r.DB.Exec(`UPDATE habits SET position = ? WHERE id = ?`, i+1, id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *HabitRepository) MaxPosition() (int, error) {
	var maxPos sql.NullInt64
	err := r.DB.QueryRow(`SELECT MAX(position) FROM habits`).Scan(&maxPos)
	if err != nil {
		return 0, err
	}
	if !maxPos.Valid {
		return 0, nil
	}
	return int(maxPos.Int64), nil
}

func (r *HabitRepository) SwapPositions(id1, pos1, id2, pos2 int) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE habits SET position = ? WHERE id = ?`, pos2, id1)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE habits SET position = ? WHERE id = ?`, pos1, id2)
	if err != nil {
		return err
	}

	return tx.Commit()
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
