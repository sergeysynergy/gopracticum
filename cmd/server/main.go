package main

import (
	"time"

	"github.com/sergeysynergy/gopracticum/internal/httpserver"
)

func main() {
	cfg := httpserver.Config{
		Port:            "8080",
		ShutdownTimeout: 2 * time.Millisecond,
	}
	s := httpserver.New(cfg)

	s.Serve()
}
