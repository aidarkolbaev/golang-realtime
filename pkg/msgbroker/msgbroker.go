package msgbroker

// MessageBroker used for sending and receiving messages
type MessageBroker interface {
	// Publish sends msg to channels
	Publish(msg []byte, channel string) error
	// Subscribe returns Message channel that subscribed to channels
	Subscribe(channel string, cb MessageHandler) error
	// Unsubscribe from the given channels
	Unsubscribe(channels ...string) error
	// Close closes subscriptions
	Close() error
}

// MessageHandler is a callback function that processes messages delivered to subscribers.
type MessageHandler func(msg *Message)

// Message is the representation of transmitted data
type Message struct {
	Channel string
	Data    []byte
}
