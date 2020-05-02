package storage

import (
	"errors"
	"github.com/go-redis/redis/v7"
	"smotri.me/model"
	"smotri.me/pkg/utils"
	"time"
)

type Storage interface {
	CreateTempRoom(room *model.Room, exp time.Duration) (ID string, err error)
	GetTempRoom(roomID string) (*model.Room, error)
	IncrVisits() (int64, error)
	RoomExist(roomID string) bool
}

type storage struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) Storage {
	return &storage{rdb: rdb}
}

func (s *storage) CreateTempRoom(room *model.Room, exp time.Duration) (string, error) {
	var ID string
	for i := 3; i <= 20; i++ {
		newID := utils.RandString(i)
		if !s.RoomExist(newID) {
			ID = newID
			break
		}
	}

	if ID == "" {
		return "", errors.New("unable to generate a unique ID")
	}

	data := map[string]interface{}{
		"id":        ID,
		"title":     room.Title,
		"movie_url": room.MovieURL,
	}

	affectedFields := s.rdb.HSet("room:"+ID, data).Val()
	if affectedFields != 3 {
		return "", errors.New("invalid affected fields num")
	}
	ok := s.rdb.Expire("room:"+ID, exp).Val()
	if !ok {
		return "", errors.New("timeout was not set, key " + ID + " does not exist")
	}
	return ID, nil
}

func (s *storage) GetTempRoom(roomID string) (*model.Room, error) {
	var r model.Room
	data := s.rdb.HGetAll("room:" + roomID).Val()
	if len(data) == 0 {
		return nil, errors.New("room " + roomID + " not found")
	}
	r.ID = data["id"]
	r.Title = data["title"]
	r.MovieURL = data["movie_url"]
	return &r, nil
}

func (s *storage) IncrVisits() (int64, error) {
	return s.rdb.Incr("visits:" + time.Now().Format("02.01.06")).Result()
}

func (s *storage) RoomExist(roomID string) bool {
	return s.rdb.Exists("room:"+roomID).Val() == 1
}
