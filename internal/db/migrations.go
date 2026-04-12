package db

import "database/sql"

func RunMigrations(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id               INTEGER  PRIMARY KEY AUTOINCREMENT,
		username         TEXT     NOT NULL UNIQUE,
		telegram_id      INTEGER  NOT NULL DEFAULT 0,
		telegram_chat_id INTEGER  NOT NULL DEFAULT 0,
		created_at       DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS projects (
		id          INTEGER  PRIMARY KEY AUTOINCREMENT,
		user_id     INTEGER  NOT NULL REFERENCES users(id),
		name        TEXT     NOT NULL,
		description TEXT     NOT NULL DEFAULT '',
		status      TEXT     NOT NULL DEFAULT 'ACTIVE',
		due_date    DATETIME,
		created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS tags (
		id         INTEGER  PRIMARY KEY AUTOINCREMENT,
		user_id    INTEGER  NOT NULL REFERENCES users(id),
		name       TEXT     NOT NULL,
		color      TEXT     NOT NULL DEFAULT '#ffffff',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id         INTEGER  PRIMARY KEY AUTOINCREMENT,
		user_id    INTEGER  NOT NULL REFERENCES users(id),
		project_id INTEGER  REFERENCES projects(id),
		tag_id     INTEGER  REFERENCES tags(id),
		content    TEXT     NOT NULL,
		status     TEXT     NOT NULL DEFAULT 'TODO',
		priority   TEXT     NOT NULL DEFAULT 'MEDIUM',
		due_date   DATETIME,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS notes (
		id         INTEGER  PRIMARY KEY AUTOINCREMENT,
		user_id    INTEGER  NOT NULL REFERENCES users(id),
		project_id INTEGER  REFERENCES projects(id),
		tag_id     INTEGER  REFERENCES tags(id),
		title      TEXT     NOT NULL,
		content    TEXT     NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS reminders (
		id            INTEGER  PRIMARY KEY AUTOINCREMENT,
		user_id       INTEGER  NOT NULL REFERENCES users(id),
		project_id    INTEGER  REFERENCES projects(id),
		tag_id        INTEGER  REFERENCES tags(id),
		content       TEXT     NOT NULL,
		reminder_time DATETIME NOT NULL,
		status        TEXT     NOT NULL DEFAULT 'PENDING',
		created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS pomodoro_sessions (
		id             INTEGER  PRIMARY KEY AUTOINCREMENT,
		user_id        INTEGER  NOT NULL REFERENCES users(id),
		project_id     INTEGER  REFERENCES projects(id),
		start_time     DATETIME,
		work_duration  INTEGER  NOT NULL DEFAULT 25,
		finish_time    DATETIME,
		remaining_time INTEGER  NOT NULL DEFAULT 0,
		state          TEXT     NOT NULL DEFAULT 'IDLE',
		created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(schema)
	return err
}
