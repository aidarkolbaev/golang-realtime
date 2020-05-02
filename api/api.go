package api

import (
	"encoding/json"
	"github.com/aidarkolbaev/pubsub"
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
	pubSub     *pubsub.PubSub
	msgBroker  msgbroker.MessageBroker
	workerPool *workerpool.WorkerPool
}

func New(c *config.Config, s storage.Storage, mb msgbroker.MessageBroker) *API {
	api := &API{
		echo:       echo.New(),
		config:     c,
		storage:    s,
		pubSub:     pubsub.New(websocket.Pusher{}),
		msgBroker:  mb,
		workerPool: workerpool.New(c.MaxWorkers),
	}

	err := api.msgBroker.Subscribe("messages", api.handleMessages)

	api.echo.HideBanner = true
	api.echo.Use(middleware.CORS())

	api.echo.GET("/", api.ping)
	api.echo.POST("/room", api.createRoom)
	api.echo.GET("/room/:roomID", api.getRoom)
	api.echo.Any("/ws", api.websocket)

	return api
}

func (api *API) Start() error {
	return api.echo.Start(":" + strconv.Itoa(api.config.HttpPort))
}

func (api *API) Close() error {
	api.workerPool.StopWait()
	_ = api.msgBroker.Close()
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

// Endpoint to establish websocket connection
func (api *API) websocket(c echo.Context) error {
	username := c.QueryParam("username")
	roomID := c.QueryParam("room_id")
	if !api.storage.RoomExist(roomID) {
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
		ID:     roomID + username + utils.RandString(3),
		Name:   username,
		RoomID: roomID,
		Conn:   conn,
	}
	api.serveUser(user)
	return nil
}

// Serves user websocket connection
func (api *API) serveUser(u *model.User) {
	api.pubSub.Subscribe(u, u.RoomID)

	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	done := make(chan bool)

	go func() {
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
	}()

	for {
		b, err := wsutil.ReadClientText(u.Conn)
		if err != nil {
			done <- true
			break
		}
		var req websocket.Request
		err = json.Unmarshal(b, &req)
		if err != nil {

		}

		// TODO implement
		err = api.msgBroker.Publish(b, "messages")
	}

	api.pubSub.Unsubscribe(u, u.RoomID)
	_ = u.Conn.Close()
	log.Infof("user %s disconnected from room %s", u.Name, u.RoomID)
}

// Message handler
func (api *API) handleMessages(msg *msgbroker.Message) {
	api.workerPool.Submit(func() {
		var req websocket.Request
		err := json.Unmarshal(msg.Data, &req)
		if err != nil {
			log.Error(err)
			return
		}
		api.pubSub.Publish(msg.Data, req.User.RoomID)
	})
}
