package main

import (
	"github.com/sergeysynergy/gopracticum/internal/httpserver"
)

func main() {
	cfg := httpserver.Config{
		Port: "8080",
	}
	s := httpserver.New(cfg)

	s.Serve()
}
