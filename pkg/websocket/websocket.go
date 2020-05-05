package websocket

import (
	"fmt"
	"smotri.me/model"
	"strings"
	"sync"
	"time"
)

type Channels interface {
	Subscribe(u *model.User, channels ...string)
	Unsubscribe(u *model.User, channels ...string)
	GetSubscribers(channel string) []*model.User
}

type (
	channels struct {
		sync.Mutex
		storage map[string]map[string]*model.User
	}

	Message struct {
		ID     string                 `json:"id"`
		UserID string                 `json:"user_id"`
		RoomID string                 `json:"room_id"`
		Method string                 `json:"method"`
		SentAt time.Time              `json:"sent_at"`
		Params map[string]interface{} `json:"params"`
	}
)

func NewChannels() Channels {
	return &channels{
		storage: make(map[string]map[string]*model.User),
	}
}

func (h *channels) Subscribe(u *model.User, channels ...string) {
	h.Lock()
	for _, ch := range channels {
		_, exists := h.storage[ch]
		if !exists {
			h.storage[ch] = make(map[string]*model.User)
		}
		h.storage[ch][u.ID] = u
	}
	h.Unlock()
}

func (h *channels) Unsubscribe(u *model.User, channels ...string) {
	h.Lock()
	for _, ch := range channels {
		_, exists := h.storage[ch]
		if exists {
			delete(h.storage[ch], u.ID)
		}
	}
	h.Unlock()
}

func (h *channels) GetSubscribers(channel string) []*model.User {
	var result []*model.User
	h.Lock()
	subscribers, channelExists := h.storage[channel]
	h.Unlock()
	if channelExists {
		for _, s := range subscribers {
			result = append(result, s)
		}
	}
	return result
}

func (m *Message) Validate() error {
	if strings.TrimSpace(m.ID) == "" {
		return fmt.Errorf("invalid request id")
	}

	if strings.TrimSpace(m.RoomID) == "" {
		return fmt.Errorf("invalid room id")
	}

	switch m.Method {
	case "new_message":
		content, ok := m.Params["content"].(string)
		if !ok || strings.TrimSpace(content) == "" {
			return fmt.Errorf("invalid '%s' request, param 'content' is required and must be string", m.Method)
		}
	case "edit_message":
		_, ok := m.Params["message_id"].(int64)
		if !ok {
			return fmt.Errorf("invalid '%s' request, param 'message_id' is required and must be int", m.Method)
		}

		content, ok := m.Params["content"].(string)
		if !ok || strings.TrimSpace(content) == "" {
			return fmt.Errorf("invalid '%s' request, param 'content' is required and must be string", m.Method)
		}
	case "remove_message":
		_, ok := m.Params["message_id"].(int64)
		if !ok {
			return fmt.Errorf("invalid '%s' request, param 'message_id' is required and must be int", m.Method)
		}
	default:
		return fmt.Errorf("invalid request method: '%s'", m.Method)
	}

	return nil
}
