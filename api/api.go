package api

import (
	"github.com/gobwas/ws"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"net/http"
	"smotri.me/config"
	"smotri.me/model"
	"smotri.me/pkg/utils"
	"smotri.me/storage"
	"strconv"
	"time"
)

type API struct {
	e *echo.Echo
	c *config.Config
	s storage.Storage
}

func New(c *config.Config, s storage.Storage) *API {
	api := &API{
		e: echo.New(),
		c: c,
		s: s,
	}

	api.e.HideBanner = true
	api.e.Use(middleware.CORS())

	api.e.GET("/", api.ping)
	api.e.POST("/room", api.createRoom)
	api.e.GET("/room/:roomID", api.getRoom)
	api.e.Any("/ws", api.websocket)

	return api
}

func (api *API) Start() error {
	return api.e.Start(":" + strconv.Itoa(api.c.HttpPort))
}

func (api *API) Close() error {
	return api.e.Close()
}

func (api *API) ping(c echo.Context) error {
	_, err := api.s.IncrVisits()
	if err != nil {
		log.Error(err)
	}
	return c.String(http.StatusOK, "OK")
}

func (api *API) createRoom(c echo.Context) error {
	var room model.Room
	err := c.Bind(&room)
	if err != nil || !room.Valid() {
		if err != nil {
			log.Warn(err)
		}
		return echo.NewHTTPError(http.StatusUnprocessableEntity)
	}

	room.ID, err = api.s.CreateTempRoom(&room, time.Minute*15)
	if err != nil {
		log.Error(err)
		return echo.NewHTTPError(http.StatusConflict)
	}

	return c.JSON(http.StatusOK, &room)
}

func (api *API) getRoom(c echo.Context) error {
	roomID := c.Param("roomID")
	room, err := api.s.GetTempRoom(roomID)
	if err != nil {
		log.Info(err)
		return echo.NewHTTPError(http.StatusNotFound)
	}
	return c.JSON(http.StatusOK, room)
}

func (api *API) websocket(c echo.Context) error {
	username := c.QueryParam("username")
	roomID := c.QueryParam("room_id")
	if !api.s.RoomExist(roomID) {
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
		Name:   username,
		RoomID: roomID,
		Conn:   conn,
	}
	go user.ServeWebsocket()
	return c.NoContent(http.StatusOK)
}
