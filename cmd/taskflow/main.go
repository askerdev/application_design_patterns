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
