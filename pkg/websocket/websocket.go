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

	Request struct {
		ID     string                 `json:"id"`
		UserID string                 `json:"user_id"`
		RoomID string                 `json:"room_id"`
		Method string                 `json:"method"`
		SentAt time.Time              `json:"sent_at"`
		Params map[string]interface{} `json:"params"`
	}

	Response struct {
		ID     string                 `json:"id"`
		Result map[string]interface{} `json:"result"`
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

func (r *Request) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return fmt.Errorf("invalid request id")
	}

	if strings.TrimSpace(r.RoomID) == "" {
		return fmt.Errorf("invalid room id")
	}

	switch r.Method {
	case "new_message":
		content, ok := r.Params["content"].(string)
		if !ok || strings.TrimSpace(content) == "" {
			return fmt.Errorf("invalid '%s' request, param 'content' is required and must be string", r.Method)
		}
	case "edit_message":
		_, ok := r.Params["message_id"].(int64)
		if !ok {
			return fmt.Errorf("invalid '%s' request, param 'message_id' is required and must be int", r.Method)
		}

		content, ok := r.Params["content"].(string)
		if !ok || strings.TrimSpace(content) == "" {
			return fmt.Errorf("invalid '%s' request, param 'content' is required and must be string", r.Method)
		}
	case "remove_message":
		_, ok := r.Params["message_id"].(int64)
		if !ok {
			return fmt.Errorf("invalid '%s' request, param 'message_id' is required and must be int", r.Method)
		}
	default:
		return fmt.Errorf("invalid request method: '%s'", r.Method)
	}

	return nil
}
