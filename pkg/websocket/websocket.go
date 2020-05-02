package websocket

import (
	"github.com/gobwas/ws/wsutil"
	"github.com/labstack/gommon/log"
	"io"
	"smotri.me/model"
	"time"
)

type Request struct {
	ID     string                 `json:"id"`
	User   *model.User            `json:"user"`
	Method string                 `json:"method"`
	Params map[string]interface{} `json:"params"`
	SentAt time.Time              `json:"sent_at"`
}

type Response struct {
	RequestID string `json:"request_id"`
	Success   bool   `json:"success"`
}

type Pusher struct{}

func (_ Pusher) Push(w io.Writer, msg []byte) {
	err := wsutil.WriteServerText(w, msg)
	if err != nil {
		log.Warn(err)
	}
}
