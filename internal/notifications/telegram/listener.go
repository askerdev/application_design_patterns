package telegram

import (
	"strings"
	"taskflow/internal/app"
)

type BotClient interface {
	SendMessage(text string) error
	GetUpdates(offset int64) ([]Update, error)
}

type BotListener struct {
	client BotClient
	facade *app.AppFacade
	chatID int64
}

func NewBotListener(client BotClient, facade *app.AppFacade, chatID int64) *BotListener {
	return &BotListener{
		client: client,
		facade: facade,
		chatID: chatID,
	}
}

func (l *BotListener) Start() {
	var offset int64
	for {
		updates, err := l.client.GetUpdates(offset)
		if err != nil {
			continue
		}

		for _, u := range updates {
			if u.UpdateID >= offset {
				offset = u.UpdateID + 1
			}

			if u.Message == nil || u.Message.Chat.ID != l.chatID {
				continue
			}

			l.HandleMessage(u.Message.Text)
		}
	}
}

func (l *BotListener) HandleMessage(text string) {
	if !strings.HasPrefix(text, "/task ") {
		return
	}

	content := strings.TrimSpace(strings.TrimPrefix(text, "/task "))
	if content == "" {
		_ = l.client.SendMessage("Usage: /task <title>")
		return
	}

	task, err := l.facade.AddTask(content, "MEDIUM")
	if err != nil {
		_ = l.client.SendMessage("❌ Error creating task: " + err.Error())
		return
	}

	_ = l.client.SendMessage("✅ Task created: " + task.Content)
}
