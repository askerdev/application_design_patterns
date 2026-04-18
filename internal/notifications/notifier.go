package notifications

type MessageSender interface {
	Send(message string) error
	IsConfigured() bool
}
