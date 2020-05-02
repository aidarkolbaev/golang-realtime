package main

import (
	"github.com/go-redis/redis/v7"
	"github.com/labstack/gommon/log"
	"smotri.me/api"
	"smotri.me/config"
	"smotri.me/pkg/msgbroker"
	"smotri.me/storage"
)

func main() {
	// APP configuration
	c := config.Get()

	// Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.RedisAddr,
		Password: c.RedisPassword,
		DB:       c.RedisDB,
	})
	err := rdb.Ping().Err()
	if err != nil {
		log.Fatal(err)
	}
	defer rdb.Close()

	// Storage
	s := storage.New(rdb)
	// Message broker
	mb := msgbroker.NewRedisBroker(rdb)
	defer mb.Close()

	// API
	a := api.New(c, s, mb)
	defer a.Close()
	// Starting API
	log.Fatal(a.Start())
}
