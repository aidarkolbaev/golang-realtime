package main

import (
	"context"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/gommon/log"
	"os"
	"os/signal"
	"smotri.me/api"
	"smotri.me/config"
	"smotri.me/pkg/msgbroker"
	"smotri.me/storage"
	"time"
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

	// Storage
	s := storage.New(rdb)
	// Message broker
	mb := msgbroker.NewRedisBroker(rdb)

	// API
	a := api.New(c, s, mb)

	go func() {
		// Starting API
		if err := a.Start(); err != nil {
			log.Warn(err)
		}
	}()

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, os.Kill)
	// waiting for signals
	quit := <-signals
	log.Infof("signal %s received, stopping server...", quit)
	// Stopping server
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	if err = a.Close(ctx); err != nil {
		log.Error(err)
	}
	cancel()

	if err = mb.Close(); err != nil {
		log.Error(err)
	}
	if err = rdb.Close(); err != nil {
		log.Error(err)
	}
}
