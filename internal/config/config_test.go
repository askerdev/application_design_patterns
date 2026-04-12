package config_test

import (
	"os"
	"testing"

	"taskflow/internal/config"
)

func TestConfigReadsTelegramToken(t *testing.T) {
	config.ResetForTesting()
	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token-123")
	defer func() {
		os.Unsetenv("TELEGRAM_BOT_TOKEN")
		config.ResetForTesting()
	}()

	cfg := config.Instance()
	if cfg.TelegramBotToken != "test-token-123" {
		t.Errorf("expected token 'test-token-123', got '%s'", cfg.TelegramBotToken)
	}
}

func TestConfigDefaultDBPath(t *testing.T) {
	config.ResetForTesting()
	os.Unsetenv("DB_PATH")
	defer config.ResetForTesting()

	cfg := config.Instance()
	if cfg.DBPath != "taskflow.db" {
		t.Errorf("expected default DBPath 'taskflow.db', got '%s'", cfg.DBPath)
	}
}

func TestConfigCustomDBPath(t *testing.T) {
	config.ResetForTesting()
	os.Setenv("DB_PATH", "/tmp/test.db")
	defer func() {
		os.Unsetenv("DB_PATH")
		config.ResetForTesting()
	}()

	cfg := config.Instance()
	if cfg.DBPath != "/tmp/test.db" {
		t.Errorf("expected DBPath '/tmp/test.db', got '%s'", cfg.DBPath)
	}
}

func TestConfigSingleton(t *testing.T) {
	config.ResetForTesting()
	defer config.ResetForTesting()

	a := config.Instance()
	b := config.Instance()
	if a != b {
		t.Error("expected Instance() to return the same pointer each time")
	}
}
