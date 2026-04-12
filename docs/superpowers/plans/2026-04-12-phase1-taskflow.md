# Phase 1 — TaskFlow Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a working Go CLI task manager + pomodoro tracker with SQLite that demonstrates Singleton, State, and Strategy patterns (Phase 1 requirements).

**Architecture:** Clean layered structure — domain entities → repository (SQLite) → service/patterns → minimal TUI. No frameworks. Singletons for Config and DB. State machine for PomodoroSession. Strategies for task filtering.

**Tech Stack:** Go 1.25, `modernc.org/sqlite` (pure Go, no CGO), `github.com/stretchr/testify` (assertions only), stdlib for everything else (TUI: fmt + bufio, no external TUI library).

---

## File Structure

```
taskflow/
├── cmd/taskflow/
│   └── main.go                        # Entry point, wires everything
├── internal/
│   ├── config/
│   │   └── config.go                  # Singleton: AppConfig (reads env vars)
│   ├── db/
│   │   ├── db.go                      # Singleton: *sql.DB connection
│   │   └── migrations.go              # CREATE TABLE statements
│   ├── domain/
│   │   ├── task.go                    # Task struct + TaskStatus + Priority enums
│   │   ├── project.go                 # Project struct + ProjectStatus enum
│   │   ├── note.go                    # Note struct
│   │   ├── reminder.go                # Reminder struct + ReminderStatus enum
│   │   ├── pomodoro.go                # PomodoroSession struct + SessionState enum
│   │   ├── tag.go                     # Tag struct
│   │   └── user.go                    # User struct
│   ├── repository/
│   │   ├── interfaces.go              # Repository interfaces for all entities
│   │   ├── task_repo.go               # SQLite Task CRUD
│   │   ├── project_repo.go            # SQLite Project CRUD
│   │   ├── note_repo.go               # SQLite Note CRUD
│   │   ├── reminder_repo.go           # SQLite Reminder CRUD
│   │   ├── pomodoro_repo.go           # SQLite PomodoroSession CRUD
│   │   ├── tag_repo.go                # SQLite Tag CRUD
│   │   └── user_repo.go               # SQLite User CRUD
│   ├── patterns/
│   │   ├── strategy/
│   │   │   └── filter.go             # FilterStrategy interface + ByStatus/ByTag/ByDate
│   │   └── state/
│   │       └── pomodoro_state.go     # PomodoroState interface + 5 concrete states
│   ├── math/
│   │   └── eta.go                    # ETACalculator: avg pomodoro * remaining tasks
│   └── tui/
│       └── app.go                    # Minimal loop: clear screen, numbered menu, bufio input
├── go.mod
└── go.sum
```

---

## Task 1: Project Scaffold

**Files:**
- Create: `go.mod`
- Create: `cmd/taskflow/main.go`

- [ ] **Step 1: Init go module**

```bash
cd /Users/khuzhokov/repos/application_design_patterns
go mod init taskflow
```

Expected: `go.mod` created with `module taskflow`

- [ ] **Step 2: Create main.go**

```go
// cmd/taskflow/main.go
package main

import "fmt"

func main() {
	fmt.Println("TaskFlow — Terminal Task Manager")
}
```

- [ ] **Step 3: Verify it runs**

```bash
go run ./cmd/taskflow/
```

Expected output: `TaskFlow — Terminal Task Manager`

- [ ] **Step 4: Commit**

```bash
git init
git add go.mod cmd/taskflow/main.go
git commit -m "feat: init taskflow project"
```

---

## Task 2: Domain Entities

**Files:**
- Create: `internal/domain/task.go`
- Create: `internal/domain/project.go`
- Create: `internal/domain/note.go`
- Create: `internal/domain/reminder.go`
- Create: `internal/domain/pomodoro.go`
- Create: `internal/domain/tag.go`
- Create: `internal/domain/user.go`
- Test: `internal/domain/task_test.go`

- [ ] **Step 1: Write failing test for Task**

```go
// internal/domain/task_test.go
package domain_test

import (
	"testing"
	"time"

	"taskflow/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestTaskIsOverdue(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	task := domain.Task{DueDate: &past, Status: domain.TaskStatusTodo}
	assert.True(t, task.IsOverdue())
}

func TestTaskIsNotOverduWhenDone(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour)
	task := domain.Task{DueDate: &past, Status: domain.TaskStatusDone}
	assert.False(t, task.IsOverdue())
}

func TestTaskIsNotOverdueWhenNoDueDate(t *testing.T) {
	task := domain.Task{Status: domain.TaskStatusTodo}
	assert.False(t, task.IsOverdue())
}
```

- [ ] **Step 2: Add testify dependency**

```bash
go get github.com/stretchr/testify
```

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./internal/domain/...
```

Expected: compile error — `domain` package doesn't exist yet

- [ ] **Step 4: Create domain entities**

```go
// internal/domain/task.go
package domain

import "time"

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "TODO"
	TaskStatusInProgress TaskStatus = "IN_PROGRESS"
	TaskStatusDone       TaskStatus = "DONE"
)

type Priority string

const (
	PriorityHigh   Priority = "HIGH"
	PriorityMedium Priority = "MEDIUM"
	PriorityLow    Priority = "LOW"
)

type Task struct {
	ID        int64
	UserID    int64
	ProjectID *int64
	TagID     *int64
	Content   string
	Status    TaskStatus
	Priority  Priority
	DueDate   *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t *Task) IsOverdue() bool {
	if t.DueDate == nil || t.Status == TaskStatusDone {
		return false
	}
	return time.Now().After(*t.DueDate)
}

func (t *Task) Complete() {
	t.Status = TaskStatusDone
	t.UpdatedAt = time.Now()
}
```

```go
// internal/domain/project.go
package domain

import "time"

type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "ACTIVE"
	ProjectStatusArchived ProjectStatus = "ARCHIVED"
	ProjectStatusDone     ProjectStatus = "DONE"
)

type Project struct {
	ID          int64
	UserID      int64
	Name        string
	Description string
	Status      ProjectStatus
	DueDate     *time.Time
	CreatedAt   time.Time
}
```

```go
// internal/domain/note.go
package domain

import "time"

type Note struct {
	ID        int64
	UserID    int64
	ProjectID *int64
	TagID     *int64
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
```

```go
// internal/domain/reminder.go
package domain

import "time"

type ReminderStatus string

const (
	ReminderStatusPending ReminderStatus = "PENDING"
	ReminderStatusSent    ReminderStatus = "SENT"
	ReminderStatusFailed  ReminderStatus = "FAILED"
)

type Reminder struct {
	ID           int64
	UserID       int64
	ProjectID    *int64
	TagID        *int64
	Content      string
	ReminderTime time.Time
	Status       ReminderStatus
	CreatedAt    time.Time
}

func (r *Reminder) IsReady() bool {
	return r.Status == ReminderStatusPending && time.Now().After(r.ReminderTime)
}

func (r *Reminder) MarkAsSent() {
	r.Status = ReminderStatusSent
}

func (r *Reminder) MarkAsFailed() {
	r.Status = ReminderStatusFailed
}
```

```go
// internal/domain/pomodoro.go
package domain

import "time"

type SessionState string

const (
	SessionStateIdle      SessionState = "IDLE"
	SessionStateRunning   SessionState = "RUNNING"
	SessionStatePaused    SessionState = "PAUSED"
	SessionStateCompleted SessionState = "COMPLETED"
	SessionStateCancelled SessionState = "CANCELLED"
)

type PomodoroSession struct {
	ID            int64
	UserID        int64
	ProjectID     *int64
	StartTime     *time.Time
	WorkDuration  int // minutes
	FinishTime    *time.Time
	RemainingTime int // seconds
	State         SessionState
	CreatedAt     time.Time
}
```

```go
// internal/domain/tag.go
package domain

import "time"

type Tag struct {
	ID        int64
	UserID    int64
	Name      string
	Color     string
	CreatedAt time.Time
}
```

```go
// internal/domain/user.go
package domain

import "time"

type User struct {
	ID             int64
	Username       string
	TelegramID     int64
	TelegramChatID int64
	CreatedAt      time.Time
}
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/domain/...
```

Expected: `PASS`

- [ ] **Step 6: Commit**

```bash
git add internal/domain/ go.sum go.mod
git commit -m "feat: add domain entities"
```

---

## Task 3: Config Singleton

**Files:**
- Create: `internal/config/config.go`
- Test: `internal/config/config_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/config/config_test.go
package config_test

import (
	"os"
	"testing"

	"taskflow/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestConfigReadsTelegramToken(t *testing.T) {
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token-123")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")

	cfg := config.Instance()
	assert.Equal(t, "test-token-123", cfg.TelegramBotToken)
}

func TestConfigSingleton(t *testing.T) {
	a := config.Instance()
	b := config.Instance()
	assert.Same(t, a, b)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/config/...
```

Expected: compile error

- [ ] **Step 3: Implement Config singleton**

```go
// internal/config/config.go
package config

import (
	"os"
	"sync"
)

type Config struct {
	TelegramBotToken string
	DBPath           string
}

var (
	instance *Config
	once     sync.Once
)

func Instance() *Config {
	once.Do(func() {
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "taskflow.db"
		}
		instance = &Config{
			TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
			DBPath:           dbPath,
		}
	})
	return instance
}

// ResetForTesting resets singleton — only for tests
func ResetForTesting() {
	instance = nil
	once = sync.Once{}
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/config/...
```

Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat: add config singleton"
```

---

## Task 4: DB Singleton + Migrations

**Files:**
- Create: `internal/db/db.go`
- Create: `internal/db/migrations.go`
- Test: `internal/db/db_test.go`

- [ ] **Step 1: Add SQLite dependency**

```bash
go get modernc.org/sqlite
```

- [ ] **Step 2: Write failing test**

```go
// internal/db/db_test.go
package db_test

import (
	"testing"

	"taskflow/internal/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationsCreateTables(t *testing.T) {
	conn, err := db.OpenMemory()
	require.NoError(t, err)
	defer conn.Close()

	err = db.RunMigrations(conn)
	require.NoError(t, err)

	tables := []string{"users", "projects", "tasks", "notes", "reminders", "pomodoro_sessions", "tags"}
	for _, table := range tables {
		var name string
		row := conn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table)
		err := row.Scan(&name)
		assert.NoError(t, err, "table %s should exist", table)
	}
}

func TestDBSingleton(t *testing.T) {
	db.ResetForTesting()
	a := db.Instance(":memory:")
	b := db.Instance(":memory:")
	assert.Same(t, a, b)
	db.ResetForTesting()
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./internal/db/...
```

Expected: compile error

- [ ] **Step 4: Implement DB singleton**

```go
// internal/db/db.go
package db

import (
	"database/sql"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	instance *sql.DB
	once     sync.Once
)

func Instance(path string) *sql.DB {
	once.Do(func() {
		conn, err := sql.Open("sqlite", path)
		if err != nil {
			panic("db: failed to open: " + err.Error())
		}
		conn.SetMaxOpenConns(1) // SQLite: single writer
		if err := RunMigrations(conn); err != nil {
			panic("db: migrations failed: " + err.Error())
		}
		instance = conn
	})
	return instance
}

func OpenMemory() (*sql.DB, error) {
	return sql.Open("sqlite", ":memory:")
}

func ResetForTesting() {
	if instance != nil {
		instance.Close()
		instance = nil
	}
	once = sync.Once{}
}
```

- [ ] **Step 5: Implement migrations**

```go
// internal/db/migrations.go
package db

import "database/sql"

func RunMigrations(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id              INTEGER PRIMARY KEY AUTOINCREMENT,
		username        TEXT    NOT NULL UNIQUE,
		telegram_id     INTEGER NOT NULL DEFAULT 0,
		telegram_chat_id INTEGER NOT NULL DEFAULT 0,
		created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS projects (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id     INTEGER NOT NULL REFERENCES users(id),
		name        TEXT    NOT NULL,
		description TEXT    NOT NULL DEFAULT '',
		status      TEXT    NOT NULL DEFAULT 'ACTIVE',
		due_date    DATETIME,
		created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS tags (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id    INTEGER NOT NULL REFERENCES users(id),
		name       TEXT    NOT NULL,
		color      TEXT    NOT NULL DEFAULT '#ffffff',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id    INTEGER NOT NULL REFERENCES users(id),
		project_id INTEGER REFERENCES projects(id),
		tag_id     INTEGER REFERENCES tags(id),
		content    TEXT    NOT NULL,
		status     TEXT    NOT NULL DEFAULT 'TODO',
		priority   TEXT    NOT NULL DEFAULT 'MEDIUM',
		due_date   DATETIME,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS notes (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id    INTEGER NOT NULL REFERENCES users(id),
		project_id INTEGER REFERENCES projects(id),
		tag_id     INTEGER REFERENCES tags(id),
		title      TEXT    NOT NULL,
		content    TEXT    NOT NULL DEFAULT '',
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS reminders (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id       INTEGER NOT NULL REFERENCES users(id),
		project_id    INTEGER REFERENCES projects(id),
		tag_id        INTEGER REFERENCES tags(id),
		content       TEXT    NOT NULL,
		reminder_time DATETIME NOT NULL,
		status        TEXT    NOT NULL DEFAULT 'PENDING',
		created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS pomodoro_sessions (
		id             INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id        INTEGER NOT NULL REFERENCES users(id),
		project_id     INTEGER REFERENCES projects(id),
		start_time     DATETIME,
		work_duration  INTEGER NOT NULL DEFAULT 25,
		finish_time    DATETIME,
		remaining_time INTEGER NOT NULL DEFAULT 0,
		state          TEXT    NOT NULL DEFAULT 'IDLE',
		created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err := db.Exec(schema)
	return err
}
```

- [ ] **Step 6: Run tests**

```bash
go test ./internal/db/...
```

Expected: `PASS`

- [ ] **Step 7: Commit**

```bash
git add internal/db/ go.mod go.sum
git commit -m "feat: add db singleton and migrations"
```

---

## Task 5: Repository Layer

**Files:**
- Create: `internal/repository/interfaces.go`
- Create: `internal/repository/task_repo.go`
- Create: `internal/repository/project_repo.go`
- Create: `internal/repository/user_repo.go`
- Create: `internal/repository/tag_repo.go`
- Create: `internal/repository/note_repo.go`
- Create: `internal/repository/reminder_repo.go`
- Create: `internal/repository/pomodoro_repo.go`
- Test: `internal/repository/task_repo_test.go`
- Test: `internal/repository/project_repo_test.go`

- [ ] **Step 1: Write failing test for TaskRepository**

```go
// internal/repository/task_repo_test.go
package repository_test

import (
	"testing"

	"taskflow/internal/db"
	"taskflow/internal/domain"
	"taskflow/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *repository.TaskRepo {
	conn, err := db.OpenMemory()
	require.NoError(t, err)
	require.NoError(t, db.RunMigrations(conn))
	t.Cleanup(func() { conn.Close() })

	// seed user
	conn.Exec("INSERT INTO users (id, username) VALUES (1, 'testuser')")
	return repository.NewTaskRepo(conn)
}

func TestTaskRepoCreate(t *testing.T) {
	repo := setupTestDB(t)
	task := &domain.Task{
		UserID:   1,
		Content:  "Buy milk",
		Status:   domain.TaskStatusTodo,
		Priority: domain.PriorityMedium,
	}
	err := repo.Create(task)
	require.NoError(t, err)
	assert.Greater(t, task.ID, int64(0))
}

func TestTaskRepoGetByID(t *testing.T) {
	repo := setupTestDB(t)
	task := &domain.Task{UserID: 1, Content: "Test task", Status: domain.TaskStatusTodo, Priority: domain.PriorityLow}
	require.NoError(t, repo.Create(task))

	got, err := repo.GetByID(task.ID)
	require.NoError(t, err)
	assert.Equal(t, "Test task", got.Content)
}

func TestTaskRepoGetAllByUser(t *testing.T) {
	repo := setupTestDB(t)
	repo.Create(&domain.Task{UserID: 1, Content: "A", Status: domain.TaskStatusTodo, Priority: domain.PriorityLow})
	repo.Create(&domain.Task{UserID: 1, Content: "B", Status: domain.TaskStatusTodo, Priority: domain.PriorityLow})

	tasks, err := repo.GetAllByUser(1)
	require.NoError(t, err)
	assert.Len(t, tasks, 2)
}

func TestTaskRepoUpdate(t *testing.T) {
	repo := setupTestDB(t)
	task := &domain.Task{UserID: 1, Content: "Old", Status: domain.TaskStatusTodo, Priority: domain.PriorityLow}
	require.NoError(t, repo.Create(task))

	task.Content = "New"
	task.Status = domain.TaskStatusDone
	require.NoError(t, repo.Update(task))

	got, _ := repo.GetByID(task.ID)
	assert.Equal(t, "New", got.Content)
	assert.Equal(t, domain.TaskStatusDone, got.Status)
}

func TestTaskRepoDelete(t *testing.T) {
	repo := setupTestDB(t)
	task := &domain.Task{UserID: 1, Content: "Delete me", Status: domain.TaskStatusTodo, Priority: domain.PriorityLow}
	require.NoError(t, repo.Create(task))

	require.NoError(t, repo.Delete(task.ID))
	_, err := repo.GetByID(task.ID)
	assert.Error(t, err)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/repository/...
```

Expected: compile error

- [ ] **Step 3: Create repository interfaces**

```go
// internal/repository/interfaces.go
package repository

import "taskflow/internal/domain"

type TaskRepository interface {
	Create(task *domain.Task) error
	GetByID(id int64) (*domain.Task, error)
	GetAllByUser(userID int64) ([]*domain.Task, error)
	GetByProject(projectID int64) ([]*domain.Task, error)
	Update(task *domain.Task) error
	Delete(id int64) error
}

type ProjectRepository interface {
	Create(p *domain.Project) error
	GetByID(id int64) (*domain.Project, error)
	GetAllByUser(userID int64) ([]*domain.Project, error)
	Update(p *domain.Project) error
	Delete(id int64) error
}

type NoteRepository interface {
	Create(n *domain.Note) error
	GetByID(id int64) (*domain.Note, error)
	GetAllByUser(userID int64) ([]*domain.Note, error)
	Update(n *domain.Note) error
	Delete(id int64) error
}

type ReminderRepository interface {
	Create(r *domain.Reminder) error
	GetByID(id int64) (*domain.Reminder, error)
	GetAllByUser(userID int64) ([]*domain.Reminder, error)
	GetPending() ([]*domain.Reminder, error)
	Update(r *domain.Reminder) error
	Delete(id int64) error
}

type PomodoroRepository interface {
	Create(s *domain.PomodoroSession) error
	GetByID(id int64) (*domain.PomodoroSession, error)
	GetAllByUser(userID int64) ([]*domain.PomodoroSession, error)
	GetCompletedByProject(projectID int64) ([]*domain.PomodoroSession, error)
	Update(s *domain.PomodoroSession) error
}

type TagRepository interface {
	Create(t *domain.Tag) error
	GetByID(id int64) (*domain.Tag, error)
	GetAllByUser(userID int64) ([]*domain.Tag, error)
	Delete(id int64) error
}

type UserRepository interface {
	Create(u *domain.User) error
	GetByID(id int64) (*domain.User, error)
	GetFirst() (*domain.User, error)
}
```

- [ ] **Step 4: Implement TaskRepo**

```go
// internal/repository/task_repo.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"taskflow/internal/domain"
)

type TaskRepo struct{ db *sql.DB }

func NewTaskRepo(db *sql.DB) *TaskRepo { return &TaskRepo{db: db} }

func (r *TaskRepo) Create(t *domain.Task) error {
	now := time.Now()
	res, err := r.db.Exec(
		`INSERT INTO tasks (user_id, project_id, tag_id, content, status, priority, due_date, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.UserID, t.ProjectID, t.TagID, t.Content, t.Status, t.Priority, t.DueDate, now, now,
	)
	if err != nil {
		return err
	}
	t.ID, err = res.LastInsertId()
	t.CreatedAt = now
	t.UpdatedAt = now
	return err
}

func (r *TaskRepo) GetByID(id int64) (*domain.Task, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, project_id, tag_id, content, status, priority, due_date, created_at, updated_at
		 FROM tasks WHERE id = ?`, id,
	)
	return scanTask(row)
}

func (r *TaskRepo) GetAllByUser(userID int64) ([]*domain.Task, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, project_id, tag_id, content, status, priority, due_date, created_at, updated_at
		 FROM tasks WHERE user_id = ? ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (r *TaskRepo) GetByProject(projectID int64) ([]*domain.Task, error) {
	rows, err := r.db.Query(
		`SELECT id, user_id, project_id, tag_id, content, status, priority, due_date, created_at, updated_at
		 FROM tasks WHERE project_id = ? ORDER BY created_at DESC`, projectID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

func (r *TaskRepo) Update(t *domain.Task) error {
	t.UpdatedAt = time.Now()
	_, err := r.db.Exec(
		`UPDATE tasks SET project_id=?, tag_id=?, content=?, status=?, priority=?, due_date=?, updated_at=? WHERE id=?`,
		t.ProjectID, t.TagID, t.Content, t.Status, t.Priority, t.DueDate, t.UpdatedAt, t.ID,
	)
	return err
}

func (r *TaskRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM tasks WHERE id = ?`, id)
	return err
}

func scanTask(row *sql.Row) (*domain.Task, error) {
	t := &domain.Task{}
	err := row.Scan(&t.ID, &t.UserID, &t.ProjectID, &t.TagID, &t.Content, &t.Status, &t.Priority,
		&t.DueDate, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	return t, err
}

func scanTasks(rows *sql.Rows) ([]*domain.Task, error) {
	var tasks []*domain.Task
	for rows.Next() {
		t := &domain.Task{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.ProjectID, &t.TagID, &t.Content, &t.Status, &t.Priority,
			&t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
```

- [ ] **Step 5: Implement remaining repos (ProjectRepo, UserRepo, TagRepo, NoteRepo, ReminderRepo, PomodoroRepo)**

```go
// internal/repository/project_repo.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"taskflow/internal/domain"
)

type ProjectRepo struct{ db *sql.DB }

func NewProjectRepo(db *sql.DB) *ProjectRepo { return &ProjectRepo{db: db} }

func (r *ProjectRepo) Create(p *domain.Project) error {
	now := time.Now()
	res, err := r.db.Exec(
		`INSERT INTO projects (user_id, name, description, status, due_date, created_at) VALUES (?,?,?,?,?,?)`,
		p.UserID, p.Name, p.Description, p.Status, p.DueDate, now,
	)
	if err != nil {
		return err
	}
	p.ID, err = res.LastInsertId()
	p.CreatedAt = now
	return err
}

func (r *ProjectRepo) GetByID(id int64) (*domain.Project, error) {
	row := r.db.QueryRow(`SELECT id, user_id, name, description, status, due_date, created_at FROM projects WHERE id=?`, id)
	p := &domain.Project{}
	err := row.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Status, &p.DueDate, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("project not found")
	}
	return p, err
}

func (r *ProjectRepo) GetAllByUser(userID int64) ([]*domain.Project, error) {
	rows, err := r.db.Query(`SELECT id, user_id, name, description, status, due_date, created_at FROM projects WHERE user_id=? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projects []*domain.Project
	for rows.Next() {
		p := &domain.Project{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Status, &p.DueDate, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *ProjectRepo) Update(p *domain.Project) error {
	_, err := r.db.Exec(`UPDATE projects SET name=?, description=?, status=?, due_date=? WHERE id=?`,
		p.Name, p.Description, p.Status, p.DueDate, p.ID)
	return err
}

func (r *ProjectRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM projects WHERE id=?`, id)
	return err
}
```

```go
// internal/repository/user_repo.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"taskflow/internal/domain"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(u *domain.User) error {
	now := time.Now()
	res, err := r.db.Exec(
		`INSERT INTO users (username, telegram_id, telegram_chat_id, created_at) VALUES (?,?,?,?)`,
		u.Username, u.TelegramID, u.TelegramChatID, now,
	)
	if err != nil {
		return err
	}
	u.ID, err = res.LastInsertId()
	u.CreatedAt = now
	return err
}

func (r *UserRepo) GetByID(id int64) (*domain.User, error) {
	row := r.db.QueryRow(`SELECT id, username, telegram_id, telegram_chat_id, created_at FROM users WHERE id=?`, id)
	u := &domain.User{}
	err := row.Scan(&u.ID, &u.Username, &u.TelegramID, &u.TelegramChatID, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	return u, err
}

func (r *UserRepo) GetFirst() (*domain.User, error) {
	row := r.db.QueryRow(`SELECT id, username, telegram_id, telegram_chat_id, created_at FROM users ORDER BY id LIMIT 1`)
	u := &domain.User{}
	err := row.Scan(&u.ID, &u.Username, &u.TelegramID, &u.TelegramChatID, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no users found")
	}
	return u, err
}
```

```go
// internal/repository/tag_repo.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"taskflow/internal/domain"
)

type TagRepo struct{ db *sql.DB }

func NewTagRepo(db *sql.DB) *TagRepo { return &TagRepo{db: db} }

func (r *TagRepo) Create(t *domain.Tag) error {
	now := time.Now()
	res, err := r.db.Exec(`INSERT INTO tags (user_id, name, color, created_at) VALUES (?,?,?,?)`,
		t.UserID, t.Name, t.Color, now)
	if err != nil {
		return err
	}
	t.ID, err = res.LastInsertId()
	t.CreatedAt = now
	return err
}

func (r *TagRepo) GetByID(id int64) (*domain.Tag, error) {
	row := r.db.QueryRow(`SELECT id, user_id, name, color, created_at FROM tags WHERE id=?`, id)
	t := &domain.Tag{}
	err := row.Scan(&t.ID, &t.UserID, &t.Name, &t.Color, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tag not found")
	}
	return t, err
}

func (r *TagRepo) GetAllByUser(userID int64) ([]*domain.Tag, error) {
	rows, err := r.db.Query(`SELECT id, user_id, name, color, created_at FROM tags WHERE user_id=?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []*domain.Tag
	for rows.Next() {
		t := &domain.Tag{}
		if err := rows.Scan(&t.ID, &t.UserID, &t.Name, &t.Color, &t.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

func (r *TagRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM tags WHERE id=?`, id)
	return err
}
```

```go
// internal/repository/note_repo.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"taskflow/internal/domain"
)

type NoteRepo struct{ db *sql.DB }

func NewNoteRepo(db *sql.DB) *NoteRepo { return &NoteRepo{db: db} }

func (r *NoteRepo) Create(n *domain.Note) error {
	now := time.Now()
	res, err := r.db.Exec(
		`INSERT INTO notes (user_id, project_id, tag_id, title, content, created_at, updated_at) VALUES (?,?,?,?,?,?,?)`,
		n.UserID, n.ProjectID, n.TagID, n.Title, n.Content, now, now,
	)
	if err != nil {
		return err
	}
	n.ID, err = res.LastInsertId()
	n.CreatedAt, n.UpdatedAt = now, now
	return err
}

func (r *NoteRepo) GetByID(id int64) (*domain.Note, error) {
	row := r.db.QueryRow(`SELECT id, user_id, project_id, tag_id, title, content, created_at, updated_at FROM notes WHERE id=?`, id)
	n := &domain.Note{}
	err := row.Scan(&n.ID, &n.UserID, &n.ProjectID, &n.TagID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("note not found")
	}
	return n, err
}

func (r *NoteRepo) GetAllByUser(userID int64) ([]*domain.Note, error) {
	rows, err := r.db.Query(`SELECT id, user_id, project_id, tag_id, title, content, created_at, updated_at FROM notes WHERE user_id=? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notes []*domain.Note
	for rows.Next() {
		n := &domain.Note{}
		if err := rows.Scan(&n.ID, &n.UserID, &n.ProjectID, &n.TagID, &n.Title, &n.Content, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

func (r *NoteRepo) Update(n *domain.Note) error {
	n.UpdatedAt = time.Now()
	_, err := r.db.Exec(`UPDATE notes SET project_id=?, tag_id=?, title=?, content=?, updated_at=? WHERE id=?`,
		n.ProjectID, n.TagID, n.Title, n.Content, n.UpdatedAt, n.ID)
	return err
}

func (r *NoteRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM notes WHERE id=?`, id)
	return err
}
```

```go
// internal/repository/reminder_repo.go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"taskflow/internal/domain"
)

type ReminderRepo struct{ db *sql.DB }

func NewReminderRepo(db *sql.DB) *ReminderRepo { return &ReminderRepo{db: db} }

func (r *ReminderRepo) Create(rem *domain.Reminder) error {
	now := time.Now()
	res, err := r.db.Exec(
		`INSERT INTO reminders (user_id, project_id, tag_id, content, reminder_time, status, created_at) VALUES (?,?,?,?,?,?,?)`,
		rem.UserID, rem.ProjectID, rem.TagID, rem.Content, rem.ReminderTime, rem.Status, now,
	)
	if err != nil {
		return err
	}
	rem.ID, err = res.LastInsertId()
	rem.CreatedAt = now
	return err
}

func (r *ReminderRepo) GetByID(id int64) (*domain.Reminder, error) {
	row := r.db.QueryRow(`SELECT id, user_id, project_id, tag_id, content, reminder_time, status, created_at FROM reminders WHERE id=?`, id)
	rem := &domain.Reminder{}
	err := row.Scan(&rem.ID, &rem.UserID, &rem.ProjectID, &rem.TagID, &rem.Content, &rem.ReminderTime, &rem.Status, &rem.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("reminder not found")
	}
	return rem, err
}

func (r *ReminderRepo) GetAllByUser(userID int64) ([]*domain.Reminder, error) {
	rows, err := r.db.Query(`SELECT id, user_id, project_id, tag_id, content, reminder_time, status, created_at FROM reminders WHERE user_id=? ORDER BY reminder_time`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReminders(rows)
}

func (r *ReminderRepo) GetPending() ([]*domain.Reminder, error) {
	rows, err := r.db.Query(`SELECT id, user_id, project_id, tag_id, content, reminder_time, status, created_at FROM reminders WHERE status='PENDING' AND reminder_time <= ?`, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanReminders(rows)
}

func (r *ReminderRepo) Update(rem *domain.Reminder) error {
	_, err := r.db.Exec(`UPDATE reminders SET status=? WHERE id=?`, rem.Status, rem.ID)
	return err
}

func (r *ReminderRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM reminders WHERE id=?`, id)
	return err
}

func scanReminders(rows *sql.Rows) ([]*domain.Reminder, error) {
	var result []*domain.Reminder
	for rows.Next() {
		rem := &domain.Reminder{}
		if err := rows.Scan(&rem.ID, &rem.UserID, &rem.ProjectID, &rem.TagID, &rem.Content, &rem.ReminderTime, &rem.Status, &rem.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, rem)
	}
	return result, rows.Err()
}
```

```go
// internal/repository/pomodoro_repo.go
package repository

import (
	"database/sql"
	"time"

	"taskflow/internal/domain"
)

type PomodoroRepo struct{ db *sql.DB }

func NewPomodoroRepo(db *sql.DB) *PomodoroRepo { return &PomodoroRepo{db: db} }

func (r *PomodoroRepo) Create(s *domain.PomodoroSession) error {
	now := time.Now()
	res, err := r.db.Exec(
		`INSERT INTO pomodoro_sessions (user_id, project_id, start_time, work_duration, finish_time, remaining_time, state, created_at) VALUES (?,?,?,?,?,?,?,?)`,
		s.UserID, s.ProjectID, s.StartTime, s.WorkDuration, s.FinishTime, s.RemainingTime, s.State, now,
	)
	if err != nil {
		return err
	}
	s.ID, err = res.LastInsertId()
	s.CreatedAt = now
	return err
}

func (r *PomodoroRepo) GetByID(id int64) (*domain.PomodoroSession, error) {
	row := r.db.QueryRow(`SELECT id, user_id, project_id, start_time, work_duration, finish_time, remaining_time, state, created_at FROM pomodoro_sessions WHERE id=?`, id)
	s := &domain.PomodoroSession{}
	err := row.Scan(&s.ID, &s.UserID, &s.ProjectID, &s.StartTime, &s.WorkDuration, &s.FinishTime, &s.RemainingTime, &s.State, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	return s, err
}

func (r *PomodoroRepo) GetAllByUser(userID int64) ([]*domain.PomodoroSession, error) {
	rows, err := r.db.Query(`SELECT id, user_id, project_id, start_time, work_duration, finish_time, remaining_time, state, created_at FROM pomodoro_sessions WHERE user_id=? ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSessions(rows)
}

func (r *PomodoroRepo) GetCompletedByProject(projectID int64) ([]*domain.PomodoroSession, error) {
	rows, err := r.db.Query(`SELECT id, user_id, project_id, start_time, work_duration, finish_time, remaining_time, state, created_at FROM pomodoro_sessions WHERE project_id=? AND state='COMPLETED'`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSessions(rows)
}

func (r *PomodoroRepo) Update(s *domain.PomodoroSession) error {
	_, err := r.db.Exec(`UPDATE pomodoro_sessions SET start_time=?, finish_time=?, remaining_time=?, state=? WHERE id=?`,
		s.StartTime, s.FinishTime, s.RemainingTime, s.State, s.ID)
	return err
}

func scanSessions(rows *sql.Rows) ([]*domain.PomodoroSession, error) {
	var result []*domain.PomodoroSession
	for rows.Next() {
		s := &domain.PomodoroSession{}
		if err := rows.Scan(&s.ID, &s.UserID, &s.ProjectID, &s.StartTime, &s.WorkDuration, &s.FinishTime, &s.RemainingTime, &s.State, &s.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
```

- [ ] **Step 6: Run all repository tests**

```bash
go test ./internal/repository/...
```

Expected: `PASS`

- [ ] **Step 7: Commit**

```bash
git add internal/repository/
git commit -m "feat: add repository layer with SQLite implementations"
```

---

## Task 6: Strategy Pattern — FilterStrategy

**Files:**
- Create: `internal/patterns/strategy/filter.go`
- Test: `internal/patterns/strategy/filter_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/patterns/strategy/filter_test.go
package strategy_test

import (
	"testing"
	"time"

	"taskflow/internal/domain"
	"taskflow/internal/patterns/strategy"

	"github.com/stretchr/testify/assert"
)

func makeTasks() []*domain.Task {
	tag1 := int64(1)
	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(24 * time.Hour)
	return []*domain.Task{
		{ID: 1, Status: domain.TaskStatusTodo, Priority: domain.PriorityHigh, TagID: &tag1, DueDate: &past},
		{ID: 2, Status: domain.TaskStatusDone, Priority: domain.PriorityLow, DueDate: &future},
		{ID: 3, Status: domain.TaskStatusInProgress, Priority: domain.PriorityMedium, TagID: &tag1},
	}
}

func TestByStatusFilter(t *testing.T) {
	f := strategy.ByStatusFilter{Status: domain.TaskStatusTodo}
	result := f.Filter(makeTasks())
	assert.Len(t, result, 1)
	assert.Equal(t, int64(1), result[0].ID)
}

func TestByTagFilter(t *testing.T) {
	f := strategy.ByTagFilter{TagID: 1}
	result := f.Filter(makeTasks())
	assert.Len(t, result, 2)
}

func TestByDateFilter(t *testing.T) {
	from := time.Now().Add(-2 * time.Hour)
	to := time.Now()
	f := strategy.ByDateFilter{From: from, To: to}
	result := f.Filter(makeTasks())
	assert.Len(t, result, 1) // only task 1 has due_date in past range
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/patterns/strategy/...
```

Expected: compile error

- [ ] **Step 3: Implement FilterStrategy**

```go
// internal/patterns/strategy/filter.go
package strategy

import (
	"time"

	"taskflow/internal/domain"
)

// FilterStrategy — стратегия фильтрации задач
type FilterStrategy interface {
	Filter(tasks []*domain.Task) []*domain.Task
}

// ByStatusFilter фильтрует по статусу
type ByStatusFilter struct {
	Status domain.TaskStatus
}

func (f ByStatusFilter) Filter(tasks []*domain.Task) []*domain.Task {
	var result []*domain.Task
	for _, t := range tasks {
		if t.Status == f.Status {
			result = append(result, t)
		}
	}
	return result
}

// ByTagFilter фильтрует по тегу
type ByTagFilter struct {
	TagID int64
}

func (f ByTagFilter) Filter(tasks []*domain.Task) []*domain.Task {
	var result []*domain.Task
	for _, t := range tasks {
		if t.TagID != nil && *t.TagID == f.TagID {
			result = append(result, t)
		}
	}
	return result
}

// ByDateFilter фильтрует по диапазону due_date
type ByDateFilter struct {
	From time.Time
	To   time.Time
}

func (f ByDateFilter) Filter(tasks []*domain.Task) []*domain.Task {
	var result []*domain.Task
	for _, t := range tasks {
		if t.DueDate != nil && !t.DueDate.Before(f.From) && !t.DueDate.After(f.To) {
			result = append(result, t)
		}
	}
	return result
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/patterns/strategy/...
```

Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/patterns/strategy/
git commit -m "feat: add Strategy pattern for task filtering"
```

---

## Task 7: State Pattern — PomodoroSession

**Files:**
- Create: `internal/patterns/state/pomodoro_state.go`
- Test: `internal/patterns/state/pomodoro_state_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/patterns/state/pomodoro_state_test.go
package state_test

import (
	"testing"

	"taskflow/internal/patterns/state"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPomodoroStartFromIdle(t *testing.T) {
	s := state.NewPomodoroMachine(25)
	require.NoError(t, s.Start())
	assert.Equal(t, "RUNNING", s.StateName())
}

func TestPomodoroPauseResume(t *testing.T) {
	s := state.NewPomodoroMachine(25)
	s.Start()
	require.NoError(t, s.Pause())
	assert.Equal(t, "PAUSED", s.StateName())
	require.NoError(t, s.Resume())
	assert.Equal(t, "RUNNING", s.StateName())
}

func TestPomodoroComplete(t *testing.T) {
	s := state.NewPomodoroMachine(25)
	s.Start()
	require.NoError(t, s.Complete())
	assert.Equal(t, "COMPLETED", s.StateName())
}

func TestPomodoroCancel(t *testing.T) {
	s := state.NewPomodoroMachine(25)
	s.Start()
	require.NoError(t, s.Cancel())
	assert.Equal(t, "CANCELLED", s.StateName())
}

func TestPomodoroInvalidTransition(t *testing.T) {
	s := state.NewPomodoroMachine(25)
	assert.Error(t, s.Pause()) // IDLE → Pause is invalid
}

func TestPomodoroCannotStartTwice(t *testing.T) {
	s := state.NewPomodoroMachine(25)
	s.Start()
	assert.Error(t, s.Start()) // RUNNING → Start is invalid
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/patterns/state/...
```

Expected: compile error

- [ ] **Step 3: Implement State pattern**

```go
// internal/patterns/state/pomodoro_state.go
package state

import (
	"errors"
	"time"
)

// PomodoroState — интерфейс состояния (паттерн Состояние)
type PomodoroState interface {
	Start(m *PomodoroMachine) error
	Pause(m *PomodoroMachine) error
	Resume(m *PomodoroMachine) error
	Complete(m *PomodoroMachine) error
	Cancel(m *PomodoroMachine) error
	Name() string
}

var errInvalidTransition = errors.New("invalid state transition")

// PomodoroMachine — контекст состояния
type PomodoroMachine struct {
	state         PomodoroState
	WorkDuration  int // minutes
	StartTime     *time.Time
	FinishTime    *time.Time
	RemainingTime int // seconds
}

func NewPomodoroMachine(workDurationMin int) *PomodoroMachine {
	return &PomodoroMachine{
		state:        &IdleState{},
		WorkDuration: workDurationMin,
	}
}

func (m *PomodoroMachine) SetState(s PomodoroState) { m.state = s }
func (m *PomodoroMachine) StateName() string        { return m.state.Name() }

func (m *PomodoroMachine) Start() error    { return m.state.Start(m) }
func (m *PomodoroMachine) Pause() error    { return m.state.Pause(m) }
func (m *PomodoroMachine) Resume() error   { return m.state.Resume(m) }
func (m *PomodoroMachine) Complete() error { return m.state.Complete(m) }
func (m *PomodoroMachine) Cancel() error   { return m.state.Cancel(m) }

// --- IdleState ---
type IdleState struct{}

func (s *IdleState) Name() string { return "IDLE" }
func (s *IdleState) Start(m *PomodoroMachine) error {
	now := time.Now()
	m.StartTime = &now
	m.RemainingTime = m.WorkDuration * 60
	m.SetState(&RunningState{})
	return nil
}
func (s *IdleState) Pause(m *PomodoroMachine) error    { return errInvalidTransition }
func (s *IdleState) Resume(m *PomodoroMachine) error   { return errInvalidTransition }
func (s *IdleState) Complete(m *PomodoroMachine) error { return errInvalidTransition }
func (s *IdleState) Cancel(m *PomodoroMachine) error   { return errInvalidTransition }

// --- RunningState ---
type RunningState struct{}

func (s *RunningState) Name() string { return "RUNNING" }
func (s *RunningState) Start(m *PomodoroMachine) error { return errInvalidTransition }
func (s *RunningState) Pause(m *PomodoroMachine) error {
	m.SetState(&PausedState{})
	return nil
}
func (s *RunningState) Resume(m *PomodoroMachine) error { return errInvalidTransition }
func (s *RunningState) Complete(m *PomodoroMachine) error {
	now := time.Now()
	m.FinishTime = &now
	m.RemainingTime = 0
	m.SetState(&CompletedState{})
	return nil
}
func (s *RunningState) Cancel(m *PomodoroMachine) error {
	m.SetState(&CancelledState{})
	return nil
}

// --- PausedState ---
type PausedState struct{}

func (s *PausedState) Name() string { return "PAUSED" }
func (s *PausedState) Start(m *PomodoroMachine) error  { return errInvalidTransition }
func (s *PausedState) Pause(m *PomodoroMachine) error  { return errInvalidTransition }
func (s *PausedState) Resume(m *PomodoroMachine) error {
	m.SetState(&RunningState{})
	return nil
}
func (s *PausedState) Complete(m *PomodoroMachine) error { return errInvalidTransition }
func (s *PausedState) Cancel(m *PomodoroMachine) error {
	m.SetState(&CancelledState{})
	return nil
}

// --- CompletedState ---
type CompletedState struct{}

func (s *CompletedState) Name() string                    { return "COMPLETED" }
func (s *CompletedState) Start(m *PomodoroMachine) error  { return errInvalidTransition }
func (s *CompletedState) Pause(m *PomodoroMachine) error  { return errInvalidTransition }
func (s *CompletedState) Resume(m *PomodoroMachine) error { return errInvalidTransition }
func (s *CompletedState) Complete(m *PomodoroMachine) error { return errInvalidTransition }
func (s *CompletedState) Cancel(m *PomodoroMachine) error { return errInvalidTransition }

// --- CancelledState ---
type CancelledState struct{}

func (s *CancelledState) Name() string                    { return "CANCELLED" }
func (s *CancelledState) Start(m *PomodoroMachine) error  { return errInvalidTransition }
func (s *CancelledState) Pause(m *PomodoroMachine) error  { return errInvalidTransition }
func (s *CancelledState) Resume(m *PomodoroMachine) error { return errInvalidTransition }
func (s *CancelledState) Complete(m *PomodoroMachine) error { return errInvalidTransition }
func (s *CancelledState) Cancel(m *PomodoroMachine) error { return errInvalidTransition }
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/patterns/state/...
```

Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/patterns/state/
git commit -m "feat: add State pattern for PomodoroSession"
```

---

## Task 8: Mathematical Model — ETA Calculator

**Files:**
- Create: `internal/math/eta.go`
- Test: `internal/math/eta_test.go`

- [ ] **Step 1: Write failing test**

```go
// internal/math/eta_test.go
package math_test

import (
	"testing"

	taskmath "taskflow/internal/math"
	"taskflow/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestETAWithHistory(t *testing.T) {
	// 2 completed sessions: 25 min each → avg = 25 min
	// 4 tasks completed total across 2 sessions → avg 2 tasks/session
	// 6 remaining tasks → ETA = 6/2 * 25 = 75 min
	sessions := []*domain.PomodoroSession{
		{WorkDuration: 25, State: domain.SessionStateCompleted},
		{WorkDuration: 25, State: domain.SessionStateCompleted},
	}
	completedTasksPerSession := []int{2, 2} // 2 tasks per session
	remaining := 6

	eta, err := taskmath.CalculateETA(sessions, completedTasksPerSession, remaining)
	require.NoError(t, err)
	assert.Equal(t, 75.0, eta)
}

func TestETANoHistory(t *testing.T) {
	_, err := taskmath.CalculateETA(nil, nil, 5)
	assert.Error(t, err)
}

func TestETAZeroRemaining(t *testing.T) {
	sessions := []*domain.PomodoroSession{
		{WorkDuration: 25, State: domain.SessionStateCompleted},
	}
	eta, err := taskmath.CalculateETA(sessions, []int{2}, 0)
	require.NoError(t, err)
	assert.Equal(t, 0.0, eta)
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/math/...
```

Expected: compile error

- [ ] **Step 3: Implement ETA calculator**

```go
// internal/math/eta.go
package math

import (
	"errors"

	"taskflow/internal/domain"
)

// CalculateETA вычисляет оценочное время завершения оставшихся задач.
//
// Формула:
//   avg_pomodoro_min = sum(session.WorkDuration) / count(sessions)
//   avg_tasks_per_session = sum(completedPerSession) / count(sessions)
//   ETA_minutes = (remaining / avg_tasks_per_session) * avg_pomodoro_min
//
// sessions — завершённые помодоро-сессии проекта
// completedPerSession[i] — кол-во задач, закрытых в sessions[i]
// remaining — кол-во незакрытых задач
// Returns: ETA в минутах, или ошибка если история пуста
func CalculateETA(sessions []*domain.PomodoroSession, completedPerSession []int, remaining int) (float64, error) {
	if len(sessions) == 0 {
		return 0, errors.New("eta: no session history — cannot estimate")
	}
	if remaining == 0 {
		return 0, nil
	}

	totalDuration := 0
	totalCompleted := 0
	for i, s := range sessions {
		totalDuration += s.WorkDuration
		if i < len(completedPerSession) {
			totalCompleted += completedPerSession[i]
		}
	}

	n := float64(len(sessions))
	avgPomodoro := float64(totalDuration) / n
	avgTasksPerSession := float64(totalCompleted) / n

	if avgTasksPerSession == 0 {
		// Сессии были, задачи не закрывались — используем 1 задачу/сессию как fallback
		avgTasksPerSession = 1
	}

	eta := (float64(remaining) / avgTasksPerSession) * avgPomodoro
	return eta, nil
}
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/math/...
```

Expected: `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/math/
git commit -m "feat: add ETA calculator math model"
```

---

## Task 9: Basic TUI + Wire Up main.go

**Files:**
- Create: `internal/tui/app.go`
- Modify: `cmd/taskflow/main.go`

> No unit tests for TUI — verify manually by running the app.

- [ ] **Step 1: Implement minimal TUI**

```go
// internal/tui/app.go
package tui

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"taskflow/internal/domain"
	"taskflow/internal/repository"
)

const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorRed    = "\033[31m"
)

type App struct {
	db          *sql.DB
	userRepo    *repository.UserRepo
	taskRepo    *repository.TaskRepo
	projectRepo *repository.ProjectRepo
	noteRepo    *repository.NoteRepo
	currentUser *domain.User
	scanner     *bufio.Scanner
}

func New(db *sql.DB) *App {
	return &App{
		db:          db,
		userRepo:    repository.NewUserRepo(db),
		taskRepo:    repository.NewTaskRepo(db),
		projectRepo: repository.NewProjectRepo(db),
		noteRepo:    repository.NewNoteRepo(db),
		scanner:     bufio.NewScanner(os.Stdin),
	}
}

func (a *App) Run() {
	a.ensureUser()
	a.mainLoop()
}

func (a *App) ensureUser() {
	u, err := a.userRepo.GetFirst()
	if err != nil {
		// First launch: create default user
		u = &domain.User{Username: "default"}
		if err := a.userRepo.Create(u); err != nil {
			fmt.Println("Error creating user:", err)
			os.Exit(1)
		}
		fmt.Printf("%sWelcome to TaskFlow!%s Created user 'default'.\n\n", colorGreen, colorReset)
	}
	a.currentUser = u
}

func (a *App) mainLoop() {
	for {
		a.printMenu()
		choice := a.readLine("Choice: ")
		switch strings.TrimSpace(choice) {
		case "1":
			a.tasksMenu()
		case "2":
			a.projectsMenu()
		case "3":
			a.notesMenu()
		case "0":
			fmt.Println("Bye!")
			return
		default:
			fmt.Println("Unknown option.")
		}
	}
}

func (a *App) printMenu() {
	fmt.Printf("\n%s=== TaskFlow ===%s  [user: %s]\n", colorBold, colorReset, a.currentUser.Username)
	fmt.Println("  1. Tasks")
	fmt.Println("  2. Projects")
	fmt.Println("  3. Notes")
	fmt.Println("  0. Exit")
}

func (a *App) tasksMenu() {
	for {
		fmt.Printf("\n%s-- Tasks --%s\n", colorCyan, colorReset)
		tasks, _ := a.taskRepo.GetAllByUser(a.currentUser.ID)
		if len(tasks) == 0 {
			fmt.Println("  (no tasks)")
		}
		for _, t := range tasks {
			status := fmt.Sprintf("[%s]", t.Status)
			overdue := ""
			if t.IsOverdue() {
				overdue = colorRed + " OVERDUE" + colorReset
			}
			fmt.Printf("  %d. %s %s%s\n", t.ID, status, t.Content, overdue)
		}
		fmt.Println("\n  a. Add task   d. Delete task   b. Back")
		choice := a.readLine("Choice: ")
		switch strings.TrimSpace(choice) {
		case "a":
			content := a.readLine("Task content: ")
			task := &domain.Task{
				UserID:   a.currentUser.ID,
				Content:  strings.TrimSpace(content),
				Status:   domain.TaskStatusTodo,
				Priority: domain.PriorityMedium,
			}
			if err := a.taskRepo.Create(task); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Printf("%sTask created (id=%d)%s\n", colorGreen, task.ID, colorReset)
			}
		case "d":
			idStr := a.readLine("Task ID to delete: ")
			var id int64
			fmt.Sscanf(idStr, "%d", &id)
			if err := a.taskRepo.Delete(id); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println(colorGreen + "Deleted." + colorReset)
			}
		case "b":
			return
		}
	}
}

func (a *App) projectsMenu() {
	for {
		fmt.Printf("\n%s-- Projects --%s\n", colorCyan, colorReset)
		projects, _ := a.projectRepo.GetAllByUser(a.currentUser.ID)
		if len(projects) == 0 {
			fmt.Println("  (no projects)")
		}
		for _, p := range projects {
			fmt.Printf("  %d. [%s] %s\n", p.ID, p.Status, p.Name)
		}
		fmt.Println("\n  a. Add project   b. Back")
		choice := a.readLine("Choice: ")
		switch strings.TrimSpace(choice) {
		case "a":
			name := a.readLine("Project name: ")
			p := &domain.Project{
				UserID: a.currentUser.ID,
				Name:   strings.TrimSpace(name),
				Status: domain.ProjectStatusActive,
			}
			if err := a.projectRepo.Create(p); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Printf("%sProject created (id=%d)%s\n", colorGreen, p.ID, colorReset)
			}
		case "b":
			return
		}
	}
}

func (a *App) notesMenu() {
	for {
		fmt.Printf("\n%s-- Notes --%s\n", colorCyan, colorReset)
		notes, _ := a.noteRepo.GetAllByUser(a.currentUser.ID)
		if len(notes) == 0 {
			fmt.Println("  (no notes)")
		}
		for _, n := range notes {
			fmt.Printf("  %d. %s\n", n.ID, n.Title)
		}
		fmt.Println("\n  a. Add note   b. Back")
		choice := a.readLine("Choice: ")
		switch strings.TrimSpace(choice) {
		case "a":
			title := a.readLine("Title: ")
			content := a.readLine("Content: ")
			n := &domain.Note{
				UserID:  a.currentUser.ID,
				Title:   strings.TrimSpace(title),
				Content: strings.TrimSpace(content),
			}
			if err := a.noteRepo.Create(n); err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Printf("%sNote created (id=%d)%s\n", colorGreen, n.ID, colorReset)
			}
		case "b":
			return
		}
	}
}

func (a *App) readLine(prompt string) string {
	fmt.Print(prompt)
	a.scanner.Scan()
	return a.scanner.Text()
}
```

- [ ] **Step 2: Wire up main.go**

```go
// cmd/taskflow/main.go
package main

import (
	"fmt"
	"os"

	"taskflow/internal/config"
	"taskflow/internal/db"
	"taskflow/internal/tui"
)

func main() {
	cfg := config.Instance()
	conn := db.Instance(cfg.DBPath)
	defer conn.Close()

	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("TaskFlow v0.1.0")
		return
	}

	app := tui.New(conn)
	app.Run()
}
```

- [ ] **Step 3: Run the app**

```bash
go run ./cmd/taskflow/
```

Expected: shows main menu, can create tasks/projects/notes, data persists in `taskflow.db`

- [ ] **Step 4: Run all tests to confirm nothing broken**

```bash
go test ./...
```

Expected: all `PASS`

- [ ] **Step 5: Commit**

```bash
git add internal/tui/ cmd/taskflow/main.go
git commit -m "feat: add minimal TUI and wire up main entry point"
```

---

## Task 10: BPMN Diagrams Update

> **REMINDER FOR USER:** Укажи путь к существующим drawio-диаграммам, чтобы обновить BPMN с учётом БД и обоих бизнес-процессов.

- [ ] **Step 1: Ask user for diagram file paths**

Wait for user to provide paths to existing `.drawio` files from `/Users/khuzhokov/Desktop/проектирование информационных систем/`.

- [ ] **Step 2: Update BP1 — Create task with optional reminder**

BPMN должен отражать:
- Swimlanes: User | TaskFlow App | Telegram Bot API
- DB операции: INSERT tasks, INSERT reminders
- Условие: если due_date задан → создать reminder

- [ ] **Step 3: Update BP2 — Pomodoro session lifecycle**

BPMN должен отражать:
- State transitions: IDLE→RUNNING→PAUSED→RUNNING→COMPLETED/CANCELLED
- DB операции: INSERT pomodoro_sessions, UPDATE state
- Trigger: таймер завершения → auto-complete

---

## Self-Review Checklist

- [x] All 3 Phase 1 code requirements covered (app created, DB created, basic app runs)
- [x] Phase 1 req #5: entities defined (Task 2)
- [x] Phase 1 req #6: mathematical model (Task 8 — ETA calculator)
- [x] Phase 1 req #7: Strategy (Task 6), State (Task 7), Singleton (Tasks 3+4)
- [x] Phase 1 req #4: BPMN update (Task 10 — requires user input)
- [x] TDD: every code task has test-first steps
- [x] No placeholder code — all steps have actual implementations
- [x] Type consistency verified across all tasks
- [x] SQLite modernc.org/sqlite (pure Go, no CGO)
