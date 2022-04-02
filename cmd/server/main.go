package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"

	"github.com/sergeysynergy/gopracticum/internal/httpserver"
)

func init() {
	err := godotenv.Load("./config/.env")
	if err != nil {
		err = godotenv.Load("../../config/.env")
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}
}

func main() {
	port := os.Getenv("SERVER_PORT")

	cfg := httpserver.Config{
		Port:         port,
		GraceTimeout: 2 * time.Second,
	}
	s := httpserver.New(cfg)

	s.Serve()
}
