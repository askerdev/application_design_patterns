package notifications

// MessageSender is the target interface for the Adapter pattern.
// Decouples notification consumers from the concrete Telegram client.
type MessageSender interface {
	Send(message string) error
	IsConfigured() bool
}
