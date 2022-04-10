package main

import (
	"github.com/caarlos0/env/v6"
	"log"

	"github.com/sergeysynergy/gopracticum/internal/handlers"
	"github.com/sergeysynergy/gopracticum/internal/httpserver"
)

type Config struct {
	Addr string `env:"ADDRESS"`
}

func main() {
	var cfg Config

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalln(err)
	}

	h := handlers.New()
	s := httpserver.New(h.GetRouter(), httpserver.WithAddress(cfg.Addr))

	s.Serve()
}
