package main

import (
	"github.com/labstack/gommon/log"
	"smotri.me/api"
	"smotri.me/config"
	"smotri.me/storage"
)

func main() {
	c := config.Get()
	s, err := storage.New(c)
	if err != nil {
		log.Fatal(err)
	}
	defer s.Close()

	a := api.New(c, s)
	defer a.Close()

	log.Fatal(a.Start())
}
