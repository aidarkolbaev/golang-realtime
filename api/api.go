package api

import (
	"encoding/json"
	"github.com/gammazero/workerpool"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"net/http"
	"smotri.me/config"
	"smotri.me/model"
	"smotri.me/pkg/msgbroker"
	"smotri.me/pkg/utils"
	"smotri.me/pkg/websocket"
	"smotri.me/storage"
	"strconv"
	"time"
)

type API struct {
	echo       *echo.Echo
	config     *config.Config
	storage    storage.Storage
	msgBroker  msgbroker.MessageBroker
	workerPool *workerpool.WorkerPool
	channels   websocket.Channels
}

func New(c *config.Config, s storage.Storage, mb msgbroker.MessageBroker) *API {
	api := &API{
		echo:       echo.New(),
		config:     c,
		storage:    s,
		msgBroker:  mb,
		workerPool: workerpool.New(c.MaxWorkers),
		channels:   websocket.NewChannels(),
	}

	api.echo.HideBanner = true
	api.echo.Use(middleware.CORS())

	api.echo.GET("/", api.ping)
	api.echo.POST("/room", api.createRoom)
	api.echo.GET("/room/:roomID", api.getRoom)
	api.echo.Any("/ws", api.websocketHandler)

	return api
}

func (api *API) Start() error {
	err := api.msgBroker.Subscribe("messages:*", api.handleMessages)
	if err != nil {
		return err
	}
	return api.echo.Start(":" + strconv.Itoa(api.config.HttpPort))
}

func (api *API) Close() error {
	api.workerPool.StopWait()
	return api.echo.Close()
}

// Ping handler
func (api *API) ping(c echo.Context) error {
	_, err := api.storage.IncrVisits()
	if err != nil {
		log.Error(err)
	}
	return c.String(http.StatusOK, "OK")
}

// Room creation endpoint
func (api *API) createRoom(c echo.Context) error {
	var room model.Room
	err := c.Bind(&room)
	if err != nil || !room.Valid() {
		if err != nil {
			log.Warn(err)
		}
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}

	room.ID, err = api.storage.CreateTempRoom(&room, time.Minute*15)
	if err != nil {
		log.Error(err)
		return echo.NewHTTPError(http.StatusConflict)
	}

	return c.JSON(http.StatusOK, &room)
}

// Returns room data by roomID
func (api *API) getRoom(c echo.Context) error {
	roomID := c.Param("roomID")
	room, err := api.storage.GetTempRoom(roomID)
	if err != nil {
		log.Info(err)
		return echo.NewHTTPError(http.StatusNotFound)
	}
	return c.JSON(http.StatusOK, room)
}

// Endpoint to establish websocketHandler connection
func (api *API) websocketHandler(c echo.Context) error {
	username := c.QueryParam("username")
	roomID := c.QueryParam("room_id")
	if !api.storage.TempRoomExist(roomID) {
		return c.NoContent(http.StatusNotFound)
	}

	if !utils.IsNameValid(username) {
		return c.NoContent(http.StatusUnprocessableEntity)
	}

	conn, _, _, err := ws.UpgradeHTTP(c.Request(), c.Response())
	if err != nil {
		log.Warn(err)
		return c.NoContent(http.StatusBadRequest)
	}

	user := &model.User{
		ID:     roomID + utils.RandString(5),
		Name:   username,
		RoomID: roomID,
		Conn:   conn,
	}
	api.serveUser(user)
	return nil
}

// Serves user websocketHandler connection
func (api *API) serveUser(u *model.User) {
	done := make(chan bool)

	onConnect := func() {
		api.channels.Subscribe(u, u.RoomID)

		err := api.storage.AddUserToRoom(u.RoomID, u)
		if err != nil {
			log.Error(err)
		}

		ticker := time.NewTicker(time.Second * 30)
		defer ticker.Stop()

		for {
			select {
			case <-done:
				log.Info("ticker stop")
				return
			case <-ticker.C:
				err := wsutil.WriteServerMessage(u.Conn, ws.OpPing, []byte("ping"))
				if err != nil {
					log.Warn(err)
				}
			}
		}
	}

	onDisconnect := func() {
		done <- true
		_ = u.Conn.Close()
		api.channels.Unsubscribe(u, u.RoomID)

		err := api.storage.RemoveUserFromRoom(u.RoomID, u.ID)
		if err != nil {
			log.Error(err)
		}
		log.Infof("user %s disconnected from room %s", u.Name, u.RoomID)
	}

	go onConnect()
	defer onDisconnect()

	for {
		b, err := wsutil.ReadClientText(u.Conn)
		if err != nil {
			break
		}

		var req websocket.Message
		err = json.Unmarshal(b, &req)
		if err != nil {
			continue
		}

		if err = req.Validate(); err != nil {
			log.Warn(err)
			continue
		}

		req.UserID = u.ID
		req.RoomID = u.RoomID
		req.SentAt = time.Now()
		b, err = json.Marshal(&req)
		if err != nil {
			log.Error(err)
			continue
		}

		err = api.msgBroker.Publish(b, "messages:"+req.RoomID)
		if err != nil {
			log.Warn(err)
		}
	}
}

// Message handler
func (api *API) handleMessages(msg *msgbroker.Message) {
	api.workerPool.Submit(func() {
		log.Info(msg.Channel)
		if len(msg.Channel) > len("messages:") {
			roomID := msg.Channel[len("messages:"):]
			log.Info("roomID:" + roomID)
			users := api.channels.GetSubscribers(roomID)
			for _, u := range users {
				err := wsutil.WriteServerText(u.Conn, msg.Data)
				if err != nil {
					log.Warn(err)
				}
			}
		}

	})
}
