package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sergeysynergy/gopracticum/agent"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	cfg := agent.Config{
		PollInterval:   1 * time.Second, // in prod 2
		ReportInterval: 2 * time.Second, // in prod 10
		Address:        "127.0.0.1",
		Port:           ":8080",
	}
	agent, err := agent.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Запускаем агент, получаем функцию отметы выполнения контекста.
	cancel := agent.Run()
	// Функцию cancel нужно обязательно выполнить в коде, иначе сборщик мусора не удалит созданный дочерний контекст
	// и произойдёт утечка памяти.
	defer cancel()

	// Агент должен штатно завершаться по сигналам: syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	s := <-c
	log.Println("Получен сигнал завершения работы:", s)
	log.Println("Работа агента штатно завершена")
}
