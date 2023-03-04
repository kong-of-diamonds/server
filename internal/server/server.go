package server

import (
	"kod/server/internal/manager"
	"net/http"
)

type server struct {
	http.Server
	gm *manager.GamesManager
}

func NewServer() *server {
	var server server = server{}
	server.Server = http.Server{
		Addr: ":8080",
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/session", server.gameSession)
	server.Handler = mux

	server.gm = manager.NewGamesManager()

	return &server
}

func (s *server) Run() error {
	return s.ListenAndServe()
}
