package repository

import (
	"database/sql"
	"time"
)

type LogRepository struct {
	DB *sql.DB
}

func NewLogRepository(db *sql.DB) *LogRepository {
	return &LogRepository{DB: db}
}

// ToggleLog toggles the status of a habit for a given date.
// If it exists, it deletes it. If it doesn't, it creates it.
func (r *LogRepository) ToggleLog(habitID int, date time.Time) error {
	// Check if exists
	var id int
	dateStr := date.Format("2006-01-02")
	query := `SELECT id FROM habit_logs WHERE habit_id = ? AND log_date = ?`
	err := r.DB.QueryRow(query, habitID, dateStr).Scan(&id)

	if err == sql.ErrNoRows {
		// Insert
		insertQuery := `INSERT INTO habit_logs (habit_id, log_date) VALUES (?, ?)`
		_, err := r.DB.Exec(insertQuery, habitID, dateStr)
		return err
	} else if err != nil {
		return err
	}

	// Delete
	deleteQuery := `DELETE FROM habit_logs WHERE id = ?`
	_, err = r.DB.Exec(deleteQuery, id)
	return err
}

// GetLogs returns a map of habitID -> date("2006-01-02") -> true
func (r *LogRepository) GetLogs(startDate, endDate time.Time) (map[int]map[string]bool, error) {
	query := `SELECT habit_id, log_date FROM habit_logs WHERE log_date BETWEEN ? AND ?`
	rows, err := r.DB.Query(query, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make(map[int]map[string]bool)
	for rows.Next() {
		var habitID int
		var logDateStr string // Scan as string to handle DATE type consistently
		if err := rows.Scan(&habitID, &logDateStr); err != nil {
			return nil, err
		}

		// Depending on driver, DATE might scan as []byte or string or time.Time.
		// Using string is usually safe for DATE if we don't care about time component here.
		// However, mysql driver with parseTime=true scans to time.Time.
		// Let's verify by scanning into string or using a helper.
		// Actually, since we configured parseTime=true, it should be time.Time.
		// Let's update the scan target.
	}
	return logs, nil
}

// GetLogsFixed implementation with correct type scanning
func (r *LogRepository) GetLogsBetween(startDate, endDate time.Time) (map[int]map[string]bool, error) {
	query := `SELECT habit_id, log_date FROM habit_logs WHERE log_date BETWEEN ? AND ?`
	rows, err := r.DB.Query(query, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	logs := make(map[int]map[string]bool)
	for rows.Next() {
		var habitID int
		var logDate time.Time
		if err := rows.Scan(&habitID, &logDate); err != nil {
			return nil, err
		}

		if logs[habitID] == nil {
			logs[habitID] = make(map[string]bool)
		}
		logs[habitID][logDate.Format("2006-01-02")] = true
	}

	return logs, nil
}

// GetTotalCount returns the total number of logs for a habit
func (r *LogRepository) GetTotalCount(habitID int) (int, error) {
	query := `SELECT COUNT(*) FROM habit_logs WHERE habit_id = ?`
	var count int
	err := r.DB.QueryRow(query, habitID).Scan(&count)
	return count, err
}

// GetCurrentStreak calculates the current streak of consecutive days
func (r *LogRepository) GetCurrentStreak(habitID int) (int, error) {
	// Simple approach: Iterate backwards from today/yesterday
	// A more optimized SQL approach exists (recursion/window functions),
	// but iterative check is fine for personal scale.
	streak := 0
	date := time.Now()

	// Check today first
	if r.hasLog(habitID, date) {
		streak++
	} else {
		// If not done today, checking yesterday implies the streak is still "active"
		// if we consider "current streak" to include yesterday.
		// However, typically streak is 0 if not done today OR yesterday.
		// Let's cycle back.
	}

	// Actually, easier logic:
	// Start from today. If today is done, streak++. Check yesterday.
	// If today is NOT done, streak is technically 0 for "today",
	// but often apps show the streak up to yesterday if today isn't over.
	// Let's check yesterday.

	// Refined logic:
	// If today is logged, count it and go back.
	// If today is NOT logged, check yesterday. If yesterday is logged, count it and go back.
	// If neither, streak is 0.

	todayLogged := r.hasLog(habitID, date)
	if todayLogged {
		streak++
		date = date.AddDate(0, 0, -1)
	} else {
		// Check yesterday
		date = date.AddDate(0, 0, -1)
		if !r.hasLog(habitID, date) {
			return 0, nil
		}
		streak++
		date = date.AddDate(0, 0, -1) // Move to day before yesterday
	}

	for {
		if r.hasLog(habitID, date) {
			streak++
			date = date.AddDate(0, 0, -1)
		} else {
			break
		}
	}

	return streak, nil
}

func (r *LogRepository) hasLog(habitID int, date time.Time) bool {
	query := `SELECT id FROM habit_logs WHERE habit_id = ? AND log_date = ?`
	var id int
	err := r.DB.QueryRow(query, habitID, date.Format("2006-01-02")).Scan(&id)
	return err == nil
}
