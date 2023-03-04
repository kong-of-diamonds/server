package server

import (
	"net/http"

	"github.com/gorilla/websocket"
)

func upgradeConn(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return newUpgrader().Upgrade(w, r, nil)

}

func newUpgrader() (upgrader *websocket.Upgrader) {
	upgrader = &websocket.Upgrader{}
	upgrader.CheckOrigin = upgraderCheckOrigin
	upgrader.WriteBufferSize = 256
	return
}

func upgraderCheckOrigin(r *http.Request) bool {
	return true
}
