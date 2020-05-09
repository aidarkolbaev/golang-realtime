package api

import (
	"context"
	"encoding/json"
	"github.com/gammazero/workerpool"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"net/http"
	"net/url"
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
	api.echo.HidePort = true
	api.echo.Use(middleware.CORS())

	api.echo.GET("/", api.ping)
	api.echo.GET("/visits", api.getVisits)
	api.echo.POST("/room", api.createRoom)
	api.echo.GET("/room/:roomID", api.getRoom)
	api.echo.Any("/ws", api.websocketHandler)

	return api
}

// Starts server
func (api *API) Start() error {
	err := api.msgBroker.Subscribe("messages:*", api.handleMessages)
	if err != nil {
		return err
	}
	log.Infof("server started at port %d", api.config.HttpPort)
	return api.echo.Start(":" + strconv.Itoa(api.config.HttpPort))
}

// Closes server
func (api *API) Close(ctx context.Context) error {
	api.workerPool.StopWait()
	return api.echo.Shutdown(ctx)
}

// Ping handler
func (api *API) ping(c echo.Context) error {
	_, err := api.storage.IncrVisits()
	if err != nil {
		log.Error(err)
	}
	return c.String(http.StatusOK, "OK")
}

// Returns visits count by date
func (api *API) getVisits(c echo.Context) error {
	d, err := url.QueryUnescape(c.QueryParam("date"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	date, err := time.Parse("02.01.06", d)
	if err != nil {
		log.Info(err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	visits, err := api.storage.GetVisitsByDate(date)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}
	return c.JSON(http.StatusOK, map[string]int64{"visits": visits})
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

	room.ID, err = api.storage.CreateTempRoom(&room, time.Hour*24)
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
	api.userConnect(user)
	api.serveUser(user)
	api.userDisconnect(user)
	return nil
}

// Serves user websocket connection
func (api *API) serveUser(u *model.User) {
	done := make(chan bool)

	go func() {
		ticker := time.NewTicker(time.Second * 30)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				err := wsutil.WriteServerMessage(u.Conn, ws.OpPing, []byte("ping"))
				if err != nil {
					log.Warn(err)
				}
			}
		}
	}()

	for {
		b, err := wsutil.ReadClientText(u.Conn)
		if err != nil {
			done <- true
			break
		}

		var req websocket.Message
		err = json.Unmarshal(b, &req)
		if err != nil {
			log.Warn(err)
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

// Websocket connect handler
func (api *API) userConnect(u *model.User) {
	api.channels.Subscribe(u, u.RoomID)

	err := api.storage.AddUserToRoom(u.RoomID, u)
	if err != nil {
		log.Error(err)
	}

	b, err := json.Marshal(&websocket.Message{
		UserID: u.ID,
		RoomID: u.RoomID,
		Method: "user_connect",
		SentAt: time.Now(),
		Params: map[string]interface{}{
			"name": u.Name,
		},
	})

	if err != nil {
		log.Error(err)
	} else {
		err = api.msgBroker.Publish(b, "messages:"+u.RoomID)
		if err != nil {
			log.Error(err)
		}
	}
}

// Websocket disconnect handler
func (api *API) userDisconnect(u *model.User) {
	_ = u.Conn.Close()
	api.channels.Unsubscribe(u, u.RoomID)

	err := api.storage.RemoveUserFromRoom(u.RoomID, u.ID)
	if err != nil {
		log.Error(err)
	}

	b, err := json.Marshal(&websocket.Message{
		UserID: u.ID,
		RoomID: u.RoomID,
		Method: "user_disconnect",
		SentAt: time.Now(),
	})
	if err != nil {
		log.Error(err)
	} else {
		err = api.msgBroker.Publish(b, "messages:"+u.RoomID)
		if err != nil {
			log.Error(err)
		}
	}

	log.Infof("user '%s' disconnected from room '%s'", u.Name, u.RoomID)
}

// Message broker messages handler
func (api *API) handleMessages(msg *msgbroker.Message) {
	api.workerPool.Submit(func() {
		if len(msg.Channel) > len("messages:") {
			roomID := msg.Channel[len("messages:"):]
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
