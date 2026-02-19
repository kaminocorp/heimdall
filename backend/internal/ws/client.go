package ws

// Client represents an individual WebSocket connection.
type Client struct {
	ID   string
	send chan Message
}

func NewClient(id string) *Client {
	return &Client{
		ID:   id,
		send: make(chan Message, 64),
	}
}

func (c *Client) Send(msg Message) {
	select {
	case c.send <- msg:
	default:
		// drop message if buffer full
	}
}

func (c *Client) Messages() <-chan Message {
	return c.send
}
