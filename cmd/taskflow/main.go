package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"taskflow/internal/config"
	"taskflow/internal/db"
	"taskflow/internal/repository"
	"taskflow/internal/telegram"
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

	// Set up Observer pattern: ReminderService (subject) + TelegramNotifier (observer)
	tgClient := telegram.NewClient(cfg.TelegramBotToken, cfg.TelegramChatID)
	reminderRepo := repository.NewReminderRepo(conn)
	reminderService := telegram.NewReminderService(reminderRepo)
	reminderService.SetClient(tgClient)
	reminderService.Register(telegram.NewTelegramNotifier(tgClient))

	// Background goroutine: check reminders every minute
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			reminderService.CheckAndNotify()
		}
	}()

	if _, err := tui.New(conn, reminderService).Run(); err != nil {
		log.Fatal(err)
	}
}
