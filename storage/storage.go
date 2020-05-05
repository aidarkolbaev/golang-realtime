package storage

import (
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v7"
	"smotri.me/model"
	"smotri.me/pkg/utils"
	"time"
)

type Storage interface {
	CreateTempRoom(room *model.Room, exp time.Duration) (ID string, err error)
	GetTempRoom(roomID string) (*model.Room, error)
	UpdateTempRoom(room *model.Room) error
	AddUserToRoom(roomID string, u *model.User) error
	RemoveUserFromRoom(roomID string, userID string) error
	IncrVisits() (int64, error)
	TempRoomExist(roomID string) bool
}

type storage struct {
	rdb *redis.Client
}

func New(rdb *redis.Client) Storage {
	return &storage{rdb: rdb}
}

func (s *storage) CreateTempRoom(room *model.Room, exp time.Duration) (string, error) {
	var ID string
	for i := 5; i <= 15; i++ {
		newID := utils.RandString(i)
		if !s.TempRoomExist(newID) {
			ID = newID
			break
		}
	}

	if ID == "" {
		return "", errors.New("unable to generate an unique ID")
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

	membersJSON, exists := data["members"]
	if exists {
		err := json.Unmarshal([]byte(membersJSON), &r.Members)
		if err != nil {
			return nil, err
		}
	}

	r.ID = data["id"]
	r.Title = data["title"]
	r.MovieURL = data["movie_url"]
	return &r, nil
}

func (s *storage) UpdateTempRoom(room *model.Room) error {
	if room.ID == "" {
		return errors.New("room id is required")
	}

	membersJSON, err := json.Marshal(room.Members)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"id":        room.ID,
		"title":     room.Title,
		"movie_url": room.MovieURL,
		"members":   string(membersJSON),
	}

	affectedFields := s.rdb.HSet("room:"+room.ID, data).Val()
	if affectedFields != 4 {
		return errors.New("invalid affected fields num")
	}
	return nil
}

func (s *storage) AddUserToRoom(roomID string, u *model.User) error {
	room, err := s.GetTempRoom(roomID)
	if err != nil {
		return err
	}
	room.Members = append(room.Members, u)
	return s.UpdateTempRoom(room)
}

func (s *storage) RemoveUserFromRoom(roomID string, userID string) error {
	room, err := s.GetTempRoom(roomID)
	if err != nil {
		return err
	}
	for i, u := range room.Members {
		if u.ID == userID {
			lastElem := len(room.Members) - 1
			room.Members[i] = room.Members[lastElem]
			room.Members[lastElem] = nil
			room.Members = room.Members[:lastElem]
			return s.UpdateTempRoom(room)
		}
	}
	return nil
}

func (s *storage) IncrVisits() (int64, error) {
	return s.rdb.Incr("visits:" + time.Now().Format("02.01.06")).Result()
}

func (s *storage) TempRoomExist(roomID string) bool {
	return s.rdb.Exists("room:"+roomID).Val() == 1
}
