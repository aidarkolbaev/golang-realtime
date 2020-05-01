package model

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/labstack/gommon/log"
	"net"
	"smotri.me/pkg/utils"
)

type (
	Room struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		MovieURL string `json:"movie_url"`
	}

	User struct {
		Name   string   `json:"name"`
		RoomID string   `json:"room_id"`
		Conn   net.Conn `json:"-"`
	}
)

func (r *Room) Valid() bool {
	return utils.IsLengthValid(r.Title, 2, 100) && utils.IsUrlValid(r.MovieURL)
}

func (u *User) ServeWebsocket() {
	_ = ws.WriteFrame(u.Conn, ws.NewPingFrame([]byte("ping")))
	for {
		b, err := wsutil.ReadClientText(u.Conn)
		if err != nil {
			break
		}

		err = wsutil.WriteServerText(u.Conn, b)
		if err != nil {
			log.Warn(err)
		}
	}

	_ = u.Conn.Close()
}
