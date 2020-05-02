package msgbroker

import (
	"errors"
	"github.com/go-redis/redis/v7"
	"sync"
)

// redisBroker is the implementation of MessageBroker using Redis
type redisBroker struct {
	client *redis.Client
	pubSub *redis.PubSub
	sync.RWMutex
	handlers map[string]MessageHandler
}

// NewRedisBroker returns a implementation of MessageBroker using Redis
func NewRedisBroker(r *redis.Client) MessageBroker {
	rb := &redisBroker{
		client:   r,
		pubSub:   r.Subscribe(),
		handlers: make(map[string]MessageHandler),
	}
	go rb.serveMessages()
	return rb
}

func (rb *redisBroker) serveMessages() {
	for msg := range rb.pubSub.Channel() {
		go func() {
			rb.RLock()
			handler, exists := rb.handlers[msg.Channel]
			rb.RUnlock()
			if exists {
				handler(&Message{
					Channel: msg.Channel,
					Data:    []byte(msg.Payload),
				})
			}
		}()
	}
}

func (rb *redisBroker) Close() error {
	return rb.pubSub.Close()
}

func (rb *redisBroker) Publish(msg []byte, channel string) error {
	if rb.client.Publish(channel, string(msg)).Val() == 0 {
		return errors.New("no recipients")
	}
	return nil
}

func (rb *redisBroker) Subscribe(channel string, cb MessageHandler) error {
	err := rb.pubSub.Subscribe(channel)
	if err != nil {
		return err
	}
	_, err = rb.pubSub.Receive()
	if err != nil {
		return err
	}
	rb.Lock()
	rb.handlers[channel] = cb
	rb.Unlock()
	return nil
}

func (rb *redisBroker) Unsubscribe(channels ...string) error {
	if len(channels) > 0 {
		rb.Lock()
		for _, ch := range channels {
			delete(rb.handlers, ch)
		}
		rb.Unlock()
		return rb.pubSub.Unsubscribe(channels...)
	}
	return nil
}
