package telegram_test

import (
	"strings"
	"taskflow/internal/app"
	"taskflow/internal/db"
	domain "taskflow/internal/domain"
	"taskflow/internal/notifications/telegram"
	"taskflow/internal/repository/sqlite"
	"testing"
)

type mockBotClient struct {
	sentMessages []string
}

func (m *mockBotClient) SendMessage(text string) error {
	m.sentMessages = append(m.sentMessages, text)
	return nil
}

func (m *mockBotClient) GetUpdates(offset int64) ([]telegram.Update, error) {
	return nil, nil
}

func TestBotListener_HandleMessage_CreateTask(t *testing.T) {
	// Setup DB & Facade
	conn, err := db.OpenMemory()
	if err != nil {
		t.Fatalf("open memory db: %v", err)
	}
	defer conn.Close()
	
	if err := db.RunMigrations(conn); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	conn.Exec("INSERT INTO users (id, username) VALUES (1, 'testuser')")
	
	taskRepo := sqlite.NewTaskRepo(conn)
	projectRepo := sqlite.NewProjectRepo(conn)
	noteRepo := sqlite.NewNoteRepo(conn)
	
	taskSvc := domain.NewTaskService(taskRepo)
	projectSvc := domain.NewProjectService(projectRepo)
	noteSvc := domain.NewNoteService(noteRepo)
	
	user := &domain.User{ID: 1, Username: "testuser"}
	facade := app.NewAppFacade(taskSvc, projectSvc, noteSvc, user)
	
	mockClient := &mockBotClient{}
	chatID := int64(12345)
	listener := telegram.NewBotListener(mockClient, facade, chatID)
	
	// Test command
	listener.HandleMessage("/task Buy milk")
	
	// Assert task created in DB
	tasks, err := taskSvc.List(user.ID)
	if err != nil {
		t.Fatalf("List tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Content != "Buy milk" {
		t.Errorf("expected task content 'Buy milk', got %q", tasks[0].Content)
	}
	
	// Assert confirmation message sent
	if len(mockClient.sentMessages) != 1 {
		t.Fatalf("expected 1 confirmation message, got %d", len(mockClient.sentMessages))
	}
	if !strings.Contains(mockClient.sentMessages[0], "✅ Task created: Buy milk") {
		t.Errorf("unexpected confirmation message: %q", mockClient.sentMessages[0])
	}
}

func TestBotListener_HandleMessage_InvalidCommand(t *testing.T) {
	mockClient := &mockBotClient{}
	listener := telegram.NewBotListener(mockClient, nil, 12345)
	
	listener.HandleMessage("just some text")
	if len(mockClient.sentMessages) != 0 {
		t.Errorf("expected no messages sent for invalid command, got %d", len(mockClient.sentMessages))
	}
}

func TestBotListener_HandleMessage_EmptyTitle(t *testing.T) {
	mockClient := &mockBotClient{}
	listener := telegram.NewBotListener(mockClient, nil, 12345)
	
	listener.HandleMessage("/task ")
	if len(mockClient.sentMessages) != 1 {
		t.Fatalf("expected 1 message sent, got %d", len(mockClient.sentMessages))
	}
	if !strings.Contains(mockClient.sentMessages[0], "Usage: /task <title>") {
		t.Errorf("expected usage message, got %q", mockClient.sentMessages[0])
	}
}
