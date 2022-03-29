package main

import (
	"log"
	"net/http"
	"time"

	"github.com/sergeysynergy/gopracticum/server"
)

// HelloWorld — обработчик запроса.
func HelloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("<h1>Hello, World</h1>"))
}

func main() {
	// маршрутизация запросов обработчику
	//http.HandleFunc("/", HelloWorld)
	// запуск сервера с адресом localhost, порт 8080
	//http.ListenAndServe(":8080", nil)
	handler2 := http.HandlerFunc(HelloWorld)
	//handlers
	// storage
	// server
	s := server.New()

	hAddr := "127.0.0.1:8080"
	httpServer := &http.Server{
		Addr:           hAddr,
		Handler:        handler2,         // if nil use default http.DefaultServeMux
		ReadTimeout:    time.Second * 10, // max duration reading entire request
		WriteTimeout:   time.Second * 10, // max timing write response
		IdleTimeout:    time.Second * 10, // max time wait for the next request
		MaxHeaderBytes: 1 << 20,          // 2^20 = 128 Kb
	}
	func() {
		log.Printf("starting http server at: %s\n", httpServer.Addr)
		log.Fatal(httpServer.ListenAndServe())
	}()
}
