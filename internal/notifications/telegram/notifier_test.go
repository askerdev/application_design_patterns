package telegram_test

import (
	"errors"
	"testing"

	domain "taskflow/internal/domain"
	"taskflow/internal/notifications/telegram"
)

type stubSender struct {
	sent        []string
	errToReturn error
	configured  bool
}

func (s *stubSender) Send(message string) error {
	s.sent = append(s.sent, message)
	return s.errToReturn
}

func (s *stubSender) IsConfigured() bool {
	return s.configured
}

func TestTelegramNotifier_Notify_SendsFormattedMessage(t *testing.T) {
	stub := &stubSender{configured: true}
	notifier := telegram.NewTelegramNotifier(stub)

	r := &domain.Reminder{Content: "Stand-up meeting"}
	if err := notifier.Notify(r); err != nil {
		t.Fatalf("Notify returned unexpected error: %v", err)
	}

	if len(stub.sent) != 1 {
		t.Fatalf("expected 1 message sent, got %d", len(stub.sent))
	}
	want := "⏰ Reminder: Stand-up meeting"
	if stub.sent[0] != want {
		t.Errorf("message mismatch\n got:  %q\n want: %q", stub.sent[0], want)
	}
}

func TestTelegramNotifier_Notify_PropagatesSenderError(t *testing.T) {
	sendErr := errors.New("network timeout")
	stub := &stubSender{configured: true, errToReturn: sendErr}
	notifier := telegram.NewTelegramNotifier(stub)

	r := &domain.Reminder{Content: "Deploy prod"}
	err := notifier.Notify(r)
	if !errors.Is(err, sendErr) {
		t.Errorf("expected error %v, got %v", sendErr, err)
	}
}
