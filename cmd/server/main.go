package main

import (
	"time"

	"github.com/sergeysynergy/gopracticum/internal/httpserver"
)

func main() {
	cfg := httpserver.Config{
		Address:         "127.0.0.1",
		Port:            "8080",
		ShutdownTimeout: 2 * time.Millisecond,
	}
	s := httpserver.New(cfg)

	s.Serve()
}
