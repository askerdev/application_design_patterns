package telegram

import notifications "taskflow/internal/notifications"

// ClientAdapter adapts *Client to the notifications.MessageSender interface.
type ClientAdapter struct {
	client *Client
}

// NewClientAdapter wraps a *Client and returns it as a notifications.MessageSender.
func NewClientAdapter(c *Client) notifications.MessageSender {
	return &ClientAdapter{client: c}
}

func (a *ClientAdapter) Send(message string) error {
	return a.client.SendMessage(message)
}

func (a *ClientAdapter) IsConfigured() bool {
	return a.client.IsConfigured()
}
