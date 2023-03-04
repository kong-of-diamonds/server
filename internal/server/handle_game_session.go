package server

import (
	"encoding/json"
	"log"
	"net/http"
)

type playerData struct {
	ID   string `msgpack:"id" json:"id"`
	name string
}

type pickMessage struct {
	Number float64 `msgpack:"number" json:"number"`
}

func (s *server) gameSession(w http.ResponseWriter, r *http.Request) {
	ws, errUpgrade := upgradeConn(w, r)

	if errUpgrade != nil {
		log.Println(errUpgrade)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer ws.Close()

	_, data, errReadPlayerData := ws.ReadMessage()
	if errReadPlayerData != nil {
		log.Printf("gameSession ReadMessage: %v", errReadPlayerData)
		return
	}

	var p playerData
	if err := json.Unmarshal(data, &p); err != nil {
		log.Printf("gameSession Unmarshal playerData: %v", err)
		return
	}

	game := s.gm.GetGame()
	player, errAddPlayer := game.AddPlayer(p.ID, ws)
	if errAddPlayer != nil {
		log.Printf("gameSession AddPlayer: %v", errAddPlayer)
		return
	}

	for {
		_, eventData, errReadMessage := ws.ReadMessage()
		if errReadMessage != nil {
			log.Printf("gameSession ReadMessage: %v", errReadMessage)
			break
		}

		var pickedData pickMessage
		if err := json.Unmarshal(eventData, &pickedData); err != nil {
			log.Printf("gameSession Unmarshal pickMessage: %v", err)
			break
		}

		turn := game.CurrentTurn()
		if turn == nil {
			//@todo message to user
			break
		}

		if err := turn.AddPick(player, pickedData.Number); err != nil {
			log.Printf("gameSession AddPick: %v", err)
			//@todo message to user
			break
		}
	}
	game.RmPlayer(player)
}
