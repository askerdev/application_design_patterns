package config

import (
	"os"
	"strconv"
	"sync"
)

type Config struct {
	TelegramBotToken string
	TelegramChatID   int64
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
		chatID, _ := strconv.ParseInt(os.Getenv("TELEGRAM_CHAT_ID"), 10, 64)
		instance = &Config{
			TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
			TelegramChatID:   chatID,
			DBPath:           dbPath,
		}
	})
	return instance
}

func ResetForTesting() {
	instance = nil
	once = sync.Once{}
}
