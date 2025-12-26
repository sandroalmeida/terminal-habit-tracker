CREATE DATABASE IF NOT EXISTS habit_tracker;
USE habit_tracker;

CREATE TABLE IF NOT EXISTS habits (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    position INT DEFAULT 0,
    goal_target INT DEFAULT 0,
    is_archived BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS habit_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    habit_id INT NOT NULL,
    log_date DATE NOT NULL,
    UNIQUE KEY unique_log (habit_id, log_date),
    FOREIGN KEY (habit_id) REFERENCES habits(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS motivation_levels (
    id INT AUTO_INCREMENT PRIMARY KEY,
    threshold INT NOT NULL,
    emoji VARCHAR(10) NOT NULL
);
