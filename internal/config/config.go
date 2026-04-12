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

// ResetForTesting resets the singleton — only for use in tests
func ResetForTesting() {
	instance = nil
	once = sync.Once{}
}
