package telegram

import notifications "taskflow/internal/notifications"

type ClientAdapter struct {
	client *Client
}

func NewClientAdapter(c *Client) notifications.MessageSender {
	return &ClientAdapter{client: c}
}

func (a *ClientAdapter) Send(message string) error {
	return a.client.SendMessage(message)
}

func (a *ClientAdapter) IsConfigured() bool {
	return a.client.IsConfigured()
}
