package websocket

import (
	"fmt"
	"github.com/gobwas/ws/wsutil"
	"github.com/labstack/gommon/log"
	"io"
	"strings"
	"time"
)

type Request struct {
	ID     string                 `json:"id"`
	UserID string                 `json:"user_id"`
	RoomID string                 `json:"room_id"`
	Method string                 `json:"method"`
	SentAt time.Time              `json:"sent_at"`
	Params map[string]interface{} `json:"params"`
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

type Response struct {
	ID     string                 `json:"id"`
	Result map[string]interface{} `json:"result"`
}

type Pusher struct{}

func (_ Pusher) Push(w io.Writer, msg []byte) {
	err := wsutil.WriteServerText(w, msg)
	if err != nil {
		log.Warn(err)
	}
}
