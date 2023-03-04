package game

import (
	"math/rand"
	"sync"
	"time"
)

const (
	factorMin = 0.1
	factorMax = 0.9

	playersCount    = 5
	gameDuretionSec = 30
)

type Game struct {
	id         string
	created    int64
	players    []*player
	turns      []*turn
	events     []*event
	Factor     float64 `msgpack:"factor" json:"factor"`
	MaxPlayers int     `msgpack:"max_players" json:"max_players"`

	timeout time.Duration

	done chan struct{}

	muPlayers sync.Mutex
	muTurns   sync.Mutex
	muEvents  sync.Mutex
}

func NewGame() *Game {
	rand.Seed(time.Now().Unix())

	return &Game{
		created:    time.Now().UnixMicro(),
		players:    make([]*player, 0, playersCount),
		turns:      make([]*turn, 0),
		events:     make([]*event, 0),
		Factor:     factorMin + rand.Float64()*(factorMax-factorMin),
		done:       make(chan struct{}),
		timeout:    gameDuretionSec * time.Second,
		MaxPlayers: playersCount,
	}
}

func (g *Game) isStarted() bool {
	return len(g.turns) > 0
}

func (g *Game) endGame() {
	println("finishing game")
	g.addEvent(newEventEndGame(g.players))
	g.shutdown()
}

func (g *Game) shutdown() {
	g.muEvents.Lock()
	close(g.done)
	g.muEvents.Unlock()
}

func (g *Game) Done() chan struct{} {
	return g.done
}
