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
