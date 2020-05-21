package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"smotri.me/model"
	"smotri.me/pkg/utils"
	"time"
)

type Storage interface {
	TempRoomExist(roomID string) bool
	CreateTempRoom(room *model.Room, exp time.Duration) (ID string, err error)
	GetTempRoom(roomID string) (*model.Room, error)
	UpdateTempRoom(room *model.Room) error
	AddUserToRoom(roomID string, u *model.User) error
	UpdateRoomUser(roomID string, u *model.User) error
	RemoveUserFromRoom(roomID string, userID string) error
	IncrVisits() (int64, error)
	GetVisitsByDate(date time.Time) (int64, error)
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
		"video_url": room.VideoURL,
	}

	affectedFields := s.rdb.HSet("room:"+ID, data).Val()
	if affectedFields != 3 {
		return "", fmt.Errorf("invalid affected fields num: %d", affectedFields)
	}
	ok := s.rdb.Expire("room:"+ID, exp).Val()
	if !ok {
		return "", fmt.Errorf("timeout was not set, key '%s' does not exist", ID)
	}
	return ID, nil
}

func (s *storage) GetTempRoom(roomID string) (*model.Room, error) {
	var r model.Room
	data := s.rdb.HGetAll("room:" + roomID).Val()
	if len(data) == 0 {
		return nil, fmt.Errorf("room '%s' not found", roomID)
	}

	membersJSON, exists := data["members"]
	if exists {
		err := json.Unmarshal([]byte(membersJSON), &r.Members)
		if err != nil {
			return nil, err
		}
	} else {
		r.Members = []*model.User{}
	}

	r.ID = data["id"]
	r.Title = data["title"]
	r.VideoURL = data["video_url"]
	return &r, nil
}

func (s *storage) UpdateTempRoom(room *model.Room) error {
	if room.ID == "" {
		return fmt.Errorf("invalid room id: %s", room.ID)
	}

	data := map[string]interface{}{
		"title":     room.Title,
		"video_url": room.VideoURL,
	}

	_ = s.rdb.HSet("room:"+room.ID, data).Val()
	return nil
}

func (s *storage) AddUserToRoom(roomID string, u *model.User) error {
	room, err := s.GetTempRoom(roomID)
	if err != nil {
		return err
	}
	for _, member := range room.Members {
		if member.ID == u.ID {
			return fmt.Errorf("member with ID:%s already exists", member.ID)
		}
	}
	room.Members = append(room.Members, u)
	membersJSON, err := json.Marshal(room.Members)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"members": string(membersJSON),
	}
	_ = s.rdb.HSet("room:"+room.ID, data).Val()
	return nil
}

func (s *storage) UpdateRoomUser(roomID string, u *model.User) error {
	room, err := s.GetTempRoom(roomID)
	if err != nil {
		return err
	}
	for idx, member := range room.Members {
		if member.ID == u.ID {
			room.Members[idx] = u
			break
		}
	}

	membersJSON, err := json.Marshal(room.Members)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"members": string(membersJSON),
	}
	_ = s.rdb.HSet("room:"+room.ID, data).Val()
	return nil
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
		}
	}

	membersJSON, err := json.Marshal(room.Members)
	if err != nil {
		return err
	}

	data := map[string]interface{}{
		"members": string(membersJSON),
	}
	_ = s.rdb.HSet("room:"+room.ID, data).Val()
	return nil
}

func (s *storage) IncrVisits() (int64, error) {
	return s.rdb.Incr("visits:" + time.Now().Format("02.01.06")).Result()
}

func (s *storage) GetVisitsByDate(date time.Time) (int64, error) {
	return s.rdb.Get("visits:" + date.Format("02.01.06")).Int64()
}

func (s *storage) TempRoomExist(roomID string) bool {
	return s.rdb.Exists("room:"+roomID).Val() == 1
}
