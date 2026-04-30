package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

// Config — конфигурация приложения. Загружается из ~/.config/taskflow/config.json,
// ENV-переменные имеют приоритет над файлом.
type Config struct {
	TelegramBotToken     string `json:"telegram_bot_token"`
	TelegramChatID       int64  `json:"telegram_chat_id"`
	NotificationsEnabled bool   `json:"notifications_enabled"`
	DBPath               string `json:"-"`
}

var (
	instance *Config
	once     sync.Once
)

// Path возвращает путь к JSON-файлу конфига.
func Path() string {
	if p := os.Getenv("TASKFLOW_CONFIG"); p != "" {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "taskflow_config.json"
	}
	return filepath.Join(home, ".config", "taskflow", "config.json")
}

func loadFile() Config {
	var fc Config
	data, err := os.ReadFile(Path())
	if err != nil {
		return fc
	}
	_ = json.Unmarshal(data, &fc)
	return fc
}

func Instance() *Config {
	once.Do(func() {
		fc := loadFile()

		if t := os.Getenv("TELEGRAM_BOT_TOKEN"); t != "" {
			fc.TelegramBotToken = t
		}
		if v := os.Getenv("TELEGRAM_CHAT_ID"); v != "" {
			id, _ := strconv.ParseInt(v, 10, 64)
			fc.TelegramChatID = id
		}

		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			dbPath = "taskflow.db"
		}
		fc.DBPath = dbPath

		instance = &fc
	})
	return instance
}

// Save сохраняет текущий инстанс в JSON-файл (без DBPath — он только из ENV).
func Save(cfg *Config) error {
	path := Path()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func ResetForTesting() {
	instance = nil
	once = sync.Once{}
}
