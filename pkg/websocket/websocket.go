package websocket

import (
	"fmt"
	"smotri.me/model"
	"smotri.me/pkg/utils"
	"strings"
	"sync"
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
		ID       string                 `json:"id,omitempty"`
		UserID   string                 `json:"user_id"`
		Method   string                 `json:"method"`
		Params   map[string]interface{} `json:"params,omitempty"`
		Response bool                   `json:"-"`
	}
)

func NewMessage() *Message {
	return &Message{
		Params: make(map[string]interface{}),
	}
}

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
	switch m.Method {
	case "new_message":
		content, ok := m.Params["content"].(string)
		if !ok || strings.TrimSpace(content) == "" {
			return fmt.Errorf("invalid '%s' request, param 'content' is required and must be string", m.Method)
		}
	case "edit_message":
		_, ok := m.Params["message_id"].(string)
		if !ok {
			return fmt.Errorf("invalid '%s' request, param 'message_id' is required and must be int", m.Method)
		}

		content, ok := m.Params["content"].(string)
		if !ok || strings.TrimSpace(content) == "" {
			return fmt.Errorf("invalid '%s' request, param 'content' is required and must be string", m.Method)
		}
	case "remove_message":
		_, ok := m.Params["message_id"].(string)
		if !ok {
			return fmt.Errorf("invalid '%s' request, param 'message_id' is required and must be int", m.Method)
		}
	case "rename_member":
		name, ok := m.Params["name"].(string)
		if !ok {
			return fmt.Errorf("invalid '%s' request, param 'name' is required and must be string", m.Method)
		}
		if !utils.IsNameValid(name) {
			return fmt.Errorf("invalid '%s' request, param 'name' is invalid", m.Method)
		}
	case "update_room":
		title, ok := m.Params["title"].(string)
		if !ok {
			return fmt.Errorf("invalid '%s' request, param 'title' is required and must be string", m.Method)
		}
		if !utils.IsLengthValid(title, 2, 100) {
			return fmt.Errorf("invalid '%s' request, param 'title' is invalid", m.Method)
		}

		videoURL, ok := m.Params["video_url"].(string)
		if !ok {
			return fmt.Errorf("invalid '%s' request, param 'video_url' is required and must be string", m.Method)
		}
		if !utils.IsUrlValid(videoURL) {
			return fmt.Errorf("invalid '%s' request, param 'video_url' is invalid", m.Method)
		}
	case "video_sync", "video_play", "video_pause":
	case "get_members", "get_me":
	default:
		return fmt.Errorf("invalid request method: '%s'", m.Method)
	}

	return nil
}
