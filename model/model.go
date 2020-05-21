package model

import (
	"net"
	"smotri.me/pkg/utils"
)

type (
	Room struct {
		ID       string  `json:"id"`
		Title    string  `json:"title"`
		VideoURL string  `json:"video_url"`
		Members  []*User `json:"members"`
	}

	User struct {
		ID     string   `json:"id"`
		Name   string   `json:"name"`
		RoomID string   `json:"room_id"`
		Color  string   `json:"color"`
		Conn   net.Conn `json:"-"`
	}
)

func (r *Room) Valid() bool {
	return utils.IsLengthValid(r.Title, 2, 100) && utils.IsUrlValid(r.VideoURL)
}
