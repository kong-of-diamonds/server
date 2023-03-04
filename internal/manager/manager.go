package manager

import (
	"kod/server/internal/game"
	"log"
	"sync"
)

const numberOfSessions int = 1

type GamesManager struct {
	games []*game.Game
	mu    sync.Mutex
}

func NewGamesManager() *GamesManager {
	gm := &GamesManager{
		games: make([]*game.Game, numberOfSessions),
	}
	gm.updateSessins()
	return gm
}

func (gm *GamesManager) GetGame() *game.Game {
	return gm.games[len(gm.games)-1]
}

func (gm *GamesManager) updateSessins() {
	for i := 0; i < numberOfSessions; i++ {
		go func(idx int) {
			gm.updateSession(idx)
		}(i)
	}
}

func (gm *GamesManager) updateSession(idx int) {
	for {
		log.Printf("Game %d is restarted", idx)
		var game *game.Game = game.NewGame()
		gm.mu.Lock()
		gm.games[idx] = game
		gm.mu.Unlock()
		<-game.Done()
		println("chan freed")
		gm.games[idx] = nil
	}
}
