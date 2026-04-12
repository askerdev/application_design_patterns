package telegram

// MessageSender is the target interface for the Adapter pattern.
// It decouples consumers (e.g. TelegramNotifier) from the concrete *Client.
type MessageSender interface {
	Send(message string) error
}

// ClientAdapter adapts *Client to the MessageSender interface.
type ClientAdapter struct {
	client *Client
}

// NewClientAdapter wraps a *Client and returns it as a MessageSender.
func NewClientAdapter(c *Client) MessageSender {
	return &ClientAdapter{client: c}
}

// Send delegates to the underlying client's SendMessage method.
func (a *ClientAdapter) Send(message string) error {
	return a.client.SendMessage(message)
}
