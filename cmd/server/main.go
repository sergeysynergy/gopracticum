package main

import (
	"github.com/sergeysynergy/gopracticum/internal/httpserver"
	"time"
)

func main() {
	cfg := httpserver.Config{
		Port:         "8080",
		GraceTimeout: 2 * time.Second,
	}
	s := httpserver.New(cfg)

	s.Serve()
}
