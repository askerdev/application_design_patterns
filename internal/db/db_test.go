package db_test

import (
	"testing"

	"taskflow/internal/db"
)

func TestMigrationsCreateTables(t *testing.T) {
	conn, err := db.OpenMemory()
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	defer conn.Close()

	if err := db.RunMigrations(conn); err != nil {
		t.Fatalf("migrations failed: %v", err)
	}

	tables := []string{"users", "projects", "tags", "tasks", "notes", "reminders", "pomodoro_sessions"}
	for _, table := range tables {
		var name string
		row := conn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table)
		if err := row.Scan(&name); err != nil {
			t.Errorf("table %q should exist but got error: %v", table, err)
		}
	}
}

func TestDBSingleton(t *testing.T) {
	db.ResetForTesting()
	defer db.ResetForTesting()

	a := db.Instance(":memory:")
	b := db.Instance(":memory:")
	if a != b {
		t.Error("expected Instance() to return the same pointer each time")
	}
}
