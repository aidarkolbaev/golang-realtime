package config

import (
	"github.com/kelseyhightower/envconfig"
	"github.com/labstack/gommon/log"
	"sync"
)

type Config struct {
	HttpPort      int    `envconfig:"HTTP_PORT" required:"true"`
	RedisAddr     string `envconfig:"REDIS_ADDR" required:"true"`
	RedisPassword string `envconfig:"REDIS_PASSWORD" required:"true"`
	RedisDB       int    `envconfig:"REDIS_DB" required:"false" default:"0"`
}

var (
	c    Config
	once sync.Once
)

func Get() *Config {
	once.Do(func() {
		err := envconfig.Process("", &c)
		if err != nil {
			log.Fatal(err)
		}
	})
	return &c
}
