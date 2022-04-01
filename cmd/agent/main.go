package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/sergeysynergy/gopracticum/agent"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	cfg := agent.Config{
		PollInterval:   2 * time.Second,  // in prod 2
		ReportInterval: 10 * time.Second, // in prod 10
		Address:        "127.0.0.1",
		Port:           ":8080",
	}
	agent, err := agent.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	agent.Run()
}
