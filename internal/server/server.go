package server

import "net/http"

type Config struct {
	Addr string
}

type server struct {
	http.Server
	Cfg Config
}

func New() {

}
