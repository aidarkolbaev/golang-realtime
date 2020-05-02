package model

import (
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
		ID     string   `json:"id"`
		Name   string   `json:"name"`
		RoomID string   `json:"room_id"`
		Conn   net.Conn `json:"-"`
	}
)

func (u *User) GetID() string {
	return u.ID
}

func (u *User) Write(p []byte) (int, error) {
	return u.Conn.Write(p)
}

func (r *Room) Valid() bool {
	return utils.IsLengthValid(r.Title, 2, 100) && utils.IsUrlValid(r.MovieURL)
}
